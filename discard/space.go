package discard

import (
	"github.com/modern-go/parse"
	"unicode"
)

// Space discard ascii spaces
func Space(src *parse.Source) int {
	count := 0
	for src.Error() == nil {
		buf := src.Peek()
		for i := 0; i < len(buf); i++ {
			switch buf[i] {
			case '\t', '\n', '\v', '\f', '\r', ' ':
				count++
				continue
			default:
				src.ConsumeN(i)
				return count
			}
		}
		src.Consume()
	}
	return count
}

// UnicodeSpace discard unicode spaces
func UnicodeSpace(src *parse.Source) int {
	count := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if unicode.IsSpace(r) {
			src.ConsumeN(n)
			count += n
			continue
		}
		return count
	}
	return count
}
