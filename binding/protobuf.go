package binding

import (
	"errors"
	"net/http"

	"google.golang.org/protobuf/proto"

	"github.com/zatxm/hblade/v4/tools"
)

type protobufBinding struct{}

func (protobufBinding) Name() string {
	return "protobuf"
}

func (b protobufBinding) Bind(req *http.Request, obj any) error {
	buf, err := tools.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return b.BindBody(buf, obj)
}

func (protobufBinding) BindBody(body []byte, obj any) error {
	msg, ok := obj.(proto.Message)
	if !ok {
		return errors.New("obj is not ProtoMessage")
	}
	if err := proto.Unmarshal(body, msg); err != nil {
		return err
	}
	// Here it's same to return validate(obj), but until now we can't add
	// `binding:""` to the struct which automatically generate by gen-proto
	return nil
}
