package amqp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Ja7ad/amqp/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type channel struct {
	logger              logger.Logger
	channel             *amqp.Channel
	connManager         *connection
	channelMu           *sync.RWMutex
	reconnectInterval   time.Duration
	reconnectionCount   uint
	reconnectionCountMu *sync.Mutex
	dispatcher          *dispatcher
}

func newChannelMgr(connMgr *connection, log logger.Logger, reconnectInterval time.Duration) (*channel, error) {
	ch, err := getNewChannel(connMgr)
	if err != nil {
		return nil, err
	}

	chMgr := &channel{
		logger:              log,
		channel:             ch,
		connManager:         connMgr,
		channelMu:           new(sync.RWMutex),
		reconnectInterval:   reconnectInterval,
		reconnectionCount:   0,
		reconnectionCountMu: new(sync.Mutex),
		dispatcher:          newDispatcher(log),
	}
	go chMgr.startNotifyCancelOrClosed()
	return chMgr, nil
}

func getNewChannel(connManager *connection) (*amqp.Channel, error) {
	conn := connManager.checkoutConnection()
	defer connManager.checkinConnection()

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (c *channel) startNotifyCancelOrClosed() {
	notifyCloseChan := c.channel.NotifyClose(make(chan *amqp.Error, 1))
	notifyCancelChan := c.channel.NotifyCancel(make(chan string, 1))

	select {
	case err := <-notifyCloseChan:
		if err != nil {
			c.logger.Warn(fmt.Sprintf("attempting to reconnect to amqp server after close with error: %v", err))
			c.reconnectLoop()
			c.logger.Warn("successfully reconnected to amqp server")
			c.dispatcher.dispatch(err)
		}
		if err == nil {
			c.logger.Warn("amqp channel closed gracefully")
		}
	case err := <-notifyCancelChan:
		c.logger.Warn(fmt.Sprintf("attempting to reconnect to amqp server after cancel with error: %s", err))
		c.reconnectLoop()
		c.logger.Info("successfully reconnected to amqp server after cancel")
		c.dispatcher.dispatch(errors.New(err))
	}
}

func (c *channel) getReconnectionCount() uint {
	c.reconnectionCountMu.Lock()
	defer c.reconnectionCountMu.Unlock()
	return c.reconnectionCount
}

func (c *channel) incrementReconnectionCount() {
	c.reconnectionCountMu.Lock()
	defer c.reconnectionCountMu.Unlock()
	c.reconnectionCount++
}

func (c *channel) reconnectLoop() {
	for {
		c.logger.Warn(fmt.Sprintf("waiting %s seconds to attempt to reconnect to amqp server", c.reconnectInterval))
		time.Sleep(c.reconnectInterval)
		err := c.reconnect()
		if err != nil {
			c.logger.Error(fmt.Sprintf("error reconnecting to amqp server: %v", err))
		} else {
			c.incrementReconnectionCount()
			go c.startNotifyCancelOrClosed()
			return
		}
	}
}

func (c *channel) reconnect() error {
	c.channelMu.Lock()
	defer c.channelMu.Unlock()
	newChannel, err := getNewChannel(c.connManager)
	if err != nil {
		return err
	}

	if err = c.channel.Close(); err != nil {
		c.logger.Warn(fmt.Sprintf("error closing channel while reconnecting: %v", err))
	}

	c.channel = newChannel
	return nil
}

func (c *channel) close() error {
	c.logger.Warn("closing channel manager...")
	c.channelMu.Lock()
	defer c.channelMu.Unlock()
	return c.channel.Close()
}

func (c *channel) notifyReconnect() (<-chan error, chan<- struct{}) {
	return c.dispatcher.addSubscriber()
}

func (c *channel) qosSafe(
	prefetchCount int, prefetchSize int, global bool,
) error {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.Qos(
		prefetchCount,
		prefetchSize,
		global,
	)
}

// queueDeclarePassiveSafe safely wraps the (*amqp.Channel).QueueDeclarePassive method
func (c *channel) queueDeclarePassiveSafe(
	name string,
	durable bool,
	autoDelete bool,
	exclusive bool,
	noWait bool,
	args amqp.Table,
) (amqp.Queue, error) {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.QueueDeclarePassive(
		name,
		durable,
		autoDelete,
		exclusive,
		noWait,
		args,
	)
}

// queueDeclareSafe safely wraps the (*amqp.Channel).QueueDeclare method
func (c *channel) queueDeclareSafe(
	name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table,
) (amqp.Queue, error) {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.QueueDeclare(
		name,
		durable,
		autoDelete,
		exclusive,
		noWait,
		args,
	)
}

// exchangeDeclarePassiveSafe safely wraps the (*amqp.Channel).ExchangeDeclarePassive method
func (c *channel) exchangeDeclarePassiveSafe(
	name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp.Table,
) error {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.ExchangeDeclarePassive(
		name,
		kind,
		durable,
		autoDelete,
		internal,
		noWait,
		args,
	)
}

// exchangeDeclareSafe safely wraps the (*amqp.Channel).ExchangeDeclare method
func (c *channel) exchangeDeclareSafe(
	name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp.Table,
) error {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.ExchangeDeclare(
		name,
		kind,
		durable,
		autoDelete,
		internal,
		noWait,
		args,
	)
}

// queueBindSafe safely wraps the (*amqp.Channel).QueueBind method
func (c *channel) queueBindSafe(
	name string, key string, exchange string, noWait bool, args amqp.Table,
) error {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.QueueBind(
		name,
		key,
		exchange,
		noWait,
		args,
	)
}

// consumeSafe safely wraps the (*amqp.Channel).Consume method
func (c *channel) consumeSafe(
	queue,
	consumer string,
	autoAck,
	exclusive,
	noLocal,
	noWait bool,
	args amqp.Table,
) (<-chan amqp.Delivery, error) {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.Consume(
		queue,
		consumer,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)
}

// publishSafe safely wraps the (*amqp.Channel).Publish method.
func (c *channel) publishSafe(
	exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing,
) error {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.PublishWithContext(
		context.Background(),
		exchange,
		key,
		mandatory,
		immediate,
		msg,
	)
}

// publishWithContextSafe safely wraps the (*amqp.Channel).PublishWithContext method.
func (c *channel) publishWithContextSafe(
	ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing,
) error {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.PublishWithContext(
		ctx,
		exchange,
		key,
		mandatory,
		immediate,
		msg,
	)
}

func (c *channel) publishWithDeferredConfirmWithContextSafe(
	ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing,
) (*amqp.DeferredConfirmation, error) {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.PublishWithDeferredConfirmWithContext(
		ctx,
		exchange,
		key,
		mandatory,
		immediate,
		msg,
	)
}

// notifyReturnSafe safely wraps the (*amqp.Channel).NotifyReturn method
func (c *channel) notifyReturnSafe(
	ch chan amqp.Return,
) chan amqp.Return {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.NotifyReturn(
		ch,
	)
}

// confirmSafe safely wraps the (*amqp.Channel).Confirm method
func (c *channel) confirmSafe(
	noWait bool,
) error {
	c.channelMu.Lock()
	defer c.channelMu.Unlock()

	return c.channel.Confirm(
		noWait,
	)
}

// notifyPublishSafe safely wraps the (*amqp.Channel).NotifyPublish method
func (c *channel) notifyPublishSafe(
	confirm chan amqp.Confirmation,
) chan amqp.Confirmation {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.NotifyPublish(
		confirm,
	)
}

// notifyFlowSafe safely wraps the (*amqp.Channel).NotifyFlow method
func (c *channel) notifyFlowSafe(
	ch chan bool,
) chan bool {
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	return c.channel.NotifyFlow(
		ch,
	)
}
