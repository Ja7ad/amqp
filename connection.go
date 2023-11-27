package amqp

import (
	"fmt"
	"sync"
	"time"

	"github.com/Ja7ad/amqp/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type connection struct {
	logger              logger.Logger
	url                 string
	connection          *amqp.Connection
	amqpConfig          amqp.Config
	connectionRWMu      *sync.RWMutex
	reconnectInterval   time.Duration
	reconnectionCount   uint
	reconnectionCountMu *sync.Mutex
	dispatcher          *dispatcher
}

func newConnMgr(url string, conf amqp.Config, log logger.Logger, reconnectInterval time.Duration) (*connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	mgr := &connection{
		logger:              log,
		url:                 url,
		amqpConfig:          conf,
		connection:          conn,
		connectionRWMu:      new(sync.RWMutex),
		reconnectInterval:   reconnectInterval,
		reconnectionCount:   0,
		reconnectionCountMu: new(sync.Mutex),
		dispatcher:          newDispatcher(log),
	}

	go mgr.startNotifyClose()
	return mgr, nil
}

func (c *connection) startNotifyClose() {
	notifyCloseChan := c.connection.NotifyClose(make(chan *amqp.Error, 1))

	err := <-notifyCloseChan
	if err != nil {
		c.logger.Error(fmt.Sprintf("attempting to reconnect to amqp server after connection close with error: %v", err))
		c.reconnectLoop()
		c.logger.Warn("successfully reconnected to amqp server")
		c.dispatcher.dispatch(err)
	}
	if err == nil {
		c.logger.Warn("amqp connection closed gracefully")
	}
}

func (c *connection) close() error {
	c.logger.Warn("closing connection manager...")
	c.connectionRWMu.Lock()
	defer c.connectionRWMu.Unlock()
	return c.connection.Close()
}

func (c *connection) notifyReconnect() (<-chan error, chan<- struct{}) {
	return c.dispatcher.addSubscriber()
}

func (c *connection) checkoutConnection() *amqp.Connection {
	c.connectionRWMu.RLock()
	return c.connection
}

func (c *connection) checkinConnection() {
	c.connectionRWMu.RUnlock()
}

func (c *connection) getReconnectionCount() uint {
	c.reconnectionCountMu.Lock()
	defer c.reconnectionCountMu.Unlock()
	return c.reconnectionCount
}

func (c *connection) incrementReconnectionCount() {
	c.reconnectionCountMu.Lock()
	defer c.reconnectionCountMu.Unlock()
	c.reconnectionCount++
}

func (c *connection) reconnectLoop() {
	reconnectTimer := time.NewTimer(c.reconnectInterval)
	defer reconnectTimer.Stop()

	for {
		select {
		case <-reconnectTimer.C:
			c.logger.Info(fmt.Sprintf("attempting to reconnect to amqp server"))
			err := c.reconnect()
			if err != nil {
				c.logger.Error(fmt.Sprintf("error reconnecting to amqp server: %v", err))
				reconnectTimer.Reset(c.reconnectInterval)
			} else {
				c.incrementReconnectionCount()
				go c.startNotifyClose()
				return
			}
		}
	}
}

func (c *connection) reconnect() error {
	c.connectionRWMu.Lock()
	defer c.connectionRWMu.Unlock()
	newConn, err := amqp.DialConfig(c.url, c.amqpConfig)
	if err != nil {
		return err
	}

	if err = c.connection.Close(); err != nil {
		c.logger.Warn(fmt.Sprintf("error closing connection while reconnecting: %v", err))
	}

	c.connection = newConn
	return nil
}

// notifyBlockedSafe safely wraps the (*amqp.Connection).NotifyBlocked method
func (c *connection) notifyBlockedSafe(
	receiver chan amqp.Blocking,
) chan amqp.Blocking {
	c.connectionRWMu.RLock()
	defer c.connectionRWMu.RUnlock()

	return c.connection.NotifyBlocked(
		receiver,
	)
}
