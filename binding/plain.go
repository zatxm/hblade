package binding

import (
	"fmt"
	"reflect"

	"github.com/valyala/fasthttp"
	"github.com/zatxm/hblade/tools"
)

type plainBinding struct{}

func (plainBinding) Name() string {
	return "plain"
}

func (plainBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	return decodePlain(req.PostBody(), obj)
}

func (plainBinding) BindBody(body []byte, obj any) error {
	return decodePlain(body, obj)
}

func decodePlain(data []byte, obj any) error {
	if obj == nil {
		return nil
	}

	v := reflect.ValueOf(obj)

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() == reflect.String {
		v.SetString(tools.BytesToString(data))
		return nil
	}

	if _, ok := v.Interface().([]byte); ok {
		v.SetBytes(data)
		return nil
	}

	return fmt.Errorf("type (%T) unknown type", v)
}
