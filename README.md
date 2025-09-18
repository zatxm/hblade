# hblade框架(刀锋)

golang简易web http框架，用于快速构建web及api服务

## 使用方法

golang 1.25+  
go get -u github.com/zatxm/hblade/v3

```golang
package main

import (
    "fmt"
    "net/http"

    "github.com/zatxm/hblade/v3"
)

func main() {
    app := hblade.New()

    // 通用方法
    app.Add(http.MethodGet, "/ping", func(c *hblade.Context) error {
        return c.String("OK")
    })

    // Get快捷方法
    app.Get("/ping/a", func(c *hblade.Context) error {
        return c.String("OK")
    })

    // Post快捷方法，返回json数据
    app.Post("/login", func(c *hblade.Context) error {
        return c.JSONAndStatus(http.StatusOK, map[string]string{"Code": "Ok", "Data": c.Path()})
    })

    // 包含参数
    app.Get("/pa/:p1/:p2", func(c *hblade.Context) error {
        fmt.Println(c.Get("p1"))
        fmt.Println(c.Get("p2"))
        return c.String(c.Path())
    })

    // 通配符
    app.Get("/common/*file", func(c *hblade.Context) error {
        // 访问/common/aaaa，输出aaaa
        // 访问/common/s/a.jpg，输出s/a.jpg
        fmt.Println(c.Get("file"))
        return c.String(c.Path())
    })

    // 静态文件,路由映射到statics目录下
    app.Static("/static", "statics/")

    err := app.Run(":8881")
    if err != nil {
        panic(err)
    }
}

```
