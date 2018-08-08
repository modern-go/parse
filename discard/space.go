package discard

import (
	"unicode"

	"github.com/modern-go/parse"
)

// UnicodeSpace discard unicode spaces
func UnicodeSpace(src *parse.Source) int {
	count := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if unicode.IsSpace(r) {
			src.ReadN(n)
			count += n
			continue
		}
		return count
	}
	return count
}

// Skip read bytes until finding a byte not belong to the target
func Skip(src *parse.Source, target []byte) int {
	if src == nil {
		return 0
	}
	count := 0
	for src.Error() == nil {
		b := src.Peek1()
		found := false
		for _, t := range target {
			if b == t {
				found = true
				break
			}
		}
		if !found {
			break
		}
		count++
		src.Read1()
	}
	return count
}

// Space reads consecutive space(\t \n \v \f \r ' ') and returns the space number
func Space(src *parse.Source) int {
	return Skip(src, []byte{'\t', '\n', '\v', '\f', '\r', ' '})
}
