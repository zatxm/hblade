package hblade

type Handler func(*Context) error

func (h Handler) Bind(middleware ...Middleware) Handler {
	l := len(middleware) - 1
	for i := l; i >= 0; i-- {
		h = middleware[i](h)
	}

	return h
}
