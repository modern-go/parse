package read

import (
	"unicode"

	"github.com/modern-go/parse"
)

// UnicodeRange read unicode until one not in table,
// the bytes will be appended to the space passed in.
func UnicodeRange(src *parse.Source, table *unicode.RangeTable) []byte {
	src.StoreSavepoint()
	length := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if !unicode.Is(table, r) {
			break
		}
		length += n
		src.ReadN(n)
	}
	if src.FatalError() != nil {
		src.DeleteSavepoint()
		return nil
	}
	src.RollbackToSavepoint()
	return src.PeekN(length)
}

// UnicodeRanges read unicode until one not in included table or encounteredd one in excluded table
func UnicodeRanges(src *parse.Source, includes []*unicode.RangeTable, excludes []*unicode.RangeTable) []byte {
	src.StoreSavepoint()
	length := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if matchRanges(excludes, r) {
			break
		}
		if len(includes) > 0 && !matchRanges(includes, r) {
			break
		}
		length += n
		src.ReadN(n)
	}
	if src.FatalError() != nil {
		src.DeleteSavepoint()
		return nil
	}
	src.RollbackToSavepoint()
	return src.PeekN(length)
}

func matchRanges(ranges []*unicode.RangeTable, r rune) bool {
	for _, rng := range ranges {
		if unicode.Is(rng, r) {
			return true
		}
	}
	return false
}
