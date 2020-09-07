package hblade

import "net/http"

type Response interface {
	Header(string) string
	Rw() http.ResponseWriter
	SetHeader(string, string)
	SetRw(http.ResponseWriter)
	Pusher() http.Pusher
}

type response struct {
	rw http.ResponseWriter
}

func (r *response) Header(key string) string {
	return r.rw.Header().Get(key)
}

func (r *response) SetHeader(key, value string) {
	r.rw.Header().Set(key, value)
}

func (r *response) Rw() http.ResponseWriter {
	return r.rw
}

func (r *response) SetRw(writer http.ResponseWriter) {
	r.rw = writer
}

func (r *response) Pusher() (pusher http.Pusher) {
	if pusher, ok := r.rw.(http.Pusher); ok {
		return pusher
	}
	return nil
}
