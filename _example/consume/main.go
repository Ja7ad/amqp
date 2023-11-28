package main

import (
	"fmt"
	"log"

	"github.com/Ja7ad/amqp"
	"github.com/Ja7ad/amqp/types"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Greeting struct {
	Msg string `json:"msg"`
}

const (
	routingKeyPerson   = "person"
	routingKeyGreeting = "greeting"
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
				Key:     "greeting",
				Declare: true,
			},
			{
				Key:     "person",
				Declare: true,
			},
		},
		handler,
		amqp.WithConcurrentConsumer(10),
	)

	if err := con.Start(); err != nil {
		log.Fatal(err)
	}

	<-done
}

func handler(routingKey string, msgFunc func(vPtr any) (types.Delivery, error)) types.Action {
	switch routingKey {
	case routingKeyPerson:
		person := new(Person)
		msg, err := msgFunc(person)
		if err != nil {
			fmt.Println(err)
			return types.NackDiscard
		}

		fmt.Printf("routingKey: %s, msg: %v\n", msg.RoutingKey, person)
		return types.Ack
	case routingKeyGreeting:
		greeting := new(Greeting)
		msg, err := msgFunc(greeting)
		if err != nil {
			fmt.Println(err)
			return types.NackDiscard
		}

		fmt.Printf("routingKey: %s, msg: %v\n", msg.RoutingKey, greeting)
		return types.Ack
	default:
		return types.Reject
	}
}
