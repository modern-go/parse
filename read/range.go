package read

import (
	"github.com/modern-go/parse"
	"unicode"
	"unicode/utf8"
)

// UnicodeRange read unicode until one not in table,
// the bytes will be appended to the space passed in.
// If space is nil, new space will be allocated from heap
func UnicodeRange(src *parse.Source, space []byte, table *unicode.RangeTable) []byte {
	for src.Error() == nil {
		buf := src.PeekUtf8()
		r, _ := utf8.DecodeRune(buf)
		if unicode.Is(table, r) {
			space = append(space, buf...)
			src.ConsumeN(len(buf))
			continue
		}
		return space
	}
	return space
}

// UnicodeRanges read unicode until one not in included table or encounteredd one in excluded table
func UnicodeRanges(src *parse.Source, space []rune, includes []*unicode.RangeTable, excludes []*unicode.RangeTable) []rune {
	for {
		r, n := src.PeekRune()
		for _, exclude := range excludes {
			if unicode.Is(exclude, r) {
				return space
			}
		}
		if len(includes) == 0 {
			src.ConsumeN(n)
			space = append(space, r)
			continue
		}
		for _, include := range includes {
			if unicode.Is(include, r) {
				src.ConsumeN(n)
				space = append(space, r)
				continue
			}
		}
		return space
	}
}
