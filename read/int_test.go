package read_test

import (
	"context"
	"github.com/modern-go/parse"
	"github.com/modern-go/parse/read"
	"github.com/modern-go/test"
	"github.com/modern-go/test/must"
	"io"
	"strings"
	"testing"
)

func Test_ConsumeUint64_from_string(t *testing.T) {
	type TestCase struct {
		Input  string
		Output uint64
	}
	testCases := []TestCase{{
		"1", 1,
	}, {
		"12", 12,
	}, {
		"123", 123,
	}}
	for _, tmp := range testCases {
		testCase := tmp
		t.Run(testCase.Input, test.Case(func(ctx context.Context) {
			src := parse.NewSourceString(testCase.Input)
			must.Equal(testCase.Output, read.Uint64(src))
		}))
	}
	t.Run("Overflow", test.Case(func(ctx context.Context) {
		src := parse.NewSourceString("18446744073709551615")
		must.Equal(uint64(18446744073709551615), read.Uint64(src))
		must.Equal(io.EOF, src.Error())
		src = parse.NewSourceString("18446744073709551616")
		must.Equal(uint64(0), read.Uint64(src))
		must.NotNil(src.Error())
		must.Pass(io.EOF != src.Error())
	}))
}

func Test_ConsumeUint64_from_reader(t *testing.T) {
	type TestCase struct {
		Input    string
		Output   uint64
		Selected bool
	}
	testCases := []TestCase{{
		Input: "1", Output: 1,
	}, {
		Input: "12", Output: 12,
	}, {
		Input: "123", Output: 123,
	}, {
		Input: "1234", Output: 1234, Selected: true,
	}, {
		Input: "12345", Output: 12345,
	}}
	for _, testCase := range testCases {
		if testCase.Selected {
			testCases = []TestCase{testCase}
			break
		}
	}
	for _, tmp := range testCases {
		testCase := tmp
		t.Run(testCase.Input, test.Case(func(ctx context.Context) {
			src := must.Call(parse.NewSource,
				strings.NewReader(testCase.Input), make([]byte, 2))[0].(*parse.Source)
			must.Equal(testCase.Output, read.Uint64(src))
		}))
	}
}
