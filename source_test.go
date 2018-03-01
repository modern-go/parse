package parse_test

import (
	"bytes"
	"context"
	"github.com/modern-go/parse"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
	"io"
	"strings"
	"testing"
)

func TestSource_Savepoint(t *testing.T) {
	t.Run("savepoint name must be unique", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("hello")
		src.Savepoint("hello")
		src.Savepoint("hello")
		must.NotNil(src.Error())
	}))
	t.Run("rollback should delete the savepoint", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("hello")
		src.Savepoint("hello")
		src.RollbackTo("hello")
		src.Savepoint("hello")
		must.Nil(src.Error())
	}))
	t.Run("rollback not existing savepoint", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("hello")
		src.RollbackTo("hello")
		must.NotNil(src.Error())
	}))
}

func TestSource_RollbackTo(t *testing.T) {
	t.Run("immediate rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 1))[0].(*parse.Source)
		src.Savepoint("s1")
		src.RollbackTo("s1")
		must.Equal([]byte{'a'}, src.Peek())
	}))
	t.Run("partial consume then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 2))[0].(*parse.Source)
		src.Savepoint("s1")
		src.Consume1('a')
		must.Equal([]byte{'b'}, src.Peek())
		src.RollbackTo("s1")
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
	t.Run("consume then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 2))[0].(*parse.Source)
		src.Savepoint("s1")
		src.Consume()
		must.Equal([]byte{'c', 'd'}, src.Peek())
		src.RollbackTo("s1")
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
	t.Run("consume twice then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		src.Savepoint("s1")
		src.Consume()
		src.Consume()
		must.Equal([]byte{'e', 'f'}, src.Peek())
		src.RollbackTo("s1")
		must.Equal([]byte{'a', 'b'}, src.Peek())
		src.Consume()
		must.Equal([]byte{'c', 'd'}, src.Peek())
	}))
	t.Run("rollback two savepoints", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 1))[0].(*parse.Source)
		src.Savepoint("s1")
		src.Consume()
		src.Savepoint("s2")
		src.Consume()
		must.Equal([]byte{'c'}, src.Peek())
		src.RollbackTo("s2")
		must.Equal([]byte{'b'}, src.Peek())
		src.RollbackTo("s1")
		must.Equal([]byte{'a'}, src.Peek())
	}))
}

func TestSource_PeekN(t *testing.T) {
	t.Run("n smaller than current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b'}, must.Call(src.PeekN, 2)[0])
	}))
	t.Run("no reader", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("abc")
		peeked, err := src.PeekN(4)
		must.Equal(io.EOF, err)
		must.Equal("abc", string(peeked))
	}))
	t.Run("peek next", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c'}, must.Call(src.PeekN, 3)[0])
		must.Equal([]byte{'a', 'b', 'c'}, must.Call(src.PeekN, 3)[0])
	}))
	t.Run("peek next next", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, must.Call(src.PeekN, 5)[0])
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, must.Call(src.PeekN, 5)[0])
	}))
	t.Run("peek beyond end", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abc"), make([]byte, 2))[0].(*parse.Source)
		peeked, err := src.PeekN(5)
		must.Equal([]byte{'a', 'b', 'c'}, peeked)
		must.NotNil(err)
		peeked, err = src.PeekN(5)
		must.Equal([]byte{'a', 'b', 'c'}, peeked)
		must.NotNil(err)
	}))
}

func TestSource_Peek1(t *testing.T) {
	t.Run("peek1", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("abc")
		must.Equal(uint8('a'), src.Peek1())
	}))
	t.Run("empty string", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("")
		must.NotNil(src.Error())
		must.Panic(func() {
			src.Peek1()
		})
	}))
	t.Run("empty reader", test.Case(func(ctx context.Context) {
		src, err := parse.NewSource(strings.NewReader(""), make([]byte, 1))
		must.NotNil(err)
		must.Panic(func() {
			src.Peek1()
		})
	}))
}

func TestSource_Consume1(t *testing.T) {
	t.Run("consume not match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2")
		src.Consume1('b')
		must.NotNil(src.Error())
	}))
	t.Run("consume match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2")
		src.Consume1('h')
		must.Nil(src.Error())
	}))
}

func TestSource_PeekRune(t *testing.T) {
	t.Run("rune in current buf", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h")
		must.Equal('h', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xC2, 0xA2}))
		must.Equal('¢', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Equal('€', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Equal('𐍈', must.Call(src.PeekRune)[0])
	}))
	t.Run("rune in multiple buf", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSource(bytes.NewBufferString("h"), make([]byte, 1))
		must.Equal('h', must.Call(src.PeekRune)[0])
		src, _ = parse.NewSource(bytes.NewReader([]byte{0xC2, 0xA2}), make([]byte, 1))
		must.Equal('¢', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Equal('€', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Equal('𐍈', must.Call(src.PeekRune)[0])
	}))
}

func TestSource_SetBuffer(t *testing.T) {
	t.Run("set buffer to new slice to avoid current buffer being reused", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		peeked := src.Peek()
		// without SetBuffer, peeked will change
		src.SetBuffer(make([]byte, 2))
		src.Consume()
		must.Equal([]byte{'a', 'b'}, peeked)
	}))
}

func TestSource_PeekUtf8(t *testing.T) {
	t.Run("rune in multiple buf", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSource(bytes.NewBufferString("h"), make([]byte, 1))
		must.Equal([]byte("h"), must.Call(src.PeekUtf8)[0])
		src, _ = parse.NewSource(bytes.NewReader([]byte{0xC2, 0xA2}), make([]byte, 1))
		must.Equal([]byte("¢"), must.Call(src.PeekUtf8)[0])
		src = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Equal([]byte("€"), must.Call(src.PeekUtf8)[0])
		src = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Equal([]byte("𐍈"), must.Call(src.PeekUtf8)[0])
	}))
}
