package parse_test

import (
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
}
