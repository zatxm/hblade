package internal

import (
	"bytes"
	"io"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 4096))
	},
}

func ReadAll(r io.Reader) ([]byte, error) {
	buffer := pool.Get().(*bytes.Buffer)
	buffer.Reset()
	_, err := io.Copy(buffer, r)
	if err != nil {
		pool.Put(buffer)
		return []byte{}, err
	}
	pool.Put(buffer)
	return buffer.Bytes(), nil
}
