package discard_test

import (
	"context"
	"strings"
	"testing"
	"unicode"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/discard"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
)

func TestUnicodeRange(t *testing.T) {
	t.Run("not found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), 2)[0].(*parse.Source)
		must.Equal(0, discard.UnicodeRange(src, unicode.White_Space))
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
	t.Run("skip partial current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader(" abcd"), 2)[0].(*parse.Source)
		must.Equal(1, discard.UnicodeRange(src, unicode.White_Space))
		must.Equal([]byte{'a'}, src.PeekN(1))
	}))
	t.Run("skip all current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("  abcd"), 2)[0].(*parse.Source)
		must.Equal(2, discard.UnicodeRange(src, unicode.White_Space))
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
}

func TestUnicodeRanges(t *testing.T) {
	t.Run("skip all current", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("  abcd"), 2)[0].(*parse.Source)
		must.Equal(2, discard.UnicodeRanges(src, []*unicode.RangeTable{
			unicode.White_Space,
			unicode.Han,
		}, nil))
		must.Equal([]byte{'a', 'b'}, src.PeekN(2))
	}))
}

func TestRane(t *testing.T) {
	t.Run("ignore consecutive chars", test.Case(func(ctx context.Context) {
		src, err := parse.NewSourceString("aaaaabckey:  value\r\n")
		must.Nil(err)
		count := discard.Range(src, []byte{'a', 'b', 'c'})
		must.Equal(7, count)
	}))
}
