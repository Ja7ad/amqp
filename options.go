package amqp

import (
	"time"

	"github.com/Ja7ad/amqp/types"

	"github.com/Ja7ad/amqp/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMQOptions struct {
	ReconnectInterval time.Duration
	AMQPConfig        amqp.Config
	Logger            logger.Logger
	EncType           types.EncodeType
}

type RabbitMQOptions func(*rabbitMQOptions)

func defaultRabbitMQOptions() *rabbitMQOptions {
	return &rabbitMQOptions{
		ReconnectInterval: 5 * time.Second,
		AMQPConfig:        amqp.Config{},
		EncType:           types.JSON,
		Logger: logger.New(logger.CONSOLE_HANDLER, logger.Options{
			EnableCaller: true,
			SkipCaller:   3,
		}),
	}
}

func ReconnectDelay(delay time.Duration) RabbitMQOptions {
	return func(cfg *rabbitMQOptions) {
		cfg.ReconnectInterval = delay
	}
}

func WithCustomAMQPConfig(config amqp.Config) RabbitMQOptions {
	return func(cfg *rabbitMQOptions) {
		cfg.AMQPConfig = config
	}
}

func WithCustomLogger(logger logger.Logger) RabbitMQOptions {
	return func(cfg *rabbitMQOptions) {
		cfg.Logger = logger
	}
}

// WithCustomEncoder change default encoder to another encoder (JSON, GOB, ProtoBuf)
//
// Note: if you change default encoder form publisher or consumer, both match encoder (encode and decode)
func WithCustomEncoder(encType types.EncodeType) RabbitMQOptions {
	return func(o *rabbitMQOptions) {
		o.EncType = encType
	}
}

type consumerOptions struct {
	Concurrency int
	QOSPrefetch int
	QOSGlobal   bool
}

type ConsumerOptions func(options *consumerOptions)

func defaultConsumerOptions() *consumerOptions {
	return &consumerOptions{
		Concurrency: 1,
		QOSPrefetch: 10,
		QOSGlobal:   false,
	}
}

// WithConcurrentConsumer many goroutines will be spawned to run the provided handler on messages
func WithConcurrentConsumer(concurrent int) ConsumerOptions {
	return func(opt *consumerOptions) {
		opt.Concurrency = concurrent
	}
}

// WithCustomQOSPrefetch which means that
// many messages will be fetched from the server in advance to help with throughput.
// This doesn't affect the handler, messages are still processed one at a time.
func WithCustomQOSPrefetch(prefetch int) ConsumerOptions {
	return func(opt *consumerOptions) {
		opt.QOSPrefetch = prefetch
	}
}

// EnableQOSGlobal which means
// these QOS settings apply to ALL existing and future
// consumers on all channels on the same connection
func EnableQOSGlobal(enable bool) ConsumerOptions {
	return func(opt *consumerOptions) {
		opt.QOSGlobal = enable
	}
}

type publisherOptions struct {
	AutoMessageID bool
	AutoTimeStamp bool
}

type PublisherOption func(options *publisherOptions)

func defaultPublisherOption() *publisherOptions {
	return &publisherOptions{
		AutoTimeStamp: false,
		AutoMessageID: false,
	}
}

// WithAutoMessageID set uuid for message on publishing
func WithAutoMessageID() PublisherOption {
	return func(opt *publisherOptions) {
		opt.AutoMessageID = true
	}
}

// WithAutoTimestamp set timestamp in message on publishing
func WithAutoTimestamp() PublisherOption {
	return func(opt *publisherOptions) {
		opt.AutoTimeStamp = true
	}
}
