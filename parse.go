package parse

import (
	"errors"
	"io"
	"reflect"
)

// String parse the string with provided lexer
func String(input string, lexer Lexer) (interface{}, error) {
	src, err := NewSourceString(input)
	if nil != err {
		return nil, err
	}
	left := Parse(src, lexer, 0)
	if src.Error() != nil {
		if src.Error() == io.EOF {
			return left, nil
		}
		return nil, src.Error()
	}
	return left, nil
}

// Parse parse the source with provided lexer, might call this recursively.
// If precedence > 0, some infix will be skipped due to precedence.
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
			InfoLogger.Println("precedence skip ", reflect.TypeOf(token), precedence, infixPrecedence)
			return left
		}
		InfoLogger.Println("infix ", ">>>", reflect.TypeOf(token))
		left = token.InfixParse(src, left)
		InfoLogger.Println("infix ", "<<<", reflect.TypeOf(token))
	}
}

// DefaultPrecedence should be used when precedence does not matter
const DefaultPrecedence = 1

// PrefixToken parse the source at prefix position
type PrefixToken interface {
	PrefixParse(src *Source) interface{}
}

// InfixToken parse the source at infix position
type InfixToken interface {
	InfixParse(src *Source, left interface{}) interface{}
}

// Lexer tell the current token in the head of source
type Lexer interface {
	PrefixToken(src *Source) PrefixToken
	InfixToken(src *Source) (InfixToken, int)
}
