package binding

import (
	"net/url"

	"github.com/valyala/fasthttp"
	"github.com/zatxm/hblade/v2/tools"
)

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	form := url.Values{}
	req.QueryArgs().All()(func(key, val []byte) bool {
		form.Add(tools.BytesToString(key), tools.BytesToString(val))
		return true
	})
	if err := mapForm(obj, form); err != nil {
		return err
	}
	return validate(obj)
}
