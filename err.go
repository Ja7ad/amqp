package amqp

import "errors"

var (
	ErrChannelIsNotCreated = errors.New("channel is not created, please check your connection or wait")
	ErrExchangeIsNil       = errors.New("exchange is nil")
	ErrQueueIsNil          = errors.New("queue is nil")
	ErrConsumerIsNil       = errors.New("consumer is nil")
	ErrRoutingKeyIsNil     = errors.New("routing key is nil")
)
