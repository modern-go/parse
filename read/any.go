package read

import "github.com/modern-go/parse"

// AnyExcept1 read any byte except b1
func AnyExcept1(src *parse.Source, space []byte, b1 byte) []byte {
	for src.Error() == nil {
		buf := src.Peek()
		for i := 0; i < len(buf); i++ {
			b := buf[i]
			if b == b1 {
				space = append(space, buf[:i]...)
				src.ConsumeN(i)
				return space
			}
		}
		space = append(space, buf...)
		src.Consume()
	}
	return space
}

// AnyExcept2 read any byte except b1 and b2
func AnyExcept2(src *parse.Source, space []byte, b1 byte, b2 byte) []byte {
	for src.Error() == nil {
		buf := src.Peek()
		for i := 0; i < len(buf); i++ {
			b := buf[i]
			if b == b1 || b == b2 {
				space = append(space, buf[:i]...)
				src.ConsumeN(i)
				return space
			}
		}
		space = append(space, buf...)
		src.Consume()
	}
	return space
}
