package hblade

import (
	"bytes"
	stdContext "context"
	"io"
	"net/http"

	"github.com/zatxm/hblade/v3/tools"
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

func (r *request) RawData() (b []byte, err error) {
	return tools.ReadAll(r.req.Body)
}

func (r *request) RawDataSetBody() (b []byte, err error) {
	b, err = tools.ReadAll(r.req.Body)
	if err != nil {
		return
	}
	r.req.Body = io.NopCloser(bytes.NewBuffer(b))
	return
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
