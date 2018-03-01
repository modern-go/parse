package parse

import (
	"errors"
	"github.com/modern-go/concurrent"
	"io"
	"reflect"
)

func String(input string, lexer Lexer) (interface{}, error) {
	src := NewSourceString(input)
	left := Parse(src, lexer, 0)
	if src.Error() != nil {
		if src.Error() == io.EOF {
			return left, nil
		}
		return nil, src.Error()
	}
	return left, nil
}

func Parse(src *Source, lexer Lexer, precedence int) interface{} {
	token := lexer.PrefixToken(src)
	if token == nil {
		src.ReportError(errors.New("can not parse"))
		return nil
	}
	InfoLogger.Println("prefix", ">>>", reflect.TypeOf(token))
	left := token.PrefixParse(src)
	InfoLogger.Println("prefix", "<<<", reflect.TypeOf(token))
	for {
		if src.Error() != nil {
			return left
		}
		token, infixPrecedence := lexer.InfixToken(src)
		if token == nil {
			return left
		}
		if precedence >= infixPrecedence {
			concurrent.InfoLogger.Println("precedence skip ", reflect.TypeOf(token), precedence, infixPrecedence)
			return left
		}
		concurrent.InfoLogger.Println("infix ", ">>>", reflect.TypeOf(token))
		left = token.InfixParse(src, left)
		concurrent.InfoLogger.Println("infix ", "<<<", reflect.TypeOf(token))
	}
	return left
}

const DefaultPrecedence = 1

type PrefixToken interface {
	PrefixParse(src *Source) interface{}
}

type InfixToken interface {
	InfixParse(src *Source, left interface{}) interface{}
}

type Lexer interface {
	PrefixToken(src *Source) PrefixToken
	InfixToken(src *Source) (InfixToken, int)
}
