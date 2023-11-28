package amqp

import (
	"fmt"
	"time"

	"github.com/Ja7ad/amqp/encoder"

	"github.com/Ja7ad/amqp/logger"
	"github.com/Ja7ad/amqp/types"
)

type AMQP struct {
	connMgr                    *connection
	reconnectErrCh             <-chan error
	closeConnectionToManagerCh chan<- struct{}
	logger                     logger.Logger
	reconnectInterval          time.Duration
	enc                        types.Encoder
}

type Broker interface {
	// Consumer create new consumer instance
	Consumer(exchange *types.Exchange,
		queue *types.Queue,
		consumer *types.Consumer,
		routingKeys []*types.RoutingKey,
		messageHandler types.ConsumerHandler,
		options ...ConsumerOptions) (Consumer, error)

	// Publisher create a publisher instance
	Publisher(exchange *types.Exchange, confirmMode bool) (Publisher, error)

	// Close rabbitmq connection
	Close() error
}

// New create amqp object for consume and publish
func New(url string, options ...RabbitMQOptions) (Broker, error) {
	defaultOpt := defaultRabbitMQOptions()
	for _, opt := range options {
		opt(defaultOpt)
	}

	rabbit := &AMQP{
		reconnectInterval: defaultOpt.ReconnectInterval,
		logger:            defaultOpt.Logger,
	}

	switch defaultOpt.EncType {
	case types.JSON:
		rabbit.enc = new(encoder.JsonEncoder)
	case types.GOB:
		rabbit.enc = new(encoder.GobEncoder)
	case types.PROTO:
		rabbit.enc = new(encoder.ProtoBufEncoder)
	}

	connMgr, err := newConnMgr(url, defaultOpt.AMQPConfig, defaultOpt.Logger, defaultOpt.ReconnectInterval)
	if err != nil {
		return nil, err
	}

	reconnectCh, closeCh := connMgr.notifyReconnect()
	rabbit.connMgr = connMgr
	rabbit.reconnectErrCh = reconnectCh
	rabbit.closeConnectionToManagerCh = closeCh
	rabbit.logger = defaultOpt.Logger

	go rabbit.handleRestarts()
	return rabbit, nil
}

func (r *AMQP) Close() error {
	r.closeConnectionToManagerCh <- struct{}{}
	return r.connMgr.close()
}

func (r *AMQP) handleRestarts() {
	for err := range r.reconnectErrCh {
		r.logger.Info(fmt.Sprintf("successful connection recovery from: %v", err))
	}
}
