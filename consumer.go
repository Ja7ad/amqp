package amqp

import (
	"fmt"
	"sync"

	"github.com/Ja7ad/amqp/errs"

	"github.com/Ja7ad/amqp/logger"
	"github.com/Ja7ad/amqp/types"
	rabbit "github.com/rabbitmq/amqp091-go"
)

type consumer struct {
	chanManager                *channel
	reconnectErrCh             <-chan error
	closeConnectionToManagerCh chan<- struct{}

	isClosedMu *sync.RWMutex
	isClosed   bool

	options *consumerOptions
	logger  logger.Logger

	exchange    *types.Exchange
	queue       *types.Queue
	consumer    *types.Consumer
	routingKeys []*types.RoutingKey
	handler     types.ConsumerHandler
	enc         types.Encoder
}

type Consumer interface {
	// Start consumer for consume messages
	Start() error
	// Close close consumer
	Close()
}

func (r *AMQP) Consumer(
	exchange *types.Exchange,
	queue *types.Queue,
	consume *types.Consumer,
	routingKeys []*types.RoutingKey,
	messageHandler types.ConsumerHandler,
	options ...ConsumerOptions,
) (Consumer, error) {
	defaultOptions := defaultConsumerOptions()
	for _, opt := range options {
		opt(defaultOptions)
	}

	if exchange == nil {
		return nil, errs.ErrExchangeIsNil
	}

	if queue == nil {
		return nil, errs.ErrQueueIsNil
	}

	if consume == nil {
		return nil, errs.ErrConsumerIsNil
	}

	if routingKeys == nil {
		return nil, errs.ErrRoutingKeyIsNil
	}

	chanMgr, err := newChannelMgr(r.connMgr, r.logger, r.reconnectInterval)
	if err != nil {
		return nil, err
	}

	reconnectCh, closeCh := chanMgr.notifyReconnect()

	cons := &consumer{
		chanManager:                chanMgr,
		reconnectErrCh:             reconnectCh,
		closeConnectionToManagerCh: closeCh,
		isClosedMu:                 new(sync.RWMutex),
		isClosed:                   false,
		options:                    defaultOptions,
		logger:                     r.logger,
		exchange:                   exchange,
		queue:                      queue,
		consumer:                   consume,
		routingKeys:                routingKeys,
		handler:                    messageHandler,
		enc:                        r.enc,
	}

	go func() {
		for err := range cons.reconnectErrCh {
			cons.logger.Info(fmt.Sprintf("successful consumer recovery from: %v", err))
			if err := cons.startConsume(); err != nil {
				cons.logger.Warn(fmt.Sprintf("error restarting consumer goroutines after cancel or close: %v", err))
			}
		}
	}()

	return cons, nil
}

func (r *consumer) Start() error {
	return r.startConsume()
}

func (r *consumer) Close() {
	r.isClosedMu.Lock()
	defer r.isClosedMu.Unlock()
	r.isClosed = true
	// close the channel so that rabbitmq server knows that the
	// r has been stopped.
	err := r.chanManager.close()
	if err != nil {
		r.logger.Error(fmt.Sprintf("error while closing the channel: %v", err))
	}

	go func() {
		r.closeConnectionToManagerCh <- struct{}{}
	}()
}

func (r *consumer) startConsume() error {
	err := r.chanManager.qosSafe(
		r.options.QOSPrefetch,
		0,
		r.options.QOSGlobal,
	)
	if err != nil {
		return fmt.Errorf("declare qos failed: %w", err)
	}

	err = declareExchange(r.chanManager, r.exchange)
	if err != nil {
		return fmt.Errorf("declare exchange failed: %w", err)
	}
	err = declareQueue(r.chanManager, r.queue)
	if err != nil {
		return fmt.Errorf("declare queue failed: %w", err)
	}
	err = declareBindings(r.chanManager, r.queue.Name, r.exchange.Name, r.routingKeys, r.consumer)
	if err != nil {
		return fmt.Errorf("declare bindings failed: %w", err)
	}

	msgs, err := r.chanManager.consumeSafe(
		r.queue.Name,
		r.consumer.Name,
		r.consumer.AutoAck,
		r.consumer.Exclusive,
		r.consumer.NoLocal,
		r.consumer.NoWait,
		r.consumer.Arguments,
	)
	if err != nil {
		return err
	}

	for i := 0; i < r.options.Concurrency; i++ {
		go handleMsg(i, r.consumer.Name, r, msgs, r.handler)
	}
	r.logger.Info(
		fmt.Sprintf("[consumer: %s] Processing messages on %v goroutines", r.consumer.Name, r.options.Concurrency),
	)
	return nil
}

func (r *consumer) getIsClosed() bool {
	r.isClosedMu.RLock()
	defer r.isClosedMu.RUnlock()
	return r.isClosed
}

func handleMsg(id int, name string, consumer *consumer, msgs <-chan rabbit.Delivery, handler types.ConsumerHandler) {
	for msg := range msgs {
		if consumer.getIsClosed() {
			break
		}

		delivery := types.Delivery{
			Headers:         msg.Headers,
			ContentType:     msg.ContentType,
			ContentEncoding: msg.ContentEncoding,
			DeliveryMode:    msg.DeliveryMode,
			Priority:        msg.Priority,
			CorrelationId:   msg.CorrelationId,
			ReplyTo:         msg.ReplyTo,
			Expiration:      msg.Expiration,
			MessageId:       msg.MessageId,
			Timestamp:       msg.Timestamp,
			Type:            msg.Type,
			UserId:          msg.UserId,
			AppId:           msg.AppId,
			ConsumerTag:     msg.ConsumerTag,
			MessageCount:    msg.MessageCount,
			DeliveryTag:     msg.DeliveryTag,
			Redelivered:     msg.Redelivered,
			Exchange:        msg.Exchange,
			RoutingKey:      msg.RoutingKey,
			Body:            msg.Body,
		}

		if consumer.consumer.AutoAck {
			handler(msg.RoutingKey, func(vPtr any) (types.Delivery, error) {
				return delivery, consumer.enc.Decode(msg.Body, vPtr)
			})
			continue
		}

		switch handler(msg.RoutingKey, func(vPtr any) (types.Delivery, error) {
			return delivery, consumer.enc.Decode(msg.Body, vPtr)
		}) {
		case types.Ack:
			err := msg.Ack(false)
			if err != nil {
				consumer.logger.Error(fmt.Sprintf("can't ack message: %v", err))
			}
		case types.AckMultiple:
			err := msg.Ack(true)
			if err != nil {
				consumer.logger.Error(fmt.Sprintf("can't ack message: %v", err))
			}
		case types.NackDiscard:
			err := msg.Nack(false, false)
			if err != nil {
				consumer.logger.Error(fmt.Sprintf("can't nack message: %v", err))
			}
		case types.NackRequeue:
			err := msg.Nack(false, true)
			if err != nil {
				consumer.logger.Error(fmt.Sprintf("can't nack message: %v", err))
			}
		case types.Reject:
			err := msg.Reject(false)
			if err != nil {
				consumer.logger.Error(fmt.Sprintf("can't nack message: %v", err))
			}
		case types.RejectRequeue:
			err := msg.Reject(true)
			if err != nil {
				consumer.logger.Error(fmt.Sprintf("can't nack message: %v", err))
			}
		}
	}
	consumer.logger.Error(fmt.Sprintf("[Consumer %d: %s] consumer goroutine closed", id, name))
}
