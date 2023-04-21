package tools

import (
	"bytes"
	"io"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 10240))
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
	temp := buffer.Bytes()
	length := len(temp)
	var body []byte
	if cap(temp) > (length + length/10) {
		body = make([]byte, length)
		copy(body, temp)
	} else {
		body = temp
	}
	return body, nil
}
