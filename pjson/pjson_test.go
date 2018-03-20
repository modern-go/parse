package pjson_test

import (
	"testing"
	"github.com/modern-go/test"
	"context"
	"github.com/modern-go/parse"
	"github.com/modern-go/parse/pjson"
	"github.com/modern-go/test/must"
)

func TestExpectPlainString(t *testing.T) {
	t.Run("found", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString(`"abc"`)
		must.Pass(pjson.ExpectPlainString(src, "abc"))
	}))
	t.Run("not found", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString(`123`)
		must.Pass(!pjson.ExpectPlainString(src, "abc"))
		src = parse.NewSourceString(`"abcd"`)
		must.Pass(!pjson.ExpectPlainString(src, "abc"))
	}))
}

func TestConsumeObjectStart(t *testing.T) {
	t.Run("found", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString(`{"abc":"def"}`)
		pjson.ConsumeObjectStart(src)
		must.Pass(pjson.ExpectPlainString(src, "abc"))
	}))
}

func BenchmarkExpectPlainString(b *testing.B) {
	src := parse.NewSourceString(`"abcdefgabcdefg"`)
	buf := src.Peek()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		src.Reset(nil, buf)
		pjson.ExpectPlainString(src, "abcdefgabcdefg")
	}
}