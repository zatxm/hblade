package hblade

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

type Blade struct {
	router      Router
	middleware  []Middleware
	contextPool sync.Pool
	noFunc      func(*Context) //404
}

// New creates a new blade.
func New() *Blade {
	b := &Blade{
		noFunc: nil,
	}

	// Context pool
	b.contextPool.New = func() interface{} {
		return &Context{b: b}
	}

	return b
}

func (b *Blade) NoFunc(f func(*Context)) {
	b.noFunc = f
}

// Get registers your function to be called when the given GET path has been requested.
func (b *Blade) Get(path string, handler Handler) {
	b.router.Add(http.MethodGet, path, handler)
	b.BindMiddleware()
}

// Post registers your function to be called when the given POST path has been requested.
func (b *Blade) Post(path string, handler Handler) {
	b.router.Add(http.MethodPost, path, handler)
	b.BindMiddleware()
}

// Delete registers your function to be called when the given DELETE path has been requested.
func (b *Blade) Delete(path string, handler Handler) {
	b.router.Add(http.MethodDelete, path, handler)
	b.BindMiddleware()
}

// Put registers your function to be called when the given PUT path has been requested.
func (b *Blade) Put(path string, handler Handler) {
	b.router.Add(http.MethodPut, path, handler)
	b.BindMiddleware()
}

// Patch registers your function to be called when the given PATCH path has been requested.
func (b *Blade) Patch(path string, handler Handler) {
	b.router.Add(http.MethodPatch, path, handler)
	b.BindMiddleware()
}

// Options registers your function to be called when the given OPTIONS path has been requested.
func (b *Blade) Options(path string, handler Handler) {
	b.router.Add(http.MethodOptions, path, handler)
	b.BindMiddleware()
}

// Head registers your function to be called when the given HEAD path has been requested.
func (b *Blade) Head(path string, handler Handler) {
	b.router.Add(http.MethodHead, path, handler)
	b.BindMiddleware()
}

// Can bind static directory
// h.Static("/static", "static/")
func (b *Blade) Static(path, bind string) {
	relativePath := path + "/*file"
	handler := func(c *Context) error {
		return c.File(bind + c.Get("file"))
	}
	b.Get(relativePath, handler)
}

// Any registers your function to be called with any http method.
func (b *Blade) Any(path string, handler Handler) {
	b.router.Add(http.MethodGet, path, handler)
	b.router.Add(http.MethodPost, path, handler)
	b.router.Add(http.MethodDelete, path, handler)
	b.router.Add(http.MethodPut, path, handler)
	b.router.Add(http.MethodPatch, path, handler)
	b.router.Add(http.MethodOptions, path, handler)
	b.router.Add(http.MethodHead, path, handler)
	b.BindMiddleware()
}

// Router returns the router used by the blade.
func (b *Blade) Router() *Router {
	return &b.router
}

// Run starts your application with http.
func (b *Blade) Run(address string) error {
	Log.Debug("Listening and serving HTTP",
		zap.String("address", address))
	server := &http.Server{
		Addr:    address,
		Handler: b,
	}
	if err := server.ListenAndServe(); err != nil {
		return errors.Wrapf(err, "addrs: %v", address)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	return nil
}

// Run starts your application with https.
func (b *Blade) RunTLS(addr, certFile, keyFile string) error {
	Log.Debug("Listening and serving HTTPS",
		zap.String("address", addr))
	server := &http.Server{
		Addr:    addr,
		Handler: b,
	}
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		return errors.Wrapf(err, "tls: %s/%s:%s", addr, certFile, keyFile)

	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	return nil
}

// Run starts your application by given server and listener.
func (b *Blade) RunServer(server *http.Server, l net.Listener) error {
	Log.Debug("Listening and serving HTTP on listener what's bind with address@",
		zap.String("address", l.Addr().String()))
	server.Handler = b
	if err := server.Serve(l); err != nil {
		return errors.Wrapf(err, "listen server: %+v/%+v", server, l)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	return nil
}

// Use adds middleware to your middleware chain.
func (b *Blade) Use(middlewares ...Middleware) {
	b.middleware = append(b.middleware, middlewares...)
}

// newContext returns a new context from the pool.
func (b *Blade) newContext(req *http.Request, res http.ResponseWriter) *Context {
	c := b.contextPool.Get().(*Context)
	c.status = http.StatusOK
	c.request.req = req
	c.response.rw = res
	c.paramCount = 0
	c.modifierCount = 0
	return c
}

// ServeHTTP responds to the given request.
func (b *Blade) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	c := b.newContext(request, response)

	b.router.Lookup(request.Method, request.URL.Path, c)

	if c.handler == nil {
		if b.noFunc != nil {
			b.noFunc(c)
		} else {
			response.WriteHeader(http.StatusNotFound)
		}
		c.Close()
		return
	}

	c.handler(c)
	c.Close()
}

// Binding middleware
func (b *Blade) BindMiddleware() {
	b.router.bind(func(handler Handler) Handler {
		return handler.Bind(b.middleware...)
	})
}

// Whether to record request logs
func (b *Blade) EnableLogRequest() {
	b.Use(func(next Handler) Handler {
		return func(c *Context) error {
			Log = LogWithCtr(c)
			start := time.Now()
			st := start.Format("2006-01-02 15:04:05")
			path := c.request.Path()
			query := c.request.RawQuery()
			method := c.request.Method()

			var b []byte
			if method != "GET" && method != "OPTIONS" && method != "HEAD" {
				b, _ = c.request.RawDataSetBody()
				c.SetKey(BodyBytesKey, b)
			}

			err := next(c)

			end := time.Now()
			latency := end.Sub(start)
			if latency > time.Minute {
				latency = latency - latency%time.Second
			}
			Log.Info("Request record",
				zap.String("time", st),
				zap.Int("status", c.Status()),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.ByteString("body", b),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.request.req.UserAgent()),
				zap.Duration("latency", latency))

			Log = LogReleaseCtr(c)

			return err
		}
	})
}
