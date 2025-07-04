package binding

import (
	"github.com/pelletier/go-toml/v2"
	"github.com/valyala/fasthttp"
)

type tomlBinding struct{}

func (tomlBinding) Name() string {
	return "toml"
}

func (tomlBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	return decodeToml(req.PostBody(), obj)
}

func (tomlBinding) BindBody(body []byte, obj any) error {
	return decodeToml(body, obj)
}

func decodeToml(b []byte, obj any) error {
	if err := toml.Unmarshal(b, obj); err != nil {
		return err
	}
	return validate(obj)
}
