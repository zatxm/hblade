package hblade

import (
	"bytes"
	stdContext "context"
	"io"
	"net/http"
)

type Request interface {
	RawData() ([]byte, error)
	RawDataSetBody() ([]byte, error)
	Context() stdContext.Context
	Header(string) string
	Host() string
	Method() string
	Path() string
	Protocol() string
	Scheme() string
	RawQuery() string
	ContentType() string
	Req() *http.Request
}

type request struct {
	req *http.Request
}

// Upper bound for trusting Content-Length to pre-size the body buffer.
const maxPreallocBody = 8 << 20

func readBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	if n := req.ContentLength; n > 0 && n <= maxPreallocBody {
		b := make([]byte, n)
		m, err := io.ReadFull(req.Body, b)
		return b[:m], err
	}
	return io.ReadAll(req.Body)
}

func (r *request) RawData() ([]byte, error) {
	return readBody(r.req)
}

func (r *request) RawDataSetBody() ([]byte, error) {
	b, err := readBody(r.req)
	if err != nil {
		return b, err
	}
	if r.req.Body != nil && r.req.Body != http.NoBody {
		r.req.Body = io.NopCloser(bytes.NewReader(b))
	}
	return b, nil
}

func (r *request) Context() stdContext.Context {
	return r.req.Context()
}

func (r *request) Header(key string) string {
	return r.req.Header.Get(key)
}

func (r *request) Method() string {
	return r.req.Method
}

func (r *request) Protocol() string {
	return r.req.Proto
}

func (r *request) Host() string {
	return r.req.Host
}

func (r *request) Path() string {
	return r.req.URL.Path
}

func (r *request) RawQuery() string {
	return r.req.URL.RawQuery
}

func (r *request) Scheme() string {
	scheme := r.Header("X-Forwarded-Proto")
	if scheme != "" {
		return scheme
	}

	if r.req.TLS != nil {
		return "https"
	}

	return "http"
}

func (r *request) ContentType() string {
	return filterFlags(r.Header("Content-Type"))
}

func (r *request) Req() *http.Request {
	return r.req
}
