package parse

import (
	"errors"
	"io"
	"unicode/utf8"
)

type savepoint struct {
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
	reusableSpace  []byte      // to avoid allocation in parsing
	subSources     []*Source   // to reuse sub sources
	Attachment     interface{} // to pass context
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
	src.savepointStack = nil
}

// SetBuffer will prevent the buffer reuse,
// so that buffer returned by Peek() can be saved for later use.
func (src *Source) SetBuffer(buf []byte) {
	src.buf = buf
}

// SetNewBuff will prevent the buffer reuse
func (src *Source) SetNewBuffer() {
	src.buf = make([]byte, 64)
}

// StoreSavepoint mark current position, and start recording.
// Later we can rollback to current position.
func (src *Source) StoreSavepoint() {
	src.savepointStack = append(src.savepointStack, &savepoint{
		current: src.current,
	})
}

var errNoSavepoint = errors.New("no savepoint in stack")

func (src *Source) DeleteSavepoint() {
	if len(src.savepointStack) == 0 {
		src.ReportError(errNoSavepoint)
		return
	}
	src.savepointStack = src.savepointStack[:len(src.savepointStack)-1]
}

// RollbackToSavepoint rollback the cursor to previous savepoint.
// The bytes read from reader will be replayed.
func (src *Source) RollbackToSavepoint() {
	if len(src.savepointStack) == 0 {
		src.ReportError(errNoSavepoint)
		return
	}
	i := len(src.savepointStack) - 1
	savepoint := src.savepointStack[i]
	src.current = savepoint.current
	src.nextList = savepoint.nextList
	src.savepointStack = src.savepointStack[:i]
	src.err = nil
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
func (src *Source) PeekN(n int) []byte {
	if n <= len(src.current) {
		return src.current[:n]
	}
	if src.reader == nil {
		return src.current
	}
	src.StoreSavepoint()
	peeked := src.ReadN(n)
	src.RollbackToSavepoint()
	return peeked
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
func (src *Source) CopyN(n int) []byte {
	space := src.ClaimSpace()
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
func (src *Source) Consume1() {
	src.current = src.current[1:]
	if len(src.current) == 0 {
		src.Consume()
	}
}

// Read1 like ConsumeN, with N == 1
func (src *Source) Read1() byte {
	b := src.current[0]
	src.current = src.current[1:]
	if len(src.current) == 0 {
		src.Consume()
	}
	return b
}

func (src *Source) ReadN(n int) []byte {
	if n <= len(src.current) {
		return src.current[:n]
	}
	if src.reader == nil {
		src.ReportError(io.ErrUnexpectedEOF)
		return src.current
	}
	peeked := append(src.ClaimSpace(), src.current...)
	src.Consume()
	for {
		peeked = append(peeked, src.current...)
		if len(peeked) > n {
			src.current = peeked[n:]
			return peeked[:n]
		}
		if len(peeked) == n {
			src.Consume()
			return peeked
		}
		if src.Error() != nil {
			return peeked
		}
		src.Consume()
	}
}

var errExpectedBytesNotFound = errors.New(`expected bytes not found`)

// Expect1 like ConsumeN, with N == 1.
// bytes will not be consumed if not match
func (src *Source) Expect1(b1 byte) {
	if b1 != src.current[0] {
		src.ReportError(errExpectedBytesNotFound)
		return
	}
	src.Consume1()
}

// Expect2 like ConsumeN, with N == 2.
// bytes will not be consumed if not match
func (src *Source) Expect2(b1, b2 byte) {
	if len(src.current) >= 2 {
		c1 := b1 == src.current[0]
		c2 := b2 == src.current[1]
		if c1 && c2 {
			src.ConsumeN(2)
		} else {
			src.ReportError(errExpectedBytesNotFound)
		}
		return
	}
	bytes := src.PeekN(2)
	if len(bytes) != 2 {
		src.ReportError(errExpectedBytesNotFound)
		return
	}
	c1 := b1 == src.current[0]
	c2 := b2 == src.current[1]
	if c1 && c2 {
		src.ConsumeN(2)
		return
	}
	src.ReportError(errExpectedBytesNotFound)
}

// Expect3 like ConsumeN, with N == 3.
// bytes will not be consumed if not match
func (src *Source) Expect3(b1, b2, b3 byte) {
	if len(src.current) >= 3 {
		c1 := b1 == src.current[0]
		c2 := b2 == src.current[1]
		c3 := b3 == src.current[2]
		if c1 && c2 && c3 {
			src.ConsumeN(3)
		} else {
			src.ReportError(errExpectedBytesNotFound)
		}
		return
	}
	bytes := src.PeekN(3)
	if len(bytes) != 3 {
		src.ReportError(errExpectedBytesNotFound)
		return
	}
	c1 := b1 == src.current[0]
	c2 := b2 == src.current[1]
	c3 := b3 == src.current[2]
	if c1 && c2 && c3 {
		src.ConsumeN(3)
		return
	}
	src.ReportError(errExpectedBytesNotFound)
}

// Expect4 like ConsumeN, with N == 4.
// bytes will not be consumed if not match
func (src *Source) Expect4(b1, b2, b3, b4 byte) {
	if len(src.current) >= 4 {
		c1 := b1 == src.current[0]
		c2 := b2 == src.current[1]
		c3 := b3 == src.current[2]
		c4 := b4 == src.current[3]
		if c1 && c2 && c3 && c4 {
			src.ConsumeN(4)
		} else {
			src.ReportError(errExpectedBytesNotFound)
		}
		return
	}
	bytes := src.PeekN(4)
	if len(bytes) != 4 {
		src.ReportError(errExpectedBytesNotFound)
		return
	}
	c1 := b1 == src.current[0]
	c2 := b2 == src.current[1]
	c3 := b3 == src.current[2]
	c4 := b4 == src.current[3]
	if c1 && c2 && c3 && c4 {
		src.ConsumeN(4)
		return
	}
	src.ReportError(errExpectedBytesNotFound)
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
	fullBuf := src.PeekN(int(sz))
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

func (src *Source) BorrowSubSource(buf []byte) *Source {
	if len(src.subSources) == 0 {
		return &Source{
			current: buf,
		}
	}
	subSrc := src.subSources[0]
	src.subSources = src.subSources[1:]
	subSrc.current = buf
	return subSrc
}

func (src *Source) ReturnSubSource(subSrc *Source) {
	subSrc.Reset(nil, nil)
	subSrc.subSources = append(subSrc.subSources, subSrc)
}

func (src *Source) ReserveSpace(space []byte) {
	src.reusableSpace = space
}

func (src *Source) ClaimSpace() []byte {
	space := src.reusableSpace
	src.reusableSpace = nil
	return space
}
