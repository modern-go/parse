package parse

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"unicode/utf8"
)

type stack struct {
	buf []breakInfo
}

type breakInfo struct {
	nextIdx int
}

func (s *stack) Push(info breakInfo) {
	s.buf = append(s.buf, info)
}

func (s *stack) Pop() breakInfo {
	last := len(s.buf) - 1
	brkInfo := s.buf[last]
	s.buf = s.buf[:last]
	return brkInfo
}

func (s *stack) Empty() bool {
	return len(s.buf) == 0
}

// Source is generalization of io.Reader and []byte.
// It supports read ahead.
// It supports read byte by byte.
// It supports read unicode code point by code point (as rune or []byte).
// It supports savepoint and rollback.
type Source struct {
	err            error
	reader         io.Reader
	readBytes      []byte
	buf            []byte
	nextIdx        int
	savepointStack *stack
}

const (
	_maxBufLen = 40
)

// NewSource construct a source from io.Reader.
// At least one byte should be read from the io.Reader, otherwise error will be returned.
func NewSource(reader io.Reader, bufLen int) (*Source, error) {
	if bufLen > _maxBufLen {
		bufLen = _maxBufLen
	}
	buf := make([]byte, bufLen)
	n, err := reader.Read(buf)
	if n == 0 {
		return nil, err
	}
	readByes := make([]byte, n)

	copy(readByes, buf)
	return &Source{
		reader:         reader,
		readBytes:      readByes,
		buf:            buf,
		savepointStack: new(stack),
	}, nil
}

// NewSourceString construct a source from string. Len should >= 1.
func NewSourceString(str string) (*Source, error) {
	if len(str) == 0 {
		return nil, errors.New("source string is empty")
	}
	reader := bytes.NewReader([]byte(str))
	return NewSource(reader, _maxBufLen)
}

// StoreSavepoint mark current position, and start recording.
// Later we can rollback to current position.
// Make sure there's no error, rollback will clear the error
func (src *Source) StoreSavepoint() {
	src.savepointStack.Push(breakInfo{nextIdx: src.nextIdx})
}

var errNoSavepoint = errors.New("no savepoint in stack")

// DeleteSavepoint delete the lastest savepoint
func (src *Source) DeleteSavepoint() {
	if src.savepointStack.Empty() {
		src.ReportError(errNoSavepoint)
		return
	}
	src.savepointStack.Pop()
}

// RollbackToSavepoint rollback the cursor to previous savepoint.
// The bytes read from reader will be replayed.
func (src *Source) RollbackToSavepoint() {
	if src.savepointStack.Empty() {
		src.ReportError(errNoSavepoint)
		return
	}
	brkInfo := src.savepointStack.Pop()
	src.nextIdx = brkInfo.nextIdx
	src.err = nil
}

// Peek1 return the first byte in the buffer to parse.
func (src *Source) Peek1() byte {
	if src.nextIdx < len(src.readBytes) {
		return src.readBytes[src.nextIdx]
	}
	src.consume()
	if src.nextIdx < len(src.readBytes) {
		return src.readBytes[src.nextIdx]
	}

	// EOF
	src.ReportError(io.ErrUnexpectedEOF)
	return 0x00 //NULL
}

// Peek peeks as many bytes as possible without triggering consume
func (src *Source) Peek() []byte {
	return src.readBytes[src.nextIdx:]
}

// PeekAll peek all of the rest bytes
func (src *Source) PeekAll() []byte {
	data, _ := ioutil.ReadAll(src.reader)
	if len(data) == 0 {
		src.ReportError(io.EOF)
		if src.nextIdx < len(src.readBytes) {
			return src.readBytes[src.nextIdx:]
		}
		return nil
	}
	src.readBytes = append(src.readBytes, data...)
	return src.readBytes[src.nextIdx:]
}

// PeekN return the first N bytes in the buffer to parse.
// If N is longer than current buffer, it will read from reader.
// The cursor will not be moved.
func (src *Source) PeekN(n int) []byte {
	rest := len(src.readBytes) - src.nextIdx
	for src.Error() == nil && rest < n {
		src.consume()
		rest = len(src.readBytes) - src.nextIdx
	}
	if rest >= n {
		return src.readBytes[src.nextIdx : src.nextIdx+n]
	}
	//EOF
	src.ReportError(io.ErrUnexpectedEOF)
	return src.readBytes[src.nextIdx : src.nextIdx+rest]

}

// Read1 like ConsumeN, with N == 1
func (src *Source) Read1() byte {
	b := src.Peek1()
	src.nextIdx++
	return b
}

// ReadN read N bytes and move cursor forward
func (src *Source) ReadN(n int) []byte {
	buf := src.PeekN(n)
	src.nextIdx += len(buf)
	return buf
}

// ReadAll read all bytes until EOF
func (src *Source) ReadAll() []byte {
	buf := src.PeekAll()
	src.nextIdx += len(buf)
	return buf
}

var errExpectedBytesNotFound = errors.New(`expected bytes not found`)

// Expect1 like Read1
// bytes will not be consumed if not match
func (src *Source) Expect1(b1 byte) bool {
	if src.Peek1() == b1 {
		src.nextIdx++
		return true
	}
	return false
}

// Expect2 like ReadN, with N == 2.
// bytes will not be consumed if not match
func (src *Source) Expect2(b1, b2 byte) bool {
	buf := src.PeekN(2)
	if len(buf) < 2 {
		return false
	}
	if buf[0] == b1 && buf[1] == b2 {
		src.nextIdx += 2
		return true
	}
	return false
}

// Expect3 like ReadN, with N == 3.
// bytes will not be consumed if not match
func (src *Source) Expect3(b1, b2, b3 byte) bool {
	buf := src.PeekN(3)
	if len(buf) < 3 {
		return false
	}
	if buf[0] == b1 && buf[1] == b2 && buf[2] == b3 {
		src.nextIdx += 3
		return true
	}
	return false
}

// Expect4 like ConsumeN, with N == 4.
// bytes will not be consumed if not match
func (src *Source) Expect4(b1, b2, b3, b4 byte) bool {
	buf := src.PeekN(4)
	if len(buf) < 4 {
		return false
	}
	if buf[0] == b1 && buf[1] == b2 && buf[2] == b3 && buf[3] == b4 {
		src.nextIdx += 4
		return true
	}
	return false
}

// consume will discard whole current buffer, and read next buffer.
func (src *Source) consume() {
	if src.reader == nil {
		src.ReportError(io.EOF)
		return
	}
	n, err := src.reader.Read(src.buf)
	if err != nil || n == 0 {
		src.ReportError(err)
		return
	}
	src.readBytes = append(src.readBytes, src.buf[:n]...)
}

// PeekRune read unicode code point as rune, without moving cursor.
func (src *Source) PeekRune() (rune, int) {
	p0 := src.Peek1()
	x := first[p0]
	if x >= as {
		return utf8.DecodeRune(src.readBytes[src.nextIdx:])
	}
	sz := x & 7
	fullBuf := src.PeekN(int(sz))
	return utf8.DecodeRune(fullBuf)
}

// PeekUtf8 read one full code point without decoding into rune
func (src *Source) PeekUtf8() []byte {
	p0 := src.Peek1()
	x := first[p0]
	if x >= as {
		return []byte{p0}
	}
	sz := int(x & 7)
	fullBuf := src.PeekN(sz)
	return fullBuf
}

// ReportError set the source in error condition.
func (src *Source) ReportError(err error) {
	if src.err == nil || src.err == io.EOF {
		src.err = err
	}
}

// Error tells if the source is in error condition.
// EOF is considered as error.
func (src *Source) Error() error {
	return src.err
}

// FatalError tells if the source is in fatal error condition
func (src *Source) FatalError() error {
	if src.err == io.EOF {
		return nil
	}
	return src.err
}

// ResetError clear the error
func (src *Source) ResetError() {
	src.err = nil
}
