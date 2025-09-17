package hblade

import (
	"net/http"
	"strings"
)

// router group
type Group struct {
	name       string
	app        *Blade
	middleware []Middleware
}

func (g *Group) Clear() {
	g = nil
}

// Add registers a new handler for the given method and path.
func (g *Group) Add(method, path string, handler Handler, m ...Middleware) {
	path = g.name + "/" + strings.TrimLeft(path, "/")
	mw := append(g.middleware, m...)
	g.app.Add(method, path, handler, mw...)
}

// Get registers your function to be called when the given GET path has been requested.
func (g *Group) Get(path string, handler Handler, m ...Middleware) {
	g.Add(http.MethodGet, path, handler, m...)
}

// Post registers your function to be called when the given POST path has been requested.
func (g *Group) Post(path string, handler Handler, m ...Middleware) {
	g.Add(http.MethodPost, path, handler, m...)
}

// Put registers your function to be called when the given PUT path has been requested.
func (g *Group) Put(path string, handler Handler, m ...Middleware) {
	g.Add(http.MethodPut, path, handler, m...)
}

// Patch registers your function to be called when the given PATCH path has been requested.
func (g *Group) Patch(path string, handler Handler, m ...Middleware) {
	g.Add(http.MethodPatch, path, handler, m...)
}

// Delete registers your function to be called when the given DELETE path has been requested.
func (g *Group) Delete(path string, handler Handler, m ...Middleware) {
	g.Add(http.MethodDelete, path, handler, m...)
}

// Bind static directory
func (g *Group) Static(path, bind string, m ...Middleware) {
	relativePath := strings.Trim(path, "/") + "/*file"
	handler := func(c *Context) error {
		return c.File(bind + c.Get("file"))
	}
	g.Get(relativePath, handler, m...)
}

// Use adds middleware to your middleware chain.
func (g *Group) Use(m ...Middleware) {
	g.middleware = append(g.middleware, m...)
}

// Children group
func (g *Group) Group(name string, m ...Middleware) *Group {
	name = g.name + "/" + strings.Trim(name, "/")
	mw := append(g.middleware, m...)
	cg := &Group{app: g.app, name: name, middleware: mw}
	return cg
}
