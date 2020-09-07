package hblade

import (
	stdContext "context"
	"io/ioutil"
	"net/http"
)

type Request interface {
	RawData() ([]byte, error)
	Context() stdContext.Context
	Header(string) string
	Host() string
	Method() string
	Path() string
	Protocol() string
	Scheme() string
	RawQuery() string
	Req() *http.Request
}

type request struct {
	req *http.Request
}

func (r *request) RawData() ([]byte, error) {
	return ioutil.ReadAll(r.req.Body)
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

func (r *request) Req() *http.Request {
	return r.req
}
