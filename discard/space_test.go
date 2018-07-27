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
