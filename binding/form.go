package binding

import (
	"net/url"

	"github.com/valyala/fasthttp"
	"github.com/zatxm/hblade/tools"
)

const defaultMemory = 32 << 20

type (
	formBinding          struct{}
	formPostBinding      struct{}
	formMultipartBinding struct{}
)

func convertForm(ctx *fasthttp.RequestCtx) url.Values {
	form := url.Values{}
	ctx.QueryArgs().All()(func(key, val []byte) bool {
		form.Add(tools.BytesToString(key), tools.BytesToString(val))
		return true
	})
	ctx.PostArgs().All()(func(key, val []byte) bool {
		form.Add(tools.BytesToString(key), tools.BytesToString(val))
		return true
	})

	return form
}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	form := convertForm(req)
	if err := mapForm(obj, form); err != nil {
		return err
	}

	return validate(obj)
}

func (formPostBinding) Name() string {
	return "form-urlencoded"
}

func (formPostBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	form := url.Values{}
	req.PostArgs().All()(func(key, val []byte) bool {
		form.Add(tools.BytesToString(key), tools.BytesToString(val))
		return true
	})
	if err := mapForm(obj, form); err != nil {
		return err
	}
	return validate(obj)
}

func (formMultipartBinding) Name() string {
	return "multipart/form-data"
}

func (formMultipartBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	if err := mappingByPtr(obj, (*multipartRequest)(req), "form"); err != nil {
		return err
	}

	return validate(obj)
}
