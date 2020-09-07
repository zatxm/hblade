package hblade

import (
	"fmt"
	"net/http/httputil"
	"runtime"
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
					pl := fmt.Sprintf("http call panic: %s\n%v\n%s\n", string(rawReq), err, buf)
					Log.Fatal(pl)
					c.Error(500, err)
				}
			}()
			return next(c)
		}
	}
}
