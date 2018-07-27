package read_test

import (
	"context"
	"strings"
	"testing"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/read"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
)

func TestUntil1(t *testing.T) {
	t.Run("not found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), 2)[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c', 'd'}, read.Until1(
			src, 'g'))
		must.NotNil(src.FatalError())
	}))
	t.Run("found", test.Case(func(ctx context.Context) {
		src := must.Call(parse.NewSource,
			strings.NewReader("abcd"), 2)[0].(*parse.Source)
		must.Equal([]byte{'a', 'b', 'c'}, read.Until1(
			src, 'd'))
		must.Nil(src.FatalError())
	}))
}
