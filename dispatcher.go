package amqp

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Ja7ad/amqp/logger"
)

type dispatcher struct {
	subscribers sync.Map
	logger      logger.Logger
}

type subscriber struct {
	notifyCancelOrCloseChan chan error
	closeCh                 <-chan struct{}
}

func newDispatcher(logger logger.Logger) *dispatcher {
	return &dispatcher{
		subscribers: sync.Map{},
		logger:      logger,
	}
}

func (d *dispatcher) dispatch(err error) {
	d.subscribers.Range(func(_, value interface{}) bool {
		s := value.(subscriber)
		select {
		case <-time.After(5 * time.Second):
			d.logger.Error("Unexpected rabbitmq error: timeout in dispatcher")
		case s.notifyCancelOrCloseChan <- err:
		}
		return true
	})
}

func (d *dispatcher) addSubscriber() (<-chan error, chan<- struct{}) {
	const maxRand = math.MaxInt
	const minRand = 0
	id := rand.Intn(maxRand-minRand) + minRand

	closeCh := make(chan struct{})
	notifyCancelOrCloseChan := make(chan error)

	sub := subscriber{
		notifyCancelOrCloseChan: notifyCancelOrCloseChan,
		closeCh:                 closeCh,
	}

	d.subscribers.Store(id, sub)

	go func(id int) {
		<-closeCh
		d.subscribers.Delete(id)
		close(notifyCancelOrCloseChan)
	}(id)

	return notifyCancelOrCloseChan, closeCh
}
