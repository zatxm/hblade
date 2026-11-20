package hblade

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	defaultReadHeaderTimeout = 10 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultShutdownTimeout   = 5 * time.Second
)

type Blade struct {
	server       atomic.Pointer[http.Server] // read by Shutdown from other goroutines
	tlsCertFile  string
	tlsKeyFile   string
	router       *Router[Handler]
	middleware   []Middleware //Global middleware
	contextPool  sync.Pool
	compression  compression
	serverConfig func(*http.Server)
	notFoundFn   func(*Context) //404
	errorHandler func(*Context, error)
}

// New creates a new blade.
func New() *Blade {
	b := &Blade{
		router:     &Router[Handler]{},
		notFoundFn: nil,
		errorHandler: func(c *Context, err error) {
			Log().Error("Error in handler",
				zap.Error(err),
				zap.String("path", c.request.Path()))
		},
	}

	// Context pool
	b.contextPool.New = func() any { return &Context{b: b} }
	b.compression.init()

	return b
}

// ConfigServer customizes the *http.Server created by Run and Start, e.g. to change
// timeouts. It is applied after the defaults (ReadHeaderTimeout 10s, IdleTimeout 120s).
func (b *Blade) ConfigServer(fn func(*http.Server)) {
	b.serverConfig = fn
}

func (b *Blade) newServer(addr string) *http.Server {
	srv := &http.Server{
		Addr:              addr,
		Handler:           b,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}
	if b.serverConfig != nil {
		b.serverConfig(srv)
	}
	return srv
}

// serve starts srv in a goroutine. The returned channel receives exactly one
// value: nil after a graceful shutdown, otherwise the serve error.
func (b *Blade) serve(srv *http.Server, l net.Listener) <-chan error {
	b.server.Store(srv)
	errCh := make(chan error, 1)
	go func() {
		var err error
		tls := b.tlsCertFile != "" && b.tlsKeyFile != ""
		switch {
		case l == nil && tls:
			err = srv.ListenAndServeTLS(b.tlsCertFile, b.tlsKeyFile)
		case l == nil:
			err = srv.ListenAndServe()
		case tls:
			err = srv.ServeTLS(l, b.tlsCertFile, b.tlsKeyFile)
		default:
			err = srv.Serve(l)
		}
		if err == http.ErrServerClosed {
			err = nil
		} else if err != nil {
			Log().Error("http(s) listen error", zap.Error(err))
		}
		errCh <- err
	}()
	return errCh
}

// waitSignal blocks until the server stops or SIGINT/SIGTERM is received,
// in which case the server is gracefully shut down.
func (b *Blade) waitSignal(errCh <-chan error) (fromServer bool, err error) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case sig := <-stop:
		Log().Info("Shutting down http(s) server...", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
		return false, b.Shutdown(ctx)
	case err := <-errCh:
		return true, err
	}
}

func (b *Blade) NotFoundFn(f func(*Context)) {
	b.notFoundFn = f
}

func (b *Blade) TlsCertFile(f string) {
	b.tlsCertFile = f
}

func (b *Blade) TlsKeyFile(f string) {
	b.tlsKeyFile = f
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

// Run start your application with http(s)
func (b *Blade) Run(addr string) error {
	Log().Debug("Listening and serving HTTP(S)", zap.String("address", addr))

	fromServer, err := b.waitSignal(b.serve(b.newServer(addr), nil))
	if fromServer && err != nil {
		return errors.Wrapf(err, "http(s) server error, addr: %v", addr)
	}
	return err
}

// Run start your application by given server and listener
func (b *Blade) RunServer(srv *http.Server, l net.Listener) error {
	Log().Debug("Listening and serving HTTP(S) on listener what's bind with address",
		zap.String("address", l.Addr().String()))

	srv.Handler = b
	fromServer, err := b.waitSignal(b.serve(srv, l))
	if fromServer && err != nil {
		return errors.Wrapf(err, "listen server: %+v/%+v", srv, l)
	}
	return err
}

// Run start your application with http(s),you can control when to stop.
// It returns nil after Shutdown has been called.
func (b *Blade) Start(addr string) error {
	Log().Debug("Listening and serving HTTP(S)", zap.String("address", addr))

	if err := <-b.serve(b.newServer(addr), nil); err != nil {
		return errors.Wrapf(err, "http server error, addr: %v", addr)
	}
	return nil
}

func (b *Blade) Shutdown(ctx context.Context) error {
	srv := b.server.Load()
	if srv == nil {
		return nil
	}
	if err := srv.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "http(s) server forced to shutdown")
	}
	Log().Info("Http(s) server exited properly")
	return nil
}

// Use adds middleware to your middleware chain.
func (b *Blade) Use(m ...Middleware) {
	b.middleware = append(b.middleware, m...)
}

// newContext returns a new context from the pool.
func (b *Blade) newContext(req *http.Request, res http.ResponseWriter) *Context {
	// The remaining fields are reset by Close before a context is pooled.
	c := b.contextPool.Get().(*Context)
	c.status = http.StatusOK
	c.request.req = req
	c.response.rw = res
	return c
}

// Transform middleware
func (b *Blade) transformMiddleware(m ...Middleware) func(Handler) Handler {
	return func(handler Handler) Handler {
		return handler.Transform(slices.Concat(b.middleware, m)...)
	}
}

// Whether to record request logs
func (b *Blade) EnableLogRequest() {
	b.Use(func(next Handler) Handler {
		return func(c *Context) error {
			start := time.Now()
			req := c.request.req
			path := req.URL.Path
			query := req.URL.RawQuery
			method := req.Method

			var b []byte
			if hasRequestBody(method) {
				b, _ = c.request.RawDataSetBody()
				c.SetKey(BodyBytesKey, b)
			}

			err := next(c)

			latency := time.Since(start)
			if latency > time.Minute {
				latency = latency - latency%time.Second
			}
			Log().Info("Request record",
				zap.String("ReqId", uuid.NewString()),
				zap.String("time", start.Format(time.DateTime)),
				zap.Int("status", c.Status()),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.ByteString("body", b),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", req.UserAgent()),
				zap.Duration("latency", latency))

			return err
		}
	})
}
