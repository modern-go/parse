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

// Until2 read any byte except b1 or b2.
// If b1 not found, report error.
func Until2(src *parse.Source, b1 byte, b2 byte) []byte {
	var buf []byte
	for src.Error() == nil {
		b := src.Peek1()
		if b == b1 || b == b2 || src.Error() != nil {
			break
		}
		src.Read1()
		buf = append(buf, b)
	}
	return buf
}

// AnyExcept1 read bytes until EOF, ignore b1
func AnyExcept1(src *parse.Source, b1 byte) []byte {
	var buf []byte
	b := src.Read1()
	for src.Error() == nil {
		if b != b1 {
			buf = append(buf, b)
		}
		b = src.Read1()
	}
	return buf
}

// AnyExcepts read bytes until EOF, ignore bs
func AnyExcepts(src *parse.Source, bs []byte) []byte {
	var buf []byte
	b := src.Read1()
	for src.Error() == nil {
		found := false
		for _, b1 := range bs {
			if b1 == b {
				found = true
				break
			}
		}
		if !found {
			buf = append(buf, b)
		}
		b = src.Read1()
	}
	return buf
}
