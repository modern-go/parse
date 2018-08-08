package discard_test

import (
	"context"
	"strings"
	"testing"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/discard"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
)

func TestUnicodeSpace(t *testing.T) {
	t.Run("not found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), 2)[0].(*parse.Source)
		must.Equal(0, discard.UnicodeSpace(src))
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
	t.Run("found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader(" abcd"), 2)[0].(*parse.Source)
		must.Equal(1, discard.UnicodeSpace(src))
		must.Equal([]byte{'a'}, src.PeekN(1))
	}))
}

func TestSkip(t *testing.T) {
	t.Run("ignore consecutive chars", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("aaaaabckey:  value\r\n")
		must.Nil(err)
		count := discard.Skip(src, []byte{'a', 'b', 'c'})
		must.Equal(7, count)
	}))
}

func TestSpace(t *testing.T) {
	t.Run("ignore consecutive space", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("\t\r\n\f\v  value\r\n")
		must.Nil(err)
		count := discard.Space(src)
		must.Equal(7, count)
		must.Equal("value", string(src.ReadN(5)))
	}))
	t.Run("no space", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("v  alue\r\n")
		must.Nil(err)
		count := discard.Space(src)
		must.Equal(0, count)
		must.Equal(byte('v'), src.Peek1())
	}))
}
