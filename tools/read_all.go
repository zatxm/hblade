package tools

import "io"

func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
