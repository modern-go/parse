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

// Space reads consecutive space(\t \n \v \f \r ' ') and returns the space number
func Space(src *parse.Source) int {
	return Range(src, []byte{'\t', '\n', '\v', '\f', '\r', ' '})
}
