package hblade

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"go.uber.org/zap"
)

func tagMiddleware(tag string) Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			c.response.rw.Header().Add("X-Mw", tag)
			return next(c)
		}
	}
}

func TestSiblingGroupsDoNotShareMiddleware(t *testing.T) {
	app := New()
	g := app.Group("/g", tagMiddleware("a"))
	g.Use(tagMiddleware("b"))
	g.Use(tagMiddleware("c")) // leaves spare capacity in g.middleware
	one := g.Group("/one", tagMiddleware("one"))
	two := g.Group("/two", tagMiddleware("two"))
	one.Get("/x", func(c *Context) error { return nil })
	two.Get("/x", func(c *Context) error { return nil })

	for path, want := range map[string]string{"/g/one/x": "a,b,c,one", "/g/two/x": "a,b,c,two"} {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if got := strings.Join(w.Header().Values("X-Mw"), ","); got != want {
			t.Errorf("%s middleware = %q, want %q", path, got, want)
		}
	}
}

func TestStartReturnsAfterShutdown(t *testing.T) {
	Logger(zap.NewNop())
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()

	app := New()
	done := make(chan error, 1)
	go func() { done <- app.Start(addr) }()

	deadline := time.Now().Add(2 * time.Second)
	for {
		if conn, err := net.Dial("tcp", addr); err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after Shutdown")
	}
}

func decodeBody(t *testing.T, encoding string, body io.Reader) string {
	t.Helper()
	var r io.Reader = body
	switch encoding {
	case EncodingGzip:
		zr, err := gzip.NewReader(body)
		if err != nil {
			t.Fatal(err)
		}
		r = zr
	case EncodingZstd:
		zr, err := zstd.NewReader(body)
		if err != nil {
			t.Fatal(err)
		}
		defer zr.Close()
		r = zr
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(got)
}

func TestCompressionNegotiation(t *testing.T) {
	tests := []struct {
		name           string
		encodings      []string
		acceptEncoding string
		want           string
	}{
		{"disabled by default", nil, "zstd, gzip", ""},
		{"prefers zstd", []string{EncodingZstd, EncodingGzip}, "gzip, deflate, br, zstd", EncodingZstd},
		{"falls back to gzip", []string{EncodingZstd, EncodingGzip}, "gzip, deflate", EncodingGzip},
		{"honors q=0", []string{EncodingZstd, EncodingGzip}, "zstd;q=0, gzip;q=0.5", EncodingGzip},
		{"no supported encoding", []string{EncodingZstd, EncodingGzip}, "br", ""},
		{"server preference order", []string{EncodingGzip, EncodingZstd}, "zstd, gzip", EncodingGzip},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()
			app.SetCompression(tt.encodings...)
			app.Get("/t/:n", func(c *Context) error {
				return c.Text(strings.Repeat(c.Get("n"), 1000))
			})

			// Repeat requests so pooled encoders/buffers are reused.
			for _, n := range []string{"a", "b", "a"} {
				req := httptest.NewRequest(http.MethodGet, "/t/"+n, nil)
				req.Header.Set(acceptEncodingHeader, tt.acceptEncoding)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				if got := w.Header().Get(contentEncodingHeader); got != tt.want {
					t.Fatalf("Content-Encoding = %q, want %q", got, tt.want)
				}
				if got := decodeBody(t, tt.want, w.Body); got != strings.Repeat(n, 1000) {
					t.Fatalf("response %s corrupted: %.10q", n, got)
				}
			}
		})
	}
}

func TestShouldBindJSONKeepsBodyReadable(t *testing.T) {
	app := New()
	app.Post("/b", func(c *Context) error {
		var v struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&v); err != nil {
			return c.Error(http.StatusBadRequest, err)
		}
		rest, _ := io.ReadAll(c.Request().Req().Body)
		return c.Text(v.Name + "|" + string(rest))
	})

	body := `{"name":"hblade"}`
	req := httptest.NewRequest(http.MethodPost, "/b", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if got, want := w.Body.String(), "hblade|"+body; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
