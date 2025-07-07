package hblade

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"github.com/zatxm/hblade/tools"
	"go.uber.org/zap"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

type Blade struct {
	router                    *Router[Handler]
	middleware                []Middleware
	contextPool               sync.Pool
	fasthttpKeepHijackedConns bool           //Used in websocket, autonomously control conn closing
	notFoundFn                func(*Context) //404
	errorHandler              func(*Context, error)
}

// New creates a new blade.
func New() *Blade {
	b := &Blade{
		router:     &Router[Handler]{},
		notFoundFn: nil,
		errorHandler: func(c *Context, err error) {
			Log.Error("Error in handler",
				zap.Error(err),
				zap.String("path", c.Path()))
		},
	}

	// Context pool
	b.contextPool.New = func() any { return &Context{b: b} }

	return b
}

func (b *Blade) NotFoundFn(f func(*Context)) {
	b.notFoundFn = f
}

func (b *Blade) KeepHijackedConns(keep bool) {
	b.fasthttpKeepHijackedConns = keep
}

// Get registers your function to be called when the given GET path has been requested.
func (b *Blade) Get(path string, handler Handler) {
	b.router.Add("GET", path, handler)
	b.BindMiddleware()
}

// Post registers your function to be called when the given POST path has been requested.
func (b *Blade) Post(path string, handler Handler) {
	b.router.Add("POST", path, handler)
	b.BindMiddleware()
}

// Delete registers your function to be called when the given DELETE path has been requested.
func (b *Blade) Delete(path string, handler Handler) {
	b.router.Add("DELETE", path, handler)
	b.BindMiddleware()
}

// Put registers your function to be called when the given PUT path has been requested.
func (b *Blade) Put(path string, handler Handler) {
	b.router.Add("PUT", path, handler)
	b.BindMiddleware()
}

// Patch registers your function to be called when the given PATCH path has been requested.
func (b *Blade) Patch(path string, handler Handler) {
	b.router.Add("PATCH", path, handler)
	b.BindMiddleware()
}

// Options registers your function to be called when the given OPTIONS path has been requested.
func (b *Blade) Options(path string, handler Handler) {
	b.router.Add("OPTIONS", path, handler)
	b.BindMiddleware()
}

// Head registers your function to be called when the given HEAD path has been requested.
func (b *Blade) Head(path string, handler Handler) {
	b.router.Add("HEAD", path, handler)
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
	b.router.Add("GET", path, handler)
	b.router.Add("POST", path, handler)
	b.router.Add("DELETE", path, handler)
	b.router.Add("PUT", path, handler)
	b.router.Add("PATCH", path, handler)
	b.router.Add("OPTIONS", path, handler)
	b.router.Add("HEAD", path, handler)
	b.BindMiddleware()
}

// Router returns the router used by the blade.
func (b *Blade) Router() *Router[Handler] {
	return b.router
}

func (b *Blade) handle() func(*fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		c := b.newContext(ctx)
		c.handler = b.router.Lookup(tools.BytesToString(ctx.Method()), tools.BytesToString(ctx.Path()), c.addParameter)
		if c.handler == nil {
			if b.notFoundFn != nil {
				b.notFoundFn(c)
			} else {
				ctx.SetStatusCode(404)
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
}

// Run starts your application with http.
func (b *Blade) Run(addr string) error {
	Log.Debug("Listening and serving HTTP", zap.String("address", addr))

	s := &fasthttp.Server{
		Handler:           b.handle(),
		KeepHijackedConns: b.fasthttpKeepHijackedConns,
	}
	if err := s.ListenAndServe(addr); err != nil {
		return errors.Wrapf(err, "addrs: %v", addr)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	return nil
}

// Run starts your application with https.
func (b *Blade) RunTLS(addr, certFile, keyFile string) error {
	Log.Debug("Listening and serving HTTPS", zap.String("address", addr))

	s := &fasthttp.Server{
		Handler:           b.handle(),
		KeepHijackedConns: b.fasthttpKeepHijackedConns,
	}
	if err := s.ListenAndServeTLS(addr, certFile, keyFile); err != nil {
		return errors.Wrapf(err, "tls: %s/%s:%s", addr, certFile, keyFile)
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
func (b *Blade) newContext(ctx *fasthttp.RequestCtx) *Context {
	c := b.contextPool.Get().(*Context)
	c.status = 200
	c.ctx = ctx
	c.paramCount = 0
	return c
}

// Binding middleware
func (b *Blade) BindMiddleware() {
	b.router.Bind(func(handler Handler) Handler {
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
			path := c.ctx.Path()
			method := tools.BytesToString(c.ctx.Method())

			var b []byte
			if method != "GET" && method != "OPTIONS" && method != "HEAD" {
				b = c.ctx.PostBody()
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
				zap.ByteString("path", path),
				zap.ByteString("query", c.ctx.URI().QueryString()),
				zap.ByteString("body", b),
				zap.String("ip", c.ClientIP()),
				zap.ByteString("user-agent", c.ctx.Request.Header.UserAgent()),
				zap.Duration("latency", latency))

			Log = LogReleaseCtr(c)

			return err
		}
	})
}
