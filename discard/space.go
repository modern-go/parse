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
