package test

import (
	"context"
	"io"
	"testing"

	"github.com/modern-go/parse"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
)

func Test(t *testing.T) {
	t.Run("1＋1", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString(`1+1`)
		dst := expr.Parse(src, 0)
		must.Equal(io.EOF, src.Error())
		must.Equal(2, dst)
	}))
	t.Run("－1＋2", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString(`-1+2`)
		must.Equal(1, expr.Parse(src, 0))
	}))
	t.Run("1＋1－1", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString(`1+1-1`)
		must.Equal(1, expr.Parse(src, 0))
	}))
	t.Run("2×3＋1", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString(`2*3+1`)
		must.Equal(7, expr.Parse(src, 0))
	}))
	t.Run("4/2＋1", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString(`4/2+1`)
		must.Equal(3, expr.Parse(src, 0))
	}))
	t.Run("4/（1＋1）＋2", test.Case(func(ctx context.Context) {
		src, _ := parse.NewSourceString(`4/(1+1)+2`)
		must.Equal(4, expr.Parse(src, 0))
	}))
}

const precedenceAssignment = 1
const precedenceConditional = 2
const precedenceSum = 3
const precedenceProduct = 4
const precedenceExponent = 5
const precedencePrefix = 6
const precedencePostfix = 7
const precedenceCall = 8

type exprLexer struct {
	value    *valueToken
	plus     *plusToken
	minus    *minusToken
	multiply *multiplyToken
	divide   *divideToken
	group    *groupToken
}

var expr = newExprLexer()

func newExprLexer() *exprLexer {
	return &exprLexer{}
}

func (lexer *exprLexer) Parse(src *parse.Source, precedence int) interface{} {
	return parse.Parse(src, lexer, precedence)
}

func (lexer *exprLexer) InfixToken(src *parse.Source) (parse.InfixToken, int) {
	switch src.Peek1() {
	case '+':
		return lexer.plus, precedenceSum
	case '-':
		return lexer.minus, precedenceSum
	case '*':
		return lexer.multiply, precedenceProduct
	case '/':
		return lexer.divide, precedenceProduct
	default:
		return nil, 0
	}
}

func (lexer *exprLexer) PrefixToken(src *parse.Source) parse.PrefixToken {
	switch src.Peek1() {
	case '(':
		return lexer.group
	case '-':
		return lexer.minus
	default:
		return lexer.value
	}
}

type valueToken struct {
}

func (token *valueToken) PrefixParse(src *parse.Source) interface{} {
	return 0
	//return read.Int(src)
}

type plusToken struct {
}

func (token *plusToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	leftValue := left.(int)
	src.Expect1('+')
	rightValue := expr.Parse(src, precedenceSum).(int)
	return leftValue + rightValue
}

type minusToken struct {
}

func (token *minusToken) PrefixParse(src *parse.Source) interface{} {
	src.Expect1('-')
	expr := expr.Parse(src, precedencePrefix).(int)
	return -expr
}

func (token *minusToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	leftValue := left.(int)
	src.Expect1('-')
	rightValue := expr.Parse(src, precedenceSum).(int)
	return leftValue - rightValue
}

type multiplyToken struct {
}

func (token *multiplyToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	leftValue := left.(int)
	src.Expect1('*')
	rightValue := expr.Parse(src, precedenceProduct).(int)
	return leftValue * rightValue
}

type divideToken struct {
}

func (token *divideToken) InfixParse(src *parse.Source, left interface{}) interface{} {
	leftValue := left.(int)
	src.Expect1('/')
	rightValue := expr.Parse(src, precedenceProduct).(int)
	return leftValue / rightValue
}

type groupToken struct {
}

func (token *groupToken) PrefixParse(src *parse.Source) interface{} {
	src.Expect1('(')
	expr := expr.Parse(src, 0)
	src.Expect1(')')
	return expr
}
