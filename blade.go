package hblade

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Blade struct {
	router       *Router[Handler]
	middleware   []Middleware //Global middleware
	contextPool  sync.Pool
	notFoundFn   func(*Context) //404
	errorHandler func(*Context, error)
}

// New creates a new blade.
func New() *Blade {
	b := &Blade{
		router:     &Router[Handler]{},
		notFoundFn: nil,
		errorHandler: func(c *Context, err error) {
			Log.Error("Error in handler",
				zap.Error(err),
				zap.String("path", c.request.Path()))
		},
	}

	// Context pool
	b.contextPool.New = func() any { return &Context{b: b} }

	return b
}

func (b *Blade) NotFoundFn(f func(*Context)) {
	b.notFoundFn = f
}

// Add registers a new handler for the given method and path.
func (b *Blade) Add(method, path string, handler Handler, m ...Middleware) {
	path = "/" + strings.Trim(path, "/")
	transform := b.transformMiddleware(m...)
	b.router.Add(method, path, transform(handler))
}

// Get registers your function to be called when the given GET path has been requested.
func (b *Blade) Get(path string, handler Handler, m ...Middleware) {
	b.Add(http.MethodGet, path, handler, m...)
}

// Post registers your function to be called when the given POST path has been requested.
func (b *Blade) Post(path string, handler Handler, m ...Middleware) {
	b.Add(http.MethodPost, path, handler, m...)
}

// Put registers your function to be called when the given PUT path has been requested.
func (b *Blade) Put(path string, handler Handler, m ...Middleware) {
	b.Add(http.MethodPut, path, handler, m...)
}

// Patch registers your function to be called when the given PATCH path has been requested.
func (b *Blade) Patch(path string, handler Handler, m ...Middleware) {
	b.Add(http.MethodPatch, path, handler, m...)
}

// Delete registers your function to be called when the given DELETE path has been requested.
func (b *Blade) Delete(path string, handler Handler, m ...Middleware) {
	b.Add(http.MethodDelete, path, handler, m...)
}

// Bind static directory
// h.Static("/static", "static/")
func (b *Blade) Static(path, bind string, m ...Middleware) {
	relativePath := "/" + strings.Trim(path, "/") + "/*file"
	handler := func(c *Context) error {
		return c.File(bind + c.Get("file"))
	}
	b.Get(relativePath, handler, m...)
}

// Router returns the router used by the blade.
func (b *Blade) Router() *Router[Handler] {
	return b.router
}

// Router group
func (b *Blade) Group(name string, m ...Middleware) *Group {
	name = strings.Trim(name, "/")
	g := &Group{app: b, name: name, middleware: m}
	return g
}

// ServeHTTP responds to the given request.
func (b *Blade) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	c := b.newContext(request, response)
	c.handler = b.router.Lookup(request.Method, request.URL.Path, c.addParameter)
	if c.handler == nil {
		if b.notFoundFn != nil {
			b.notFoundFn(c)
		} else {
			response.WriteHeader(http.StatusNotFound)
		}
		c.Close()
		return
	}

	err := c.handler(c)
	if err != nil {
		b.errorHandler(c, err)
	}
	c.Close()
}

// Run starts your application with http.
func (b *Blade) Run(addr string) error {
	Log.Debug("Listening and serving HTTP", zap.String("address", addr))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{Addr: addr, Handler: b}
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Log.Error("http listen error", zap.Error(err))
			errCh <- err
		}
	}()

	select {
	case sig := <-stop:
		Log.Info("Shutting down http server...", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "http server forced to shutdown")
		}
		Log.Info("Http server exited properly")
		return nil
	case err := <-errCh:
		return errors.Wrapf(err, "http server error, addr: %v", addr)
	}
}

// Run starts your application with https.
func (b *Blade) RunTLS(addr, certFile, keyFile string) error {
	Log.Debug("Listening and serving HTTPS", zap.String("address", addr))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{Addr: addr, Handler: b}
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			Log.Error("https listen error", zap.Error(err))
			errCh <- err
		}
	}()

	select {
	case sig := <-stop:
		Log.Info("Shutting down https server...", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "https server forced to shutdown")
		}
		Log.Info("Https server exited properly")
		return nil
	case err := <-errCh:
		return errors.Wrapf(err, "https server error, tls: %s/%s:%s", addr, certFile, keyFile)
	}
}

// Run starts your application by given server and listener.
func (b *Blade) RunServer(srv *http.Server, l net.Listener) error {
	Log.Debug("Listening and serving HTTP on listener what's bind with address@",
		zap.String("address", l.Addr().String()))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	srv.Handler = b
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			Log.Error("listen server error", zap.Error(err))
			errCh <- err
		}
	}()

	select {
	case sig := <-stop:
		Log.Info("Shutting down server...", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "https server forced to shutdown")
		}
		Log.Info("Server exited properly")
		return nil
	case err := <-errCh:
		return errors.Wrapf(err, "listen server: %+v/%+v", srv, l)
	}
}

// Use adds middleware to your middleware chain.
func (b *Blade) Use(m ...Middleware) {
	b.middleware = append(b.middleware, m...)
}

// newContext returns a new context from the pool.
func (b *Blade) newContext(req *http.Request, res http.ResponseWriter) *Context {
	c := b.contextPool.Get().(*Context)
	c.status = http.StatusOK
	c.request.req = req
	c.response.rw = res
	c.paramCount = 0
	return c
}

// Transform middleware
func (b *Blade) transformMiddleware(m ...Middleware) func(Handler) Handler {
	return func(handler Handler) Handler {
		mw := append(b.middleware, m...)
		return handler.Transform(mw...)
	}
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
			if method == http.MethodPost || method == http.MethodPut && method == http.MethodPatch {
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
