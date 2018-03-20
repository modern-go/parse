package pjson

import (
	"github.com/modern-go/parse"
	"errors"
	"bytes"
	"github.com/modern-go/reflect2"
)

func NextToken(src *parse.Source) byte {
	for src.Error() == nil {
		buf := src.Peek()
		for i := 0; i < len(buf); i++ {
			b := buf[i]
			switch b {
			case '\t', '\n', '\v', '\f', '\r', ' ':
				continue
			default:
				src.ConsumeN(i)
				return b
			}
		}
		src.Consume()
	}
	return 0
}

func ConsumeArrayStart(src *parse.Source) {
	b := NextToken(src)
	if b == '[' {
		src.ConsumeN(1)
		return
	}
	src.ReportError(errors.New("expect ["))
}

func ConsumeObjectStart(src *parse.Source) {
	b := NextToken(src)
	if b == '{' {
		src.ConsumeN(1)
		return
	}
	src.ReportError(errors.New("expect {"))
}

func ExpectPlainString(src *parse.Source, expectStr string) bool {
	b := NextToken(src)
	if b != '"' {
		return false
	}
	expectBytes := reflect2.UnsafeCastString(expectStr)
	expectLen := len(expectStr)
	expectLenPlusOne := expectLen + 1
	expectLenPlusTwo := expectLen + 2
	actualBytes, _ := src.PeekN(expectLenPlusTwo)
	if bytes.Equal(actualBytes[1:expectLenPlusOne], expectBytes) &&
		actualBytes[expectLenPlusOne] == '"' {
		src.ConsumeN(expectLenPlusTwo)
		return true
	}
	return false
}
