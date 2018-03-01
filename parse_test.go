package parse_test

import (
	"context"
	"github.com/modern-go/parse"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
	"testing"
)

func TestString(t *testing.T) {
	t.Run("no error", test.Case(func(ctx context.Context) {
		parsed := must.Call(parse.String, "abc", &myLexer{})[0]
		must.Equal(uint8('a'), parsed)
	}))
	t.Run("can not parse", test.Case(func(ctx context.Context) {
		parsed, err := parse.String("bc", &myLexer{})
		must.NotNil(err)
		must.Nil(parsed)
	}))
	t.Run("EOF", test.Case(func(ctx context.Context) {
		parsed := must.Call(parse.String, "a", &myLexer{})[0]
		must.Equal(uint8('a'), parsed)
	}))
}

type myLexer struct {
}

func (lexer *myLexer) PrefixToken(src *parse.Source) parse.PrefixToken {
	switch src.Peek1() {
	case 'a':
		return &myToken{}
	default:
		return nil
	}
}

func (lexer *myLexer) InfixToken(src *parse.Source) (parse.InfixToken, int) {
	return nil, 0
}

type myToken struct {
}

func (token *myToken) PrefixParse(src *parse.Source) interface{} {
	b := src.Peek1()
	src.ConsumeN(1)
	return b
}
