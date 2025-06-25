package hblade

type Handler func(*Context) error

func (handler Handler) Bind(middleware ...Middleware) Handler {
	l := len(middleware) - 1
	for i := l; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	return handler
}
