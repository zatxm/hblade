package binding

import (
	"github.com/goccy/go-yaml"
	"github.com/valyala/fasthttp"
)

type yamlBinding struct{}

func (yamlBinding) Name() string {
	return "yaml"
}

func (yamlBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	return decodeYAML(req.PostBody(), obj)
}

func (yamlBinding) BindBody(body []byte, obj any) error {
	return decodeYAML(body, obj)
}

func decodeYAML(b []byte, obj any) error {
	if err := yaml.Unmarshal(b, obj); err != nil {
		return err
	}
	return validate(obj)
}
