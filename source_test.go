package parse_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/modern-go/parse"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
)

func TestSource_Savepoint(t *testing.T) {
	t.Run("rollback should delete the savepoint", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("hello")
		must.Nil(err)
		src.StoreSavepoint()
		src.RollbackToSavepoint()
		src.StoreSavepoint()
		must.Nil(src.Error())
	}))
	t.Run("rollback not existing savepoint", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("hello")
		must.Nil(err)
		src.RollbackToSavepoint()
		must.NotNil(src.Error())
	}))
	t.Run("recursive savepoint", test.Case(func(ctx context.Context) {
		data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		reader := bytes.NewReader(data)
		src, err := parse.NewSource(reader, 4)
		must.AssertNil(err)

		// store the first 4 bytes
		first := src.PeekN(4)
		must.Equal([]byte{1, 2, 3, 4}, first)
		must.Equal(first, src.PeekN(4))
		// break point 1
		src.StoreSavepoint()

		//consume 4 bytes
		src.ReadN(4)
		// store the next 4 bytes
		second := src.PeekN(4)
		must.Equal([]byte{5, 6, 7, 8}, second)
		must.Equal(second, src.PeekN(4))
		// break point 2
		src.StoreSavepoint()

		// consume 4 more bytes
		src.ReadN(4)

		// jump back to break point 2
		src.RollbackToSavepoint()
		must.Equal(second, src.PeekN(4))

		// jump back to break point 1
		src.RollbackToSavepoint()
		must.Equal(first, src.PeekN(4))
	}))
}

func TestSource_RollbackTo(t *testing.T) {
	t.Run("immediate rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), 1)[0].(*parse.Source)
		src.StoreSavepoint()
		src.RollbackToSavepoint()
		must.Equal([]byte{'a'}, src.PeekN(1))
	}))
	t.Run("partial consume then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), 2)[0].(*parse.Source)
		src.StoreSavepoint()
		must.Equal(true, src.Expect1('a'))
		must.Equal([]byte{'b', 'c'}, src.PeekN(2))
		src.RollbackToSavepoint()
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
	t.Run("consume then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), 2)[0].(*parse.Source)
		src.StoreSavepoint()
		src.ReadN(2)
		must.Equal([]byte{'c', 'd'}, src.PeekN(2))
		src.RollbackToSavepoint()
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
	t.Run("consume twice then rollback", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), 2)[0].(*parse.Source)
		src.StoreSavepoint()
		src.ReadN(2)
		src.ReadN(2)
		must.Equal([]byte{'e', 'f'}, src.PeekN(2))
		src.RollbackToSavepoint()
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
		src.ReadN(2)
		must.Equal([]byte{'c', 'd'}, src.PeekN(2))
	}))
	t.Run("rollback two savepoints", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), 1)[0].(*parse.Source)
		src.StoreSavepoint()
		src.Read1()
		src.StoreSavepoint()
		src.Read1()
		must.Equal([]byte{'c'}, src.PeekN(1))
		src.RollbackToSavepoint()
		must.Equal([]byte{'b'}, src.PeekN(1))
		src.RollbackToSavepoint()
		must.Equal([]byte{'a'}, src.PeekN(1))
	}))
}

func TestSource_PeekN(t *testing.T) {
	t.Run("n smaller than current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), 2)[0].(*parse.Source)
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
	t.Run("no reader", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("abc")
		must.Nil(err)
		peeked := src.PeekN(4)
		must.Equal("abc", string(peeked))
	}))
	t.Run("peek next", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), 2)[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c'}, src.PeekN(3))
		must.Equal([]byte{'a', 'b', 'c'}, src.PeekN(3))
	}))
	t.Run("peek next next", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcdef"), 2)[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, src.PeekN(5))
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, src.PeekN(5))
	}))
	t.Run("peek beyond end", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abc"), 2)[0].(*parse.Source)
		peeked := src.PeekN(5)
		must.Equal([]byte{'a', 'b', 'c'}, peeked)
		peeked = src.PeekN(5)
		must.Equal([]byte{'a', 'b', 'c'}, peeked)
	}))
}

func TestSource_Peek1(t *testing.T) {
	t.Run("peek1", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("abc")
		must.Nil(err)
		must.Equal(uint8('a'), src.Peek1())
	}))
	t.Run("empty string", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("")
		must.NotNil(err)
		must.Panic(func() {
			src.Peek1()
		})
	}))
	t.Run("empty reader", test.Case(func(ctx context.Context) {
		src, err := parse.NewSource(strings.NewReader(""), 1)
		must.NotNil(err)
		must.Panic(func() {
			src.Peek1()
		})
	}))
}

func TestSource_Expect1(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2")
		must.Nil(err)
		must.Equal(false, src.Expect1('b'))
		must.Nil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2")
		must.Nil(err)
		must.Equal(true, src.Expect1('h'))
		must.Nil(src.Error())
	}))
}

func TestSource_Expect2(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2b")
		must.Nil(err)
		must.Equal(false, src.Expect2('h', '3'))
		must.Nil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2b")
		must.Nil(err)
		must.Equal(true, src.Expect2('h', '2'))
		must.Nil(src.Error())
	}))
}

func TestSource_Expect3(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2b~")
		must.Nil(err)
		must.Equal(false, src.Expect3('h', '2', 'c'))
		must.Nil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2b~")
		must.Nil(err)
		must.Equal(true, src.Expect3('h', '2', 'b'))
		must.Nil(src.Error())
	}))
}

func TestSource_Expect4(t *testing.T) {
	t.Run("not match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2bc~")
		must.Nil(err)
		must.Equal(false, src.Expect4('h', '2', 'c', 'd'))
		must.Nil(src.Error())
	}))
	t.Run("match", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h2bc~")
		must.Nil(err)
		must.Equal(true, src.Expect4('h', '2', 'b', 'c'))
		must.Nil(src.Error())
	}))
}

func TestSource_PeekRune(t *testing.T) {
	t.Run("rune in current buf", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("h")
		must.Nil(err)
		must.Equal('h', must.Call(src.PeekRune)[0])
		src, err = parse.NewSourceString(string([]byte{0xC2, 0xA2}))
		must.Nil(err)
		must.Equal('¢', must.Call(src.PeekRune)[0])
		src, err = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Nil(err)
		must.Equal('€', must.Call(src.PeekRune)[0])
		src, err = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Nil(err)
		must.Equal('𐍈', must.Call(src.PeekRune)[0])
	}))
	t.Run("rune in multiple buf", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSource(bytes.NewBufferString("h"), 1)
		must.Equal('h', must.Call(src.PeekRune)[0])
		src, _ = parse.NewSource(bytes.NewReader([]byte{0xC2, 0xA2}), 1)
		must.Equal('¢', must.Call(src.PeekRune)[0])
		src, _ = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Equal('€', must.Call(src.PeekRune)[0])
		src, _ = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Equal('𐍈', must.Call(src.PeekRune)[0])
	}))
}

func TestSource_PeekUtf8(t *testing.T) {
	t.Run("rune in multiple buf", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSource(bytes.NewBufferString("h"), 1)
		must.Equal([]byte("h"), must.Call(src.PeekUtf8)[0])
		src, _ = parse.NewSource(bytes.NewReader([]byte{0xC2, 0xA2}), 1)
		must.Equal([]byte("¢"), must.Call(src.PeekUtf8)[0])
		src, _ = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Equal([]byte("€"), must.Call(src.PeekUtf8)[0])
		src, _ = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Equal([]byte("𐍈"), must.Call(src.PeekUtf8)[0])
	}))
}

func TestSource_ReadN(t *testing.T) {
	t.Run("read from buffer", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString("abcdef")
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, src.ReadN(5))
		must.Equal([]byte{'f'}, src.PeekN(1))
	}))
	t.Run("read from reader", test.Case(func(ctx context.Context) {
		src, err := parse.NewSource(strings.NewReader("abcdef"), 2)
		must.Nil(err)
		must.Equal([]byte{'a', 'b', 'c', 'd', 'e'}, src.ReadN(5))
		must.Equal([]byte{'f'}, src.PeekN(1))
	}))
	t.Run("should report unexpected eof when not fully read", test.Case(func(ctx context.Context) {
		src, err := parse.NewSource(strings.NewReader("abcdef"), 2)
		must.Nil(err)
		src.ReadN(7)
		must.Equal(io.ErrUnexpectedEOF, src.Error())
	}))
}

func TestConsecutivePeek(t *testing.T) {
	t.Run("consecutive peek", test.Case(func(ctx context.Context) {
		data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		reader := bytes.NewReader(data)
		src, err := parse.NewSource(reader, 4)
		must.AssertNil(err)
		must.Equal(byte(1), src.Read1())
		// current 还剩3

		// Peek应该是可重入的，多次peek得到相同的结果
		expect := []byte{2, 3, 4, 5}
		must.Equal(expect, src.PeekN(4))
		must.Equal(expect, src.PeekN(4))
		must.Equal(expect, src.PeekN(4))
	}))
}

func TestReadAll(t *testing.T) {
	t.Run("partially consume and read all of the rest", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString("hello world")
		must.Equal(true, src.Expect1('h'))
		must.Equal(true, src.Expect1('e'))
		must.Equal([]byte("llo world"), src.ReadAll())
		must.Nil(src.Error())
	}))
}

func TestPeek(t *testing.T) {
	t.Run("peek without consuming more data", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSource(strings.NewReader("hello world"), 4)
		first := src.Peek()
		must.Equal([]byte("hell"), first)
		must.Equal(first, src.ReadN(4))
		must.Equal(0, len(src.Peek()))
		// trigger consume
		must.Equal(byte('o'), src.Peek1())
		second := src.Peek()
		must.Equal([]byte("o wo"), second)
	}))
}
