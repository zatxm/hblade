package hblade

import (
	"net/http/httputil"
	"runtime"

	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			defer func() {
				var rawReq []byte
				if err := recover(); err != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					req := c.request.req
					if req != nil {
						rawReq, _ = httputil.DumpRequest(req, false)
					}
					Log().Fatal("http call panic",
						zap.ByteString("rawReq", rawReq),
						zap.Any("error", err),
						zap.ByteString("buf", buf))
					c.Error(500, err)
				}
			}()
			return next(c)
		}
	}
}
