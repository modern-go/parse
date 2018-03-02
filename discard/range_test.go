package discard_test

import (
	"context"
	"github.com/modern-go/parse"
	"github.com/modern-go/parse/discard"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
	"strings"
	"testing"
	"unicode"
)

func TestUnicodeRange(t *testing.T) {
	t.Run("not found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(0, discard.UnicodeRange(src, unicode.White_Space))
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
	t.Run("skip partial current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader(" abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(1, discard.UnicodeRange(src, unicode.White_Space))
		must.Equal([]byte{'a'}, src.Peek())
	}))
	t.Run("skip all current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("  abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(2, discard.UnicodeRange(src, unicode.White_Space))
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
}

func TestUnicodeRanges(t *testing.T) {
	t.Run("skip all current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("  abcd"), make([]byte, 2))[0].(*parse.Source)
		must.Equal(2, discard.UnicodeRanges(src, []*unicode.RangeTable{
			unicode.White_Space,
			unicode.Han,
		}, nil))
		must.Equal([]byte{'a', 'b'}, src.Peek())
	}))
}
