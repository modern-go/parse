package parse_test

import (
	"testing"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
	"github.com/modern-go/parse"
	"bytes"
	"context"
)

func TestSource_PeekRune(t *testing.T) {
	t.Run("rune in current buf", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("h")
		must.Equal('h', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xC2, 0xA2}))
		must.Equal('¢', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Equal('€', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Equal('𐍈', must.Call(src.PeekRune)[0])
	}))
	t.Run("rune in multiple buf", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSource(bytes.NewBufferString("h"), make([]byte, 1))
		must.Equal('h', must.Call(src.PeekRune)[0])
		src, _ = parse.NewSource(bytes.NewReader([]byte{0xC2, 0xA2}), make([]byte, 1))
		must.Equal('¢', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xE2, 0x82, 0xAC}))
		must.Equal('€', must.Call(src.PeekRune)[0])
		src = parse.NewSourceString(string([]byte{0xF0, 0x90, 0x8D, 0x88}))
		must.Equal('𐍈', must.Call(src.PeekRune)[0])
	}))
}
