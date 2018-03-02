package discard

import (
	"github.com/modern-go/parse"
	"unicode"
)

// UnicodeRange discard the runes until found one not matching the range table
func UnicodeRange(src *parse.Source, table *unicode.RangeTable) int {
	count := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if !unicode.Is(table, r) {
			return count
		}
		src.ConsumeN(n)
		count += n
	}
	return count
}

// UnicodeRanges discard the runes until found one not matching the range tables
func UnicodeRanges(src *parse.Source, includes []*unicode.RangeTable, excludes []*unicode.RangeTable) int {
	count := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if matchRanges(excludes, r) {
			return count
		}
		if len(includes) > 0 && !matchRanges(includes, r) {
			return count
		}
		src.ConsumeN(n)
		count += n
	}
	return count
}

func matchRanges(ranges []*unicode.RangeTable, r rune) bool {
	for _, rng := range ranges {
		if unicode.Is(rng, r) {
			return true
		}
	}
	return false
}
