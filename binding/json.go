package binding

import (
	"errors"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	body := req.PostBody()
	if req == nil || len(body) == 0 {
		return errors.New("invalid request")
	}
	return decodeJSON(body, obj)
}

func (jsonBinding) BindBody(body []byte, obj any) error {
	return decodeJSON(body, obj)
}

func decodeJSON(b []byte, obj any) error {
	if err := Json.Unmarshal(b, obj); err != nil {
		return err
	}
	return validate(obj)
}
