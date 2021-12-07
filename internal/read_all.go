package internal

import (
	"bytes"
	"io"
	"sync"
	"unsafe"
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

	return Str2bytes(string(buffer.Bytes())), nil
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
