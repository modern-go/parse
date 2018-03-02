package discard_test

import (
	"context"
	"github.com/modern-go/parse"
	"github.com/modern-go/parse/discard"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
	"strings"
	"testing"
)

func TestSpace(t *testing.T) {
	t.Run("not found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(0, discard.Space(src))
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
	t.Run("found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader(" abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(1, discard.Space(src))
		must.Equal([]byte{'a'}, src.Peek())
	}))
}

func TestUnicodeSpace(t *testing.T) {
	t.Run("not found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(0, discard.UnicodeSpace(src))
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
	t.Run("found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader(" abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(1, discard.UnicodeSpace(src))
		must.Equal([]byte{'a'}, src.Peek())
	}))
}
