package discard

import (
	"unicode"

	"github.com/modern-go/parse"
)

// UnicodeRange discard the runes until found one not matching the range table
func UnicodeRange(src *parse.Source, table *unicode.RangeTable) int {
	count := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if !unicode.Is(table, r) {
			return count
		}
		src.ReadN(n)
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
		src.ReadN(n)
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

// Range discard the bytes until found one not matching the range table
// it returns how many bytes it discard
func Range(src *parse.Source, target []byte) int {
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
