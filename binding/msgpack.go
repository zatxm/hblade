package binding

import (
	"bytes"
	"io"

	"github.com/ugorji/go/codec"
	"github.com/valyala/fasthttp"
)

type msgpackBinding struct{}

func (msgpackBinding) Name() string {
	return "msgpack"
}

func (msgpackBinding) Bind(req *fasthttp.RequestCtx, obj any) error {
	return decodeMsgPack(bytes.NewReader(req.PostBody()), obj)
}

func (msgpackBinding) BindBody(body []byte, obj any) error {
	return decodeMsgPack(bytes.NewReader(body), obj)
}

func decodeMsgPack(r io.Reader, obj any) error {
	cdc := new(codec.MsgpackHandle)
	if err := codec.NewDecoder(r, cdc).Decode(&obj); err != nil {
		return err
	}
	return validate(obj)
}
