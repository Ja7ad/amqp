package amqp

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMQOptions struct {
	ReconnectInterval time.Duration
	AMQPConfig        amqp.Config
}

type RabbitMQOptions func(*rabbitMQOptions)

func defaultRabbitMQOptions() *rabbitMQOptions {
	return &rabbitMQOptions{
		ReconnectInterval: 5 * time.Second,
		AMQPConfig:        amqp.Config{},
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
