package parse_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/modern-go/parse"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
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

func TestPeek(t *testing.T) {
	t.Run("consecutive peek", test.Case(func(ctx context.Context) {
		data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		reader := bytes.NewReader(data)
		buf := make([]byte, 4)
		src, err := parse.NewSource(reader, buf)
		must.AssertNil(err)
		must.Equal(byte(1), src.Read1())
		// current 还剩3

		// Peek应该是可重入的，多次peek得到相同的结果
		expect := []byte{2, 3, 4, 5}
		for i := 0; i < 10; i++ {
			must.Equal(expect, src.PeekN(4))
		}
	}))
}

func TestRecursiveSavePoints(t *testing.T) {
	t.Run("recursive savepoint", test.Case(func(ctx context.Context) {
		data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		reader := bytes.NewReader(data)
		buf := make([]byte, 4)
		src, err := parse.NewSource(reader, buf)
		must.AssertNil(err)

		// store the first 4 bytes
		first := src.PeekN(4)
		must.Equal([]byte{1, 2, 3, 4}, first)
		for i := 0; i < 10; i++ {
			must.Equal(first, src.PeekN(4))
		}
		// break point 1
		src.StoreSavepoint()

		//consume 4 bytes
		src.ReadN(4)
		// store the next 4 bytes
		second := src.PeekN(4)
		must.Equal([]byte{5, 6, 7, 8}, second)
		for i := 0; i < 10; i++ {
			must.Equal(second, src.PeekN(4))
		}
		// break point 2
		src.StoreSavepoint()

		// consume 4 more bytes
		src.ReadN(4)

		// jump back to break point 2
		src.RollbackToSavepoint()
		for i := 0; i < 10; i++ {
			must.Equal(second, src.PeekN(4))
		}
		// jump back to break point 1
		src.RollbackToSavepoint()
		for i := 0; i < 10; i++ {
			must.Equal(second, src.PeekN(4))
		}
	}))
}
