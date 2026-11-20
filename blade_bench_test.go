package hblade

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

type benchWriter struct {
	h http.Header
}

func (w *benchWriter) Header() http.Header         { return w.h }
func (w *benchWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *benchWriter) WriteHeader(int)             {}

func benchServe(b *testing.B, app *Blade, req *http.Request) {
	w := &benchWriter{h: make(http.Header)}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		clear(w.h)
		app.ServeHTTP(w, req)
	}
}

func BenchmarkParamRoute(b *testing.B) {
	app := New()
	app.Get("/user/:id/posts/:pid", func(c *Context) error {
		_ = c.Get("id")
		return c.Text(c.Get("pid"))
	})
	benchServe(b, app, httptest.NewRequest(http.MethodGet, "/user/123/posts/456", nil))
}

func benchCompressedJSON(b *testing.B, acceptEncoding string) {
	app := New()
	app.SetCompression(EncodingZstd, EncodingGzip)
	payload := map[string]string{"data": strings.Repeat("hblade ", 400)}
	app.Get("/json", func(c *Context) error { return c.JSON(payload) })
	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	req.Header.Set(acceptEncodingHeader, acceptEncoding)
	benchServe(b, app, req)
}

func BenchmarkGzipJSON(b *testing.B) { benchCompressedJSON(b, "gzip, deflate") }
func BenchmarkZstdJSON(b *testing.B) { benchCompressedJSON(b, "gzip, deflate, zstd") }

func BenchmarkQueryClientIP(b *testing.B) {
	app := New()
	app.Get("/q", func(c *Context) error {
		_ = c.Query("a")
		_ = c.Query("b")
		_ = c.Query("c")
		_ = c.ClientIP()
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/q?a=1&b=2&c=3", nil)
	req.Header.Set(forwardedForHeader, "10.0.0.1, 10.0.0.2")
	benchServe(b, app, req)
}

func BenchmarkBindJSON(b *testing.B) {
	type in struct {
		Name string `json:"name" binding:"required"`
		Age  int    `json:"age"`
	}
	app := New()
	app.Post("/bind", func(c *Context) error {
		var v in
		return c.ShouldBindJSON(&v)
	})
	body := []byte(`{"name":"hblade","age":5}`)
	req := httptest.NewRequest(http.MethodPost, "/bind", nil)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	rd := bytes.NewReader(body)
	w := &benchWriter{h: make(http.Header)}
	b.ReportAllocs()
	for b.Loop() {
		rd.Reset(body)
		req.Body = io.NopCloser(rd)
		app.ServeHTTP(w, req)
	}
}

func BenchmarkLogRequest(b *testing.B) {
	Logger(zap.NewNop())
	app := New()
	app.EnableLogRequest()
	app.Get("/log", func(c *Context) error { return nil })
	benchServe(b, app, httptest.NewRequest(http.MethodGet, "/log", nil))
}

func BenchmarkLogParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Log()
		}
	})
}
