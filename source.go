package parse

import (
	"errors"
	"io"
	"unicode/utf8"
)

type savepoint struct {
	name     string
	current  []byte
	nextList [][]byte
}

// Source is generalization of io.Reader and []byte.
// It supports read ahead.
// It supports read byte by byte.
// It supports read unicode code point by code point (as rune or []byte).
// It supports savepoint and rollback.
type Source struct {
	err            error
	reader         io.Reader
	current        []byte
	nextList       [][]byte
	buf            []byte
	savepointStack []*savepoint
	Attachment     interface{}
}

// NewSource construct a source from io.Reader.
// At least one byte should be read from the io.Reader, otherwise error will be returned.
func NewSource(reader io.Reader, buf []byte) (*Source, error) {
	n, err := reader.Read(buf)
	if n == 0 {
		return nil, err
	}
	return &Source{
		reader:  reader,
		current: buf[:n],
		buf:     buf,
	}, nil
}

// NewSourceString construct a source from string. Len should >= 1.
func NewSourceString(str string) *Source {
	src := &Source{
		current: []byte(str),
	}
	if len(str) == 0 {
		src.ReportError(errors.New("source string is empty"))
	}
	return src
}

func (src *Source) Reset(reader io.Reader, buf []byte) {
	src.reader = reader
	src.current = buf
	src.buf = buf
	src.err = nil
}

// SetBuffer will prevent the buffer reuse,
// so that buffer returned by Peek() can be saved for later use.
func (src *Source) SetBuffer(buf []byte) {
	src.buf = buf
}

// Savepoint mark current position, and start recording.
// Later we can rollback to current position.
func (src *Source) Savepoint(savepointName string) {
	if src.Error() != nil {
		return
	}
	for _, savepoint := range src.savepointStack {
		if savepoint.name == savepointName {
			src.ReportError(errors.New("savepoint name has already been used: " + savepointName))
			return
		}
	}
	src.savepointStack = append(src.savepointStack, &savepoint{
		name:    savepointName,
		current: src.current,
	})
}

// RollbackTo rollback the cursor to previous savepoint.
// The bytes read from reader will be replayed.
func (src *Source) RollbackTo(savepointName string) {
	for i, savepoint := range src.savepointStack {
		if savepoint.name == savepointName {
			src.current = src.savepointStack[i].current
			src.nextList = src.savepointStack[i].nextList
			src.savepointStack = src.savepointStack[:i]
			src.err = nil
			return
		}
	}
	src.ReportError(errors.New("savepoint not found: " + savepointName))
}

// Peek return the current buffer ready to be parsed.
// The buffer will have at least one byte.
func (src *Source) Peek() []byte {
	return src.current
}

// Peek1 return the first byte in the buffer to parse.
func (src *Source) Peek1() byte {
	return src.current[0]
}

// PeekN return the first N bytes in the buffer to parse.
// If N is longer than current buffer, it will read from reader.
// The cursor will not be moved.
func (src *Source) PeekN(n int) ([]byte, error) {
	if n <= len(src.current) {
		return src.current[:n], nil
	}
	if src.reader == nil {
		return src.current, io.EOF
	}
	src.Savepoint("Peek")
	peeked := src.current
	src.Consume()
	for {
		peeked = append(peeked, src.current...)
		if len(peeked) >= n {
			src.RollbackTo("Peek")
			return peeked[:n], nil
		}
		if src.Error() != nil {
			err := src.Error()
			src.RollbackTo("Peek")
			return peeked, err
		}
		src.Consume()
	}
}

// ConsumeN move the cursor N bytes. If N is larger than current buffer,
// more bytes will read from reader and be discarded.
func (src *Source) ConsumeN(n int) {
	for {
		if n <= len(src.current) {
			src.current = src.current[n:]
			if len(src.current) == 0 {
				src.Consume()
			}
			return
		}
		if src.err != nil {
			if src.err == io.EOF {
				src.err = io.ErrUnexpectedEOF
			}
			return
		}
		n -= len(src.current)
		src.Consume()
	}
}

// CopyN works like ConsumeN.
// But unlike ConsumeN, it copies the bytes read into new buffer and return.
func (src *Source) CopyN(space []byte, n int) []byte {
	for {
		if n <= len(src.current) {
			space = append(space, src.current[:n]...)
			src.current = src.current[n:]
			if len(src.current) == 0 {
				src.Consume()
			}
			return space
		}
		if src.err != nil {
			if src.err == io.EOF {
				src.err = io.ErrUnexpectedEOF
			}
			return space
		}
		n -= len(src.current)
		space = append(space, src.current...)
		src.Consume()
	}
}

// Consume1 like ConsumeN, with N == 1
func (src *Source) Consume1(b1 byte) {
	if b1 != src.current[0] {
		src.ReportError(errors.New(
			"expect " + string([]byte{b1}) +
				" but found " + string([]byte{src.current[0]})))
	}
	src.ConsumeN(1)
}

// Consume will discard whole current buffer, and read next buffer.
func (src *Source) Consume() {
	if src.reader == nil {
		src.current = nil
		src.ReportError(io.EOF)
		return
	}
	if len(src.nextList) != 0 {
		src.current = src.nextList[0]
		src.nextList = src.nextList[1:]
		return
	}
	if len(src.savepointStack) > 0 {
		src.buf = make([]byte, len(src.buf))
	}
	n, err := src.reader.Read(src.buf)
	if err != nil {
		src.ReportError(err)
	}
	src.current = src.buf[:n]
	for _, savepoint := range src.savepointStack {
		savepoint.nextList = append(savepoint.nextList, src.current)
	}
}

// PeekRune read unicode code point as rune, without moving cursor.
func (src *Source) PeekRune() (rune, int) {
	p0 := src.current[0]
	x := first[p0]
	if x >= as {
		return utf8.DecodeRune(src.current)
	}
	sz := x & 7
	fullBuf, _ := src.PeekN(int(sz))
	return utf8.DecodeRune(fullBuf)
}

// PeekUtf8 read one full code point without decoding into rune
func (src *Source) PeekUtf8() []byte {
	p0 := src.current[0]
	x := first[p0]
	if x >= as {
		return src.current[:1]
	}
	sz := int(x & 7)
	fullBuf, _ := src.PeekN(sz)
	return fullBuf
}

// ReportError set the source in error condition.
func (src *Source) ReportError(err error) {
	if src.err == nil {
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
