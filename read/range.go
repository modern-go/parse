package read

import (
	"github.com/modern-go/parse"
	"unicode"
)

// UnicodeRange read unicode until one not in table,
// the bytes will be appended to the space passed in.
// If space is nil, new space will be allocated from heap
func UnicodeRange(src *parse.Source, space []byte, table *unicode.RangeTable) []byte {
	src.StoreSavepoint()
	length := 0
	for src.Error() == nil {
		r, n := src.PeekRune()
		if !unicode.Is(table, r) {
			break
		}
		length += n
		src.ConsumeN(n)
	}
	if src.FatalError() != nil {
		src.DeleteSavepoint()
		return nil
	}
	src.RollbackToSavepoint()
	return src.CopyN(space, length)
}

// UnicodeRanges read unicode until one not in included table or encounteredd one in excluded table
func UnicodeRanges(src *parse.Source, space []byte, includes []*unicode.RangeTable, excludes []*unicode.RangeTable) []byte {
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
		src.ConsumeN(n)
	}
	if src.FatalError() != nil {
		src.DeleteSavepoint()
		return nil
	}
	src.RollbackToSavepoint()
	return src.CopyN(space, length)
}

func matchRanges(ranges []*unicode.RangeTable, r rune) bool {
	for _, rng := range ranges {
		if unicode.Is(rng, r) {
			return true
		}
	}
	return false
}
