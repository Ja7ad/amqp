package errs

import "errors"

var (
	ErrExchangeIsNil         = errors.New("exchange is nil")
	ErrQueueIsNil            = errors.New("queue is nil")
	ErrConsumerIsNil         = errors.New("consumer is nil")
	ErrRoutingKeyIsNil       = errors.New("routing key is nil")
	ErrInvalidProtoMsgEncode = errors.New("nats: Invalid protobuf proto.Message object passed to encode")
	ErrInvalidProtoMsgDecode = errors.New("nats: Invalid protobuf proto.Message object passed to decode")
	ErrFailedEncode          = errors.New("failed to encode your body")
	ErrFailedDecode          = errors.New("failed to decode your body")
)
