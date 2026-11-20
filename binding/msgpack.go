package binding

import (
	"net/http"

	"github.com/ugorji/go/codec"
)

// A handle caches type metadata and is safe for concurrent use, so share one.
var msgpackHandle = new(codec.MsgpackHandle)

type msgpackBinding struct{}

func (msgpackBinding) Name() string {
	return "msgpack"
}

func (msgpackBinding) Bind(req *http.Request, obj any) error {
	if err := codec.NewDecoder(req.Body, msgpackHandle).Decode(&obj); err != nil {
		return err
	}
	return validate(obj)
}

func (msgpackBinding) BindBody(body []byte, obj any) error {
	if err := codec.NewDecoderBytes(body, msgpackHandle).Decode(&obj); err != nil {
		return err
	}
	return validate(obj)
}
