package amqp

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Ja7ad/amqp/logger"
)

type dispatcher struct {
	subscribers map[int]subscriber
	mu          *sync.Mutex
	logger      logger.Logger
}

type subscriber struct {
	notifyCancelOrCloseChan chan error
	closeCh                 <-chan struct{}
}

func newDispatcher(logger logger.Logger) *dispatcher {
	return &dispatcher{
		subscribers: make(map[int]subscriber),
		mu:          new(sync.Mutex),
		logger:      logger,
	}
}

func (d *dispatcher) dispatch(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, s := range d.subscribers {
		select {
		case <-time.After(5 * time.Second):
			d.logger.Error("Unexpected rabbitmq error: timeout in dispatcher")
		case s.notifyCancelOrCloseChan <- err:
		}
	}
}

func (d *dispatcher) addSubscriber() (<-chan error, chan<- struct{}) {
	const maxRand = math.MaxInt
	const minRand = 0
	id := rand.Intn(maxRand-minRand) + minRand

	closeCh := make(chan struct{})
	notifyCancelOrCloseChan := make(chan error)

	d.mu.Lock()
	d.subscribers[id] = subscriber{
		notifyCancelOrCloseChan: notifyCancelOrCloseChan,
		closeCh:                 closeCh,
	}
	d.mu.Unlock()

	go func(id int) {
		<-closeCh
		d.mu.Lock()
		defer d.mu.Unlock()

		sub, ok := d.subscribers[id]
		if !ok {
			return
		}
		close(sub.notifyCancelOrCloseChan)
		delete(d.subscribers, id)
	}(id)
	return notifyCancelOrCloseChan, closeCh
}
