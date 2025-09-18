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

## 路由组

```golang
package main

import (
    "net/http"

    "github.com/zatxm/hblade/v3"
)

func main() {
    app := hblade.New()

    // 路由组v1
    {
        v1 := app.Group("/v1")
        // 路由/v1/login
        v1.Post("/login", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Code": "Ok", "Data": c.Path()})
        })
        // 路由/v1/submit
        v1.Post("/submit", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Code": "Ok", "Data": c.Path()})
        })
        // 路由/v1/read
        v1.Get("/read", func(c *hblade.Context) error {
            return c.String("Ok")
        })
    }

    // 路由组v2
    {
        v2 := app.Group("/v2")
        // 路由/v2/login
        v2.Post("/login", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Code": "Ok", "Data": c.Path()})
        })
        // 路由/v2/submit
        v2.Post("/submit", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Code": "Ok", "Data": c.Path()})
        })
        // 路由/v2/read
        v2.Get("/read", func(c *hblade.Context) error {
            return c.String("Ok")
        })

        // v2嵌套路由组vc
        vc := v2.Group("/vc")
        // 路由/v2/vc/login
        vc.Post("/login", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Code": "Ok", "Data": c.Path()})
        })
        // 路由/v2/vc/submit
        vc.Post("/submit", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Code": "Ok", "Data": c.Path()})
        })
        // 路由/v2/vc/read
        vc.Get("/read", func(c *hblade.Context) error {
            return c.String("Ok")
        })
    }

    err := app.Run(":8881")
    if err != nil {
        panic(err)
    }
}

```
