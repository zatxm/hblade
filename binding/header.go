package binding

import (
	"net/textproto"
	"net/url"
	"reflect"

	"github.com/valyala/fasthttp"
	"github.com/zatxm/hblade/tools"
)

type headerBinding struct{}

func (headerBinding) Name() string {
	return "header"
}

func (headerBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	form := url.Values{}
	req.Request.Header.All()(func(key, val []byte) bool {
		form.Add(tools.BytesToString(key), tools.BytesToString(val))
		return true
	})
	if err := mapHeader(obj, form); err != nil {
		return err
	}

	return validate(obj)
}

func mapHeader(ptr any, h map[string][]string) error {
	return mappingByPtr(ptr, headerSource(h), "header")
}

type headerSource map[string][]string

var _ setter = headerSource(nil)

func (hs headerSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (bool, error) {
	return setByForm(value, field, hs, textproto.CanonicalMIMEHeaderKey(tagValue), opt)
}
