package hblade

type Handler func(*Context) error

// Transform middleware
func (h Handler) Transform(m ...Middleware) Handler {
	l := len(m) - 1
	if l < 0 {
		return h
	}
	for i := l; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
