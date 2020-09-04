package hblade

type Handler func(*Context) error

func (handler Handler) Bind(middleware ...Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	return handler
}
