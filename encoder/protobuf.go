package encoder

import (
	"github.com/Ja7ad/amqp/errs"
	"google.golang.org/protobuf/proto"
)

// ProtoBufEncoder is a protobuf implementation for EncodedConn
// This encoder will use the builtin protobuf lib to Marshal
// and Unmarshal structs.
type ProtoBufEncoder struct{}

func (p *ProtoBufEncoder) Encode(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	i, found := v.(proto.Message)
	if !found {
		return nil, errs.ErrInvalidProtoMsgEncode
	}

	b, err := proto.Marshal(i)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (p *ProtoBufEncoder) Decode(data []byte, vPtr any) error {
	if _, ok := vPtr.(*any); ok {
		return nil
	}
	i, found := vPtr.(proto.Message)
	if !found {
		return errs.ErrInvalidProtoMsgDecode
	}

	return proto.Unmarshal(data, i)
}
