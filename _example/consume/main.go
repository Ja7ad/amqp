package main

import (
	"fmt"
	"log"

	"github.com/Ja7ad/amqp"
	"github.com/Ja7ad/amqp/types"
)

func main() {
	done := make(chan struct{})

	b, err := amqp.New("uri")
	if err != nil {
		log.Fatal(err)
	}

	con, err := b.Consumer(
		&types.Exchange{
			Name:       "test",
			Kind:       types.Topic,
			Declare:    true,
			Passive:    false,
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Arguments:  nil,
		},
		&types.Queue{
			Name:       "test",
			Declare:    true,
			Durable:    true,
			Passive:    false,
			Exclusive:  false,
			AutoDelete: false,
			NoWait:     false,
			Arguments:  nil,
		},
		&types.Consumer{
			Name:      "test1",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Arguments: nil,
		},
		[]*types.RoutingKey{
			{
				Key:     "foo",
				Declare: true,
			},
			{
				Key:     "bar",
				Declare: false,
			},
		},
		handler(print),
		amqp.WithConcurrentConsumer(10),
	)

	if err := con.Start(); err != nil {
		log.Fatal(err)
	}

	<-done
}

func handler(print func(msg []byte)) types.ConsumerHandler {
	return func(d types.Delivery) (action types.Action) {
		print(d.Body)
		return types.Ack
	}
}

func print(msg []byte) {
	fmt.Println(string(msg))
}
