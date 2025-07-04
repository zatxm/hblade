package binding

import (
	"encoding/xml"

	"github.com/valyala/fasthttp"
)

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	return decodeXML(req.PostBody(), obj)
}

func (xmlBinding) BindBody(body []byte, obj any) error {
	return decodeXML(body, obj)
}

func decodeXML(b []byte, obj any) error {
	if err := xml.Unmarshal(b, obj); err != nil {
		return err
	}
	return validate(obj)
}
