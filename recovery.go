package hblade

import (
	"runtime"

	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			defer func() {
				if err := recover(); err != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					Log.Fatal("http call panic",
						zap.String("rawReq", c.ctx.Request.Header.String()),
						zap.Any("error", err),
						zap.ByteString("buf", buf))
					c.Error(500, err)
				}
			}()
			return next(c)
		}
	}
}
