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

func TestExcept(t *testing.T) {
	t.Run("except one", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString("hello world 1234 bbb")
		data := read.AnyExcept1(src, ' ')
		must.Equal("helloworld1234bbb", string(data))
	}))
	t.Run("except multi", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString("hello world 1234 bbb")
		data := read.AnyExcepts(src, []byte{' ', 'o', 'b'})
		must.Equal("hellwrld1234", string(data))
	}))
}
