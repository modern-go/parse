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

type Source struct {
	err            error
	reader         io.Reader
	current        []byte
	nextList       [][]byte
	buf            []byte
	savepointStack []*savepoint
	Attachment     interface{}
}

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

func NewSourceString(str string) *Source {
	src := &Source{
		current: []byte(str),
	}
	if len(str) == 0 {
		src.ReportError(errors.New("source string is empty"))
	}
	return src
}

func (src *Source) SetBuffer(buf []byte) {
	src.buf = buf
}

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

func (src *Source) Peek() []byte {
	return src.current
}

func (src *Source) Peek1() byte {
	return src.current[0]
}

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

func (src *Source) Consume1(b1 byte) {
	if b1 != src.current[0] {
		src.ReportError(errors.New(
			"expect " + string([]byte{b1}) +
				" but found " + string([]byte{src.current[0]})))
	}
	src.ConsumeN(1)
}

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

func (src *Source) ReportError(err error) {
	if src.err == nil {
		src.err = err
	}
}

func (src *Source) Error() error {
	return src.err
}

func (src *Source) FatalError() error {
	if src.err == io.EOF {
		return nil
	}
	return src.err
}
