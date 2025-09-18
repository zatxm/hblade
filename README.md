# hblade(刀锋)框架

golang简易web http框架，用于快速构建web及api服务

## 使用方法

1. golang版本1.25+
1. go get -u github.com/zatxm/hblade/v3

```golang
package main

import (
    "fmt"
    "net/http"

    "github.com/zatxm/hblade/v3"
)

func main() {
    app := hblade.New()

    // 通用方法Get /a
    app.Add(http.MethodGet, "/a", func(c *hblade.Context) error {
        return c.String("OK")
    })

    // Get快捷方法Get b
    app.Get("/b", func(c *hblade.Context) error {
        return c.String("OK")
    })

    // Post快捷方法Post /b，返回json数据
    app.Post("/b", func(c *hblade.Context) error {
        return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
    })

    // 包含参数Get /c/a/b
    app.Get("/c/:p1/:p2", func(c *hblade.Context) error {
        fmt.Println(c.Get("p1"))
        fmt.Println(c.Get("p2"))
        return c.String(c.Path())
    })

    // 通配符Get /d/a/b.jpg
    app.Get("/d/*file", func(c *hblade.Context) error {
        // 访问/d/aaaa，输出aaaa
        // 访问/d/s/a.jpg，输出s/a.jpg
        fmt.Println(c.Get("file"))
        return c.String(c.Path())
    })

    // 静态文件Get /e/1.png,路由映射到statics目录下
    app.Static("/e", "statics/")

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
        // 路由Post /v1/a
        v1.Post("/a", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        })
        // 路由Post /v1/b
        v1.Post("/b", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        })
        // 路由Get /v1/c
        v1.Get("/c", func(c *hblade.Context) error {
            return c.String("Ok")
        })
    }

    // 路由组v2
    {
        v2 := app.Group("/v2")
        // 路由Post /v2/a
        v2.Post("/a", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        })
        // 路由Post /v2/b
        v2.Post("/b", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        })
        // 路由Get /v2/c
        v2.Get("/c", func(c *hblade.Context) error {
            return c.String("Ok")
        })

        // v2嵌套路由组vc
        vc := v2.Group("/vc")
        // 路由Post /v2/vc/a
        vc.Post("/a", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        })
        // 路由Post /v2/vc/b
        vc.Post("/b", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        })
        // 路由Get /v2/vc/a
        vc.Get("/a", func(c *hblade.Context) error {
            return c.String("Ok")
        })
    }

    err := app.Run(":8881")
    if err != nil {
        panic(err)
    }
}

```

## Middleware中间件

路由组也支持Middleware中间件

```golang
package main

import (
    "fmt"
    "net/http"

    "github.com/zatxm/hblade/v3"
)

var (
    // 中间件m1
    m1 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m1 start")
            err := next(c)
            fmt.Println("m1 end")
            return err
        }
    }
    // 中间件m2
    m2 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m2")
            return next(c)
        }
    }
    // 中间件m3
    m3 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m3")
            return next(c)
        }
    }
    // 中间件m4
    m4 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m4")
            return next(c)
        }
    }
    // 中间件m5
    m5 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m5")
            return next(c)
        }
    }
    // 中间件m6
    m6 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m6")
            return next(c)
        }
    }
    // 中间件m7
    m7 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m7")
            return next(c)
        }
    }
    // 中间件m8
    m8 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m8")
            return next(c)
        }
    }
    // 中间件m9
    m9 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m9")
            return next(c)
        }
    }
    // 中间件m10
    m10 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m10")
            return next(c)
        }
    }
    // 中间件m11
    m11 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m11")
            return next(c)
        }
    }
    // 中间件m12
    m12 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m12")
            return next(c)
        }
    }
    // 中间件m13
    m13 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m13")
            return next(c)
        }
    }
    // 中间件m14
    m14 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m14")
            return next(c)
        }
    }
    // 中间件m15
    m15 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m15")
            return next(c)
        }
    }
    // 中间件m16
    m16 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m16")
            return next(c)
        }
    }
    // 中间件m17
    m17 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m17")
            return next(c)
        }
    }
    // 中间件m18
    m18 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m18")
            return next(c)
        }
    }
    // 中间件m19
    m19 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m19")
            return next(c)
        }
    }
    // 中间件m20
    m20 = func(next hblade.Handler) hblade.Handler {
        return func(c *hblade.Context) error {
            fmt.Println("m20")
            return next(c)
        }
    }
)

func main() {
    app := hblade.New()

    // 无中间件Get /a
    app.Post("/a", func(c *hblade.Context) error {
        return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
    })

    // 全局中间件，中间件控制中间件以下代码路由，因此上面路由不受中间件控制
    app.Use(m1, m2)

    // Get /b，生效中间件为m1、m2
    app.Get("/b", func(c *hblade.Context) error {
        return c.String(c.Path())
    })

    // Get /c，生效中间件为m1、m2、m3、m4(m3和m4该路由独有)
    app.Get("/c", func(c *hblade.Context) error {
        return c.String(c.Path())
    }, m3, m4)

    // Get /d，生效中间件为m1、m2
    app.Get("/d", func(c *hblade.Context) error {
        return c.String(c.Path())
    })

    // 路由组: v1
    {
        // 中间件m5为组内共用
        v1 := app.Group("/v1", m5)
        // 路由Post /v1/a，生效中间件为m1、m2、m5、m6
        v1.Post("/a", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        }, m6)
        v1.Use(m7)
        // 路由Post /v1/b，生效中间件为m1、m2、m5、m7、m8、m9
        v1.Post("/b", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        }, m8, m9)
        // 路由Get /v1/c，生效中间件为m1、m2、m5、m7
        v1.Get("/c", func(c *hblade.Context) error {
            return c.String("Ok")
        })
    }

    // 路由组: v2
    {
        // 中间件m10为组内共用
        v2 := app.Group("/v2", m10)
        // 路由Post /v2/a，生效中间件为m1、m2、m10、m11
        v2.Post("/a", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        }, m11)
        v2.Use(m12)
        // 路由Post /v2/b，生效中间件为m1、m2、m10、m12、m13、m14
        v2.Post("/b", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        }, m13, m14)
        // 路由Get /v2/c，生效中间件为m1、m2、m10、m12
        v2.Get("/c", func(c *hblade.Context) error {
            return c.String("Ok")
        })

        // v2嵌套路由组vc
        // 嵌套路由组拥有上级全部中间件
        // 中间件m15为嵌套组内共用
        vc := v2.Group("/vc", m15)
        // 路由Post /v2/vc/a，生效中间件为m1、m2、m10、m12、m15、m16
        vc.Post("/a", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        }, m16)
        vc.Use(m17)
        // 路由Post /v2/vc/b，生效中间件为m1、m2、m10、m12、m15、m17、m18、m19
        vc.Post("/b", func(c *hblade.Context) error {
            return c.JSONAndStatus(http.StatusOK, map[string]string{"Data": c.Path()})
        }, m18, m19)
        // 路由Get /v2/vc/a，生效中间件为m1、m2、m10、m12、m15、m17
        vc.Get("/a", func(c *hblade.Context) error {
            return c.String("Ok")
        })
    }

    // Get /e，生效中间件为m1、m2
    app.Get("/e", func(c *hblade.Context) error {
        return c.String(c.Path())
    })

    app.Use(m20)

    // Get /f，生效中间件为m1、m2、m20
    app.Get("/f", func(c *hblade.Context) error {
        return c.String(c.Path())
    })

    err := app.Run(":8881")
    if err != nil {
        panic(err)
    }
}

```
