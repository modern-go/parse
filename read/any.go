package read

import (
	"github.com/modern-go/parse"
)

// Until1 read any byte except b1.
// If b1 not found, report error.
func Until1(src *parse.Source, b1 byte) []byte {
	var buf []byte
	for src.Error() == nil {
		b := src.Peek1()
		if b == b1 || src.Error() != nil {
			break
		}
		src.Read1()
		buf = append(buf, b)
	}
	return buf
}
