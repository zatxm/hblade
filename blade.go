package hblade

import (
	"compress/gzip"
	stdContext "context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

type Blade struct {
	signAutoRun    bool //ctrl+c cannot be closed, it will start automatically
	gzip           bool
	router         Router
	rewrite        []func(RewriteContext)
	middleware     []Middleware
	onStart        []func()
	onShutdown     []func()
	onError        []func(*Context, error)
	stop           chan os.Signal
	contextPool    sync.Pool
	gzipWriterPool sync.Pool
	server         atomic.Value
}

// New creates a new blade.
func New() *Blade {
	b := &Blade{
		signAutoRun: false,
		gzip:        false,
		stop:        make(chan os.Signal, 1),
	}

	// Context pool
	b.contextPool.New = func() interface{} {
		return &Context{b: b}
	}

	return b
}
func (b *Blade) SetSignAutoRun(auto bool) {
	b.signAutoRun = auto
}

// Set whether to enable Gzip
func (b *Blade) SetGzip(gzip bool) {
	b.gzip = gzip
}

// Get Gzip
func (b *Blade) Gzip() bool {
	return b.gzip
}

// Get registers your function to be called when the given GET path has been requested.
func (b *Blade) Get(path string, handler Handler) {
	b.router.Add(http.MethodGet, path, handler)
}

// Post registers your function to be called when the given POST path has been requested.
func (b *Blade) Post(path string, handler Handler) {
	b.router.Add(http.MethodPost, path, handler)
}

// Delete registers your function to be called when the given DELETE path has been requested.
func (b *Blade) Delete(path string, handler Handler) {
	b.router.Add(http.MethodDelete, path, handler)
}

// Put registers your function to be called when the given PUT path has been requested.
func (b *Blade) Put(path string, handler Handler) {
	b.router.Add(http.MethodPut, path, handler)
}

// Patch registers your function to be called when the given PATCH path has been requested.
func (b *Blade) Patch(path string, handler Handler) {
	b.router.Add(http.MethodPatch, path, handler)
}

// Options registers your function to be called when the given OPTIONS path has been requested.
func (b *Blade) Options(path string, handler Handler) {
	b.router.Add(http.MethodOptions, path, handler)
}

// Head registers your function to be called when the given HEAD path has been requested.
func (b *Blade) Head(path string, handler Handler) {
	b.router.Add(http.MethodHead, path, handler)
}

// Any registers your function to be called with any http method.
func (b *Blade) Any(path string, handler Handler) {
	b.Get(path, handler)
	b.Post(path, handler)
	b.Delete(path, handler)
	b.Put(path, handler)
	b.Patch(path, handler)
	b.Options(path, handler)
	b.Head(path, handler)
}

// Router returns the router used by the blade.
func (b *Blade) Router() *Router {
	return &b.router
}

func (b *Blade) initRun() {
	if b.signAutoRun {
		// Receive signals
		signal.Notify(b.stop, os.Interrupt, syscall.SIGTERM)
	}
	b.BindMiddleware()
}

func (b *Blade) endRun() error {
	for _, callback := range b.onStart {
		callback()
	}

	if b.signAutoRun {
		b.wait()
		return b.Shutdown()
	}
	return nil
}

// Run starts your application with http.
func (b *Blade) Run(addr ...string) error {
	b.initRun()

	address := resolveAddress(addr)
	Log.Debug("Listening and serving HTTP on ", address)
	server := &http.Server{
		Addr:    address,
		Handler: b,
	}
	b.server.Store(server)
	if err := server.ListenAndServe(); err != nil {
		return errors.Wrapf(err, "addrs: %v", addr)
	}

	return b.endRun()
}

// Run starts your application with https.
func (b *Blade) RunTLS(addr, certFile, keyFile string) error {
	b.initRun()

	Log.Debug("Listening and serving HTTPS on ", addr)
	server := &http.Server{
		Addr:    addr,
		Handler: b,
	}
	b.server.Store(server)
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		return errors.Wrapf(err, "tls: %s/%s:%s", addr, certFile, keyFile)

	}

	for _, callback := range b.onStart {
		callback()
	}

	return b.endRun()
}

// Run starts your application with Unix.
func (b *Blade) RunUnix(file string) error {
	b.initRun()

	Log.Debug("Listening and serving HTTP on unix:/", file)
	os.Remove(file)
	listener, err := net.Listen("unix", file)
	if err != nil {
		return errors.Wrapf(err, "unix: %s", file)
	}
	defer listener.Close()
	server := &http.Server{Handler: b}
	b.server.Store(server)
	if err = server.Serve(listener); err != nil {
		return errors.Wrapf(err, "unix: %s", file)
	}

	return b.endRun()
}

// Run starts your application by given server and listener.
func (b *Blade) RunServer(server *http.Server, l net.Listener) error {
	b.initRun()

	Log.Debug("Listening and serving HTTP on listener what's bind with address@", l.Addr().String())
	server.Handler = b
	b.server.Store(server)
	if err := server.Serve(l); err != nil {
		return errors.Wrapf(err, "listen server: %+v/%+v", server, l)
	}

	return b.endRun()
}

// Use adds middleware to your middleware chain.
func (b *Blade) Use(middlewares ...Middleware) {
	b.middleware = append(b.middleware, middlewares...)
}

// wait will make the process wait until it is killed.
func (b *Blade) wait() {
	<-b.stop
}

// Shutdown will gracefully shut down the server.
func (b *Blade) Shutdown() error {
	server := b.Server()
	if server == nil {
		return errors.New("no server")
	}

	err := shutdown(server)

	for _, callback := range b.onShutdown {
		callback()
	}
	return err
}

// Server is used to load stored http server.
func (b *Blade) Server() *http.Server {
	server, ok := b.server.Load().(*http.Server)
	if !ok {
		return nil
	}
	return server
}

// OnStart registers a callback to be executed on server start.
func (b *Blade) OnStart(callback func()) {
	b.onStart = append(b.onStart, callback)
}

// OnEnd registers a callback to be executed on server shutdown.
func (b *Blade) OnEnd(callback func()) {
	b.onShutdown = append(b.onShutdown, callback)
}

// OnError registers a callback to be executed on server errors.
func (b *Blade) OnError(callback func(*Context, error)) {
	b.onError = append(b.onError, callback)
}

// Rewrite adds a URL path rewrite function.
func (b *Blade) Rewrite(rewrite func(RewriteContext)) {
	b.rewrite = append(b.rewrite, rewrite)
}

// newContext returns a new context from the pool.
func (b *Blade) newContext(req *http.Request, res http.ResponseWriter) *Context {
	c := b.contextPool.Get().(*Context)
	c.status = http.StatusOK
	c.request.inner = req
	c.response.inner = res
	c.paramCount = 0
	c.modifierCount = 0
	return c
}

// ServeHTTP responds to the given request.
func (b *Blade) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	c := b.newContext(request, response)

	for _, rewrite := range b.rewrite {
		rewrite(c)
	}

	b.router.Lookup(request.Method, request.URL.Path, c)

	if c.handler == nil {
		response.WriteHeader(http.StatusNotFound)
		c.Close()
		return
	}

	err := c.handler(c)

	if err != nil {
		for _, callback := range b.onError {
			callback(c, err)
		}
	}

	c.Close()
}

// acquireGZipWriter will return a clean gzip writer from the pool.
func (b *Blade) acquireGZipWriter(response io.Writer) *gzip.Writer {
	var writer *gzip.Writer
	obj := b.gzipWriterPool.Get()

	if obj == nil {
		writer, _ = gzip.NewWriterLevel(response, gzip.BestCompression)
		return writer
	}

	writer = obj.(*gzip.Writer)
	writer.Reset(response)
	return writer
}

// 绑定中间件
func (b *Blade) BindMiddleware() {
	b.router.bind(func(handler Handler) Handler {
		return handler.Bind(b.middleware...)
	})
}

// shutdown will gracefully shut down the server.
func shutdown(server *http.Server) error {
	if server == nil {
		return errors.New("no server")
	}

	// Add a timeout to the server shutdown
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 250*time.Millisecond)
	defer cancel()

	return errors.WithStack(server.Shutdown(ctx))
}
