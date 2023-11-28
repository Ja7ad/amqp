package amqp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Ja7ad/amqp/errs"

	"github.com/Ja7ad/amqp/logger"
	"github.com/Ja7ad/amqp/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	_defaultRetryDelay = 3 * time.Second
	_defaultMaxRetries = 5
)

type publisher struct {
	chanManager                *channel
	connManager                *connection
	reconnectErrCh             <-chan error
	closeConnectionToManagerCh chan<- struct{}

	disablePublishDueToFlow    bool
	disablePublishDueToFlowMux *sync.RWMutex

	disablePublishDueToBlocked    bool
	disablePublishDueToBlockedMux *sync.RWMutex

	handlerMux           *sync.Mutex
	notifyReturnHandler  func(r types.Return)
	notifyPublishHandler func(p types.Confirmation)

	logger   logger.Logger
	exchange *types.Exchange
	enc      types.Encoder
}

type Publisher interface {
	/*
		Publish sends a Publishing from the client to an exchange on the server.

		When you want a single message to be delivered to a single queue, you can
		publish to the default exchange with the routingKey of the queue name.  This is
		because every declared queue gets an implicit route to the default exchange.

		Since publishings are asynchronous, any undeliverable message will get returned
		by the server.  Add a listener with Channel.NotifyReturn to handle any
		undeliverable message when calling publish with either the mandatory or
		immediate parameters as true.

		Publishings can be undeliverable when the mandatory flag is true and no queue is
		bound that matches the routing key, or when the immediate flag is true and no
		consumer on the matched queue is ready to accept the delivery.

		This can return an error when the channel, connection or socket is closed.  The
		error or lack of an error does not indicate whether the server has received this
		publishing.

		It is possible for publishing to not reach the broker if the underlying socket
		is shut down without pending publishing packets being flushed from the kernel
		buffers.  The easy way of making it probable that all publishings reach the
		server is to always call Connection.Close before terminating your publishing
		application.  The way to ensure that all publishings reach the server is to add
		a listener to Channel.NotifyPublish and put the channel in confirm mode with
		Channel.Confirm.  Publishing delivery tags and their corresponding
		confirmations startConsume at 1.  Exit when all publishings are confirmed.

		When Publish does not return an error and the channel is in confirm mode, the
		internal counter for DeliveryTags with the first confirmation starts at 1.

		Note: routingKey is specific keys in queue for example (subject, topic)
	*/
	Publish(
		mandatory bool,
		immediate bool,
		msg types.Publishing,
		routingKeys ...string,
	) error

	/*
		PublishWithContext sends a Publishing from the client to an exchange on the server and control by prent context.

		When you want a single message to be delivered to a single queue, you can
		publish to the default exchange with the routingKey of the queue name.  This is
		because every declared queue gets an implicit route to the default exchange.

		Since publishings are asynchronous, any undeliverable message will get returned
		by the server.  Add a listener with Channel.NotifyReturn to handle any
		undeliverable message when calling publish with either the mandatory or
		immediate parameters as true.

		Publishings can be undeliverable when the mandatory flag is true and no queue is
		bound that matches the routing key, or when the immediate flag is true and no
		consumer on the matched queue is ready to accept the delivery.

		This can return an error when the channel, connection or socket is closed.  The
		error or lack of an error does not indicate whether the server has received this
		publishing.

		It is possible for publishing to not reach the broker if the underlying socket
		is shut down without pending publishing packets being flushed from the kernel
		buffers.  The easy way of making it probable that all publishings reach the
		server is to always call Connection.Close before terminating your publishing
		application.  The way to ensure that all publishings reach the server is to add
		a listener to Channel.NotifyPublish and put the channel in confirm mode with
		Channel.Confirm.  Publishing delivery tags and their corresponding
		confirmations startConsume at 1.  Exit when all publishings are confirmed.

		When Publish does not return an error and the channel is in confirm mode, the
		internal counter for DeliveryTags with the first confirmation starts at 1.

		Note: routingKey is specific keys in queue for example (subject, topic)
	*/
	PublishWithContext(
		ctx context.Context,
		mandatory bool,
		immediate bool,
		msg types.Publishing,
		routingKeys ...string,
	) error

	// PublishWithDeferredConfirmWithContext publishes the provided data to the given routing keys over the connection.
	// if the publisher is in confirm mode (which can be either done by calling `NotifyPublish` with a custom handler
	// or by using `WithPublisherOptionsConfirm`) a publisher confirmation is returned.
	// This confirmation can be used to check if the message was actually published or wait for this to happen.
	// If the publisher is not in confirm mode, the returned confirmation will always be nil.
	PublishWithDeferredConfirmWithContext(
		ctx context.Context,
		mandatory bool,
		immediate bool,
		msg types.Publishing,
		routingKeys ...string,
	) (types.PublisherConfirmation, error)

	// PublishWithRetry sends a Publishing from the client to an exchange on the server,
	// controlled by the provided context. It incorporates a retry mechanism, attempting
	// to publish the message multiple times with a configurable delay and maximum number
	// of retries.
	//
	// When you want a single message to be delivered to a specific queue, you can publish
	// to the default exchange with the routingKey set to the queue name. This is because
	// every declared queue gets an implicit route to the default exchange.
	//
	// Since publishings are asynchronous, any undeliverable message will be returned by
	// the server. Add a listener with Channel.NotifyReturn to handle undeliverable
	// messages when calling publish with either the mandatory or immediate parameters as true.
	//
	// Publishings can be undeliverable when the mandatory flag is true and no queue is
	// bound that matches the routing key, or when the immediate flag is true and no
	// consumer on the matched queue is ready to accept the delivery.
	//
	// This function may return an error when the channel, connection, or socket is closed.
	// The error, or lack of an error, does not indicate whether the server has received this
	// publishing.
	//
	// It is possible for publishing to not reach the broker if the underlying socket
	// is shut down without pending publishing packets being flushed from the kernel
	// buffers. To increase the likelihood that all publishings reach the server, it is
	// recommended to always call Connection.Close before terminating your publishing
	// application. Alternatively, add a listener to Channel.NotifyPublish and put the channel
	// in confirm mode with Channel.Confirm. Publishing delivery tags and their corresponding
	// confirmations start at 1. Exit when all publishings are confirmed.
	//
	// When PublishWithRetry does not return an error and the channel is in confirm mode,
	// the internal counter for DeliveryTags with the first confirmation starts at 1.
	//
	// Note: routingKey represents specific keys in the queue, such as subject or topic.
	PublishWithRetry(
		ctx context.Context,
		mandatory bool,
		immediate bool,
		msg types.Publishing,
		config types.PublisherConfig,
		routingKeys ...string,
	) error

	// NotifyReturn registers a listener for basic.return methods.
	// These can be sent from the server when a publish is undeliverable either from the mandatory or immediate flags.
	// These notifications are shared across an entire connection, so if you're creating multiple
	// publishers on the same connection keep that in mind
	NotifyReturn(handler func(r types.Return))

	// NotifyPublish registers a listener for publish confirmations, must set ConfirmPublishings option
	// These notifications are shared across an entire connection, so if you're creating multiple
	// publishers on the same connection keep that in mind
	NotifyPublish(handler func(p types.Confirmation))

	Close()
}

// Publisher create publisher interface for publishing message
func (r *AMQP) Publisher(exchange *types.Exchange, confirmMode bool) (Publisher, error) {
	if exchange == nil {
		return nil, errs.ErrExchangeIsNil
	}

	chanManager, err := newChannelMgr(r.connMgr, r.logger, r.reconnectInterval)
	if err != nil {
		return nil, err
	}

	reconnectErrCh, closeCh := chanManager.notifyReconnect()
	pub := &publisher{
		chanManager:                   chanManager,
		connManager:                   r.connMgr,
		reconnectErrCh:                reconnectErrCh,
		closeConnectionToManagerCh:    closeCh,
		disablePublishDueToFlow:       false,
		disablePublishDueToFlowMux:    &sync.RWMutex{},
		disablePublishDueToBlocked:    false,
		disablePublishDueToBlockedMux: &sync.RWMutex{},
		handlerMux:                    &sync.Mutex{},
		notifyReturnHandler:           nil,
		notifyPublishHandler:          nil,
		exchange:                      exchange,
		logger:                        r.logger,
		enc:                           r.enc,
	}

	err = pub.startup()
	if err != nil {
		return nil, err
	}

	if confirmMode {
		pub.NotifyPublish(func(_ types.Confirmation) {
			// set a blank handler to set the channel in confirm mode
		})
	}

	go func() {
		for err := range pub.reconnectErrCh {
			pub.logger.Info(fmt.Sprintf("successful publisher recovery from: %v", err))
			err := pub.startup()
			if err != nil {
				pub.logger.Error(fmt.Sprintf("error on startup for publisher after cancel or close: %v", err))
				pub.logger.Error("publisher closing, unable to recover")
				return
			}
			pub.startReturnHandler()
			pub.startPublishHandler()
		}
	}()

	return pub, nil
}

func (r *publisher) Publish(
	mandatory bool,
	immediate bool,
	msg types.Publishing,
	routingKeys ...string,
) error {
	return r.PublishWithContext(context.Background(), mandatory, immediate, msg, routingKeys...)
}

func (r *publisher) PublishWithRetry(
	ctx context.Context,
	mandatory bool,
	immediate bool,
	msg types.Publishing,
	config types.PublisherConfig,
	routingKeys ...string,
) error {
	if config.RetryDelay == 0 {
		config.RetryDelay = _defaultRetryDelay
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = _defaultMaxRetries
	}

	var lastErr error
	ticker := time.NewTicker(config.RetryDelay)

	for i := 0; i <= config.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := r.PublishWithContext(ctx, mandatory, immediate, msg, routingKeys...)
			if err == nil {
				return nil
			}
			lastErr = err
		}
	}

	return lastErr
}

func (r *publisher) PublishWithContext(
	ctx context.Context,
	mandatory bool,
	immediate bool,
	msg types.Publishing,
	routingKeys ...string,
) error {
	r.disablePublishDueToFlowMux.RLock()
	defer r.disablePublishDueToFlowMux.RUnlock()
	if r.disablePublishDueToFlow {
		return fmt.Errorf("publishing blocked due to high flow on the server")
	}

	r.disablePublishDueToBlockedMux.RLock()
	defer r.disablePublishDueToBlockedMux.RUnlock()
	if r.disablePublishDueToBlocked {
		return fmt.Errorf("publishing blocked due to TCP block on the server")
	}

	if routingKeys == nil {
		return errs.ErrRoutingKeyIsNil
	}

	body, err := r.enc.Encode(msg.Body)
	if err != nil {
		return errs.ErrFailedEncode
	}

	for _, routingKey := range routingKeys {
		// Actual publish.
		err := r.chanManager.publishWithContextSafe(
			ctx,
			r.exchange.Name,
			routingKey,
			mandatory,
			immediate,
			amqp.Publishing{
				Headers:         msg.Headers,
				ContentType:     msg.ContentType,
				ContentEncoding: msg.ContentEncoding,
				DeliveryMode:    uint8(msg.DeliveryMode),
				Priority:        msg.Priority,
				CorrelationId:   msg.CorrelationId,
				ReplyTo:         msg.ReplyTo,
				Expiration:      msg.Expiration,
				MessageId:       msg.MessageId,
				Timestamp:       msg.Timestamp,
				Type:            msg.Type,
				UserId:          msg.UserId,
				AppId:           msg.AppId,
				Body:            body,
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *publisher) PublishWithDeferredConfirmWithContext(
	ctx context.Context,
	mandatory bool,
	immediate bool,
	msg types.Publishing,
	routingKeys ...string,
) (types.PublisherConfirmation, error) {
	r.disablePublishDueToFlowMux.RLock()
	defer r.disablePublishDueToFlowMux.RUnlock()
	if r.disablePublishDueToFlow {
		return nil, fmt.Errorf("publishing blocked due to high flow on the server")
	}

	r.disablePublishDueToBlockedMux.RLock()
	defer r.disablePublishDueToBlockedMux.RUnlock()
	if r.disablePublishDueToBlocked {
		return nil, fmt.Errorf("publishing blocked due to TCP block on the server")
	}

	var deferredConfirmations []*amqp.DeferredConfirmation

	body, err := r.enc.Encode(msg.Body)
	if err != nil {
		return types.PublisherConfirmation{}, errs.ErrFailedEncode
	}

	for _, routingKey := range routingKeys {
		// Actual publish.
		conf, err := r.chanManager.publishWithDeferredConfirmWithContextSafe(
			ctx,
			r.exchange.Name,
			routingKey,
			mandatory,
			immediate,
			amqp.Publishing{
				Headers:         msg.Headers,
				ContentType:     msg.ContentType,
				ContentEncoding: msg.ContentEncoding,
				DeliveryMode:    uint8(msg.DeliveryMode),
				Priority:        msg.Priority,
				CorrelationId:   msg.CorrelationId,
				ReplyTo:         msg.ReplyTo,
				Expiration:      msg.Expiration,
				MessageId:       msg.MessageId,
				Timestamp:       msg.Timestamp,
				Type:            msg.Type,
				UserId:          msg.UserId,
				AppId:           msg.AppId,
				Body:            body,
			},
		)
		if err != nil {
			return nil, err
		}
		deferredConfirmations = append(deferredConfirmations, conf)
	}
	return deferredConfirmations, nil
}

func (r *publisher) NotifyReturn(handler func(r types.Return)) {
	r.handlerMux.Lock()
	start := r.notifyReturnHandler == nil
	r.notifyReturnHandler = handler
	r.handlerMux.Unlock()

	if start {
		r.startReturnHandler()
	}
}

func (r *publisher) NotifyPublish(handler func(p types.Confirmation)) {
	r.handlerMux.Lock()
	shouldStart := r.notifyPublishHandler == nil
	r.notifyPublishHandler = handler
	r.handlerMux.Unlock()

	if shouldStart {
		r.startPublishHandler()
	}
}

func (r *publisher) Close() {
	// close the channel so that rabbitmq server knows that the
	// publisher has been stopped.
	err := r.chanManager.close()
	if err != nil {
		r.logger.Error(fmt.Sprintf("error while closing the channel: %v", err))
	}
	r.logger.Warn("closing publisher...")
	go func() {
		r.closeConnectionToManagerCh <- struct{}{}
	}()
}

func (r *publisher) startup() error {
	err := declareExchange(r.chanManager, r.exchange)
	if err != nil {
		return fmt.Errorf("declare exchange failed: %w", err)
	}
	go r.startNotifyFlowHandler()
	go r.startNotifyBlockedHandler()
	return nil
}

func (r *publisher) startNotifyFlowHandler() {
	notifyFlowChan := r.chanManager.notifyFlowSafe(make(chan bool))
	r.disablePublishDueToFlowMux.Lock()
	r.disablePublishDueToFlow = false
	r.disablePublishDueToFlowMux.Unlock()

	for ok := range notifyFlowChan {
		r.disablePublishDueToFlowMux.Lock()
		if ok {
			r.logger.Warn("pausing publishing due to flow request from server")
			r.disablePublishDueToFlow = true
		} else {
			r.disablePublishDueToFlow = false
			r.logger.Warn("resuming publishing due to flow request from server")
		}
		r.disablePublishDueToFlowMux.Unlock()
	}
}

func (r *publisher) startNotifyBlockedHandler() {
	blocks := r.connManager.notifyBlockedSafe(make(chan amqp.Blocking))
	r.disablePublishDueToBlockedMux.Lock()
	r.disablePublishDueToBlocked = false
	r.disablePublishDueToBlockedMux.Unlock()

	for b := range blocks {
		r.disablePublishDueToBlockedMux.Lock()
		if b.Active {
			r.logger.Warn("pausing publishing due to TCP blocking from server")
			r.disablePublishDueToBlocked = true
		} else {
			r.disablePublishDueToBlocked = false
			r.logger.Warn("resuming publishing due to TCP blocking from server")
		}
		r.disablePublishDueToBlockedMux.Unlock()
	}
}

func (r *publisher) startReturnHandler() {
	r.handlerMux.Lock()
	if r.notifyReturnHandler == nil {
		r.handlerMux.Unlock()
		return
	}
	r.handlerMux.Unlock()

	go func() {
		returns := r.chanManager.notifyReturnSafe(make(chan amqp.Return, 1))
		for ret := range returns {
			go r.notifyReturnHandler(types.Return{Return: ret})
		}
	}()
}

func (r *publisher) startPublishHandler() {
	r.handlerMux.Lock()
	if r.notifyPublishHandler == nil {
		r.handlerMux.Unlock()
		return
	}
	r.handlerMux.Unlock()
	r.chanManager.confirmSafe(false)

	go func() {
		confirmationCh := r.chanManager.notifyPublishSafe(make(chan amqp.Confirmation, 1))
		for conf := range confirmationCh {
			go r.notifyPublishHandler(types.Confirmation{
				Confirmation:      conf,
				ReconnectionCount: int(r.chanManager.getReconnectionCount()),
			})
		}
	}()
}
