# parse

[![Sourcegraph](https://sourcegraph.com/github.com/modern-go/parse/-/badge.svg)](https://sourcegraph.com/github.com/modern-go/parse?badge)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/modern-go/parse)
[![Build Status](https://travis-ci.org/modern-go/parse.svg?branch=master)](https://travis-ci.org/modern-go/parse)
[![codecov](https://codecov.io/gh/modern-go/parse/branch/master/graph/badge.svg)](https://codecov.io/gh/modern-go/parse)
[![rcard](https://goreportcard.com/badge/github.com/modern-go/parse)](https://goreportcard.com/report/github.com/modern-go/parse)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/modern-go/parse/master/LICENSE)

pratt parser framework, implements https://tdop.github.io/

* the main parse loop: plugin in your lexer and token, we can parse anything
* a look-ahead parser source can read byte by byte, or rune by rune
* reusable parsing sub-routines to `read` or `discard` frequently used sequence types, like space, numeric

here is an example

```go
src := parse.NewSourceString(`4/(1+1)+2`)
parsed := parse.Parse(src, newExprLexer(), 0)
fmt.Println(parsed) // 4
```

the parser implementation is very short

```go
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
	return read.Int(src)
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
```
