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
	t.Run("rollback should delete the savepoint", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("hello")
		src.StoreSavepoint()
		src.RollbackToSavepoint()
		src.StoreSavepoint()
		must.Nil(src.Error())
	}))
	t.Run("rollback not existing savepoint", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("hello")
		src.RollbackToSavepoint()
		must.NotNil(src.Error())
	}))
}

func TestSource_RollbackTo(t *testing.T) {
	t.Run("immediate rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 1))[0].(*parse.Source)
		src.StoreSavepoint()
		src.RollbackToSavepoint()
		must.Equal([]byte{'a'}, src.Peek())
	}))
	t.Run("partial consume then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 2))[0].(*parse.Source)
		src.StoreSavepoint()
		src.Expect1('a')
		must.Equal([]byte{'b'}, src.Peek())
		src.RollbackToSavepoint()
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
	t.Run("consume then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 2))[0].(*parse.Source)
		src.StoreSavepoint()
		src.Consume()
		must.Equal([]byte{'c', 'd'}, src.Peek())
		src.RollbackToSavepoint()
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
	t.Run("consume twice then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		src.StoreSavepoint()
		src.Consume()
		src.Consume()
		must.Equal([]byte{'e', 'f'}, src.Peek())
		src.RollbackToSavepoint()
		must.Equal([]byte{'a', 'b'}, src.Peek())
		src.Consume()
		must.Equal([]byte{'c', 'd'}, src.Peek())
	}))
	t.Run("rollback two savepoints", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 1))[0].(*parse.Source)
		src.StoreSavepoint()
		src.Consume()
		src.StoreSavepoint()
		src.Consume()
		must.Equal([]byte{'c'}, src.Peek())
		src.RollbackToSavepoint()
		must.Equal([]byte{'b'}, src.Peek())
		src.RollbackToSavepoint()
		must.Equal([]byte{'a'}, src.Peek())
	}))
}

func TestSource_PeekN(t *testing.T) {
	t.Run("n smaller than current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
	t.Run("no reader", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("abc")
		peeked := src.PeekN(4)
		must.Equal("abc", string(peeked))
	}))
	t.Run("peek next", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c'}, src.PeekN(3))
		must.Equal([]byte{'a', 'b', 'c'}, src.PeekN(3))
	}))
	t.Run("peek next next", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, src.PeekN(5))
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, src.PeekN(5))
	}))
	t.Run("peek beyond end", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abc"), make([]byte, 2))[0].(*parse.Source)
		peeked := src.PeekN(5)
		must.Equal([]byte{'a', 'b', 'c'}, peeked)
		peeked = src.PeekN(5)
		must.Equal([]byte{'a', 'b', 'c'}, peeked)
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

func TestSource_Expect1(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2")
		src.Expect1('b')
		must.NotNil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2")
		src.Expect1('h')
		must.Nil(src.Error())
	}))
}

func TestSource_Expect2(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2b")
		src.Expect2('h', '3')
		must.NotNil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2b")
		src.Expect2('h', '2')
		must.Nil(src.Error())
	}))
}

func TestSource_Expect3(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2b~")
		src.Expect3('h', '2', 'c')
		must.NotNil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2b~")
		src.Expect3('h', '2', 'b')
		must.Nil(src.Error())
	}))
}

func TestSource_Expect4(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2bc~")
		src.Expect4('h', '2', 'c', 'd')
		must.NotNil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h2bc~")
		src.Expect4('h', '2', 'b', 'c')
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

func TestSource_ConsumeN(t *testing.T) {
	t.Run("consume partial current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		src.ConsumeN(1)
		must.Equal([]byte{'b'}, src.Peek())
	}))
	t.Run("consume all current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		src.ConsumeN(2)
		must.Equal([]byte{'c', 'd'}, src.Peek())
	}))
	t.Run("consume two buf", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		src.ConsumeN(3)
		must.Equal([]byte{'d'}, src.Peek())
	}))
	t.Run("consume all", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		src.ConsumeN(6)
		must.Equal([]byte{}, src.Peek())
		must.Equal(io.EOF, src.Error())
	}))
	t.Run("consume beyond end", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		src.ConsumeN(7)
		must.Equal([]byte{}, src.Peek())
		must.Equal(io.ErrUnexpectedEOF, src.Error())
	}))
}

func TestSource_CopyN(t *testing.T) {
	t.Run("consume partial current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a'}, src.CopyN(1))
		must.Equal([]byte{'b'}, src.Peek())
	}))
	t.Run("consume all current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b'}, src.CopyN(2))
		must.Equal([]byte{'c', 'd'}, src.Peek())
	}))
	t.Run("consume two buf", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c'}, src.CopyN(3))
		must.Equal([]byte{'d'}, src.Peek())
	}))
	t.Run("consume all", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e', 'f'}, src.CopyN(6))
		must.Equal([]byte{}, src.Peek())
		must.Equal(io.EOF, src.Error())
	}))
	t.Run("consume beyond end", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e', 'f'}, src.CopyN(7))
		must.Equal([]byte{}, src.Peek())
		must.Equal(io.ErrUnexpectedEOF, src.Error())
	}))
}

func TestSource_ReadN(t *testing.T) {
	t.Run("read all", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), make([]byte, 2))[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e', 'f'}, src.ReadN(6))
		must.Equal([]byte{}, src.Peek())
		must.Equal(io.EOF, src.Error())
	}))
}
