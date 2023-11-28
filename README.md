# amqp [![Go Reference](https://pkg.go.dev/badge/github.com/Ja7ad/amqp.svg)](https://pkg.go.dev/github.com/Ja7ad/amqp)
The AMQP (Advanced Message Queuing Protocol) package in Go is a wrapper for amqp091-go, offering a specific focus on stable and secure connection management. This package provides a high-level interface for interacting with RabbitMQ, emphasizing reliability and safety in connection handling. Key features include automatic reconnection strategies, a simplified API for creating consumers and publishers, and graceful connection closure. By wrapping amqp091-go with stability and safety in mind, this package facilitates robust and secure messaging in Go applications. Explore the documentation to leverage the power of AMQP with confidence in your projects.

## Goals
- **Reliable Reconnection Management:** Develop a robust reconnection mechanism with a specific connection and channel manager running in the background.
Ensure seamless and automatic re-establishment of connections in case of interruptions or failures.

- **Concurrent Safe Message Consumption:** Implement concurrent-safe methods for consuming messages, allowing multiple consumers to handle messages concurrently without conflicts.
Ensure thread-safety in message consumption to support high-concurrency scenarios.

- **Custom Publish Method with Retry Support:** Design a custom publish method to provide flexibility in message publishing.
Incorporate retry support in the publish mechanism to automatically retry failed message deliveries, enhancing message reliability.

- **Consumer with Fault Tolerance:** Develop a fault-tolerant consumer that can gracefully handle errors and exceptions during message processing.
Implement mechanisms to handle and recover from faults, ensuring continuous and reliable message consumption.

- **Auto Encoder (JSON, GOB, ProtoBuf)**: auto encode your message body with any type on publish then on consume auto decode message body on your variable

## Install

```shell
go get -u github.com/Ja7ad/amqp
```

## Example Consumer

```go
package main

import (
	"fmt"
	"github.com/Ja7ad/amqp"
	"github.com/Ja7ad/amqp/types"
	"log"
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
```

## Example Publisher

```go
package main

import (
	"fmt"
	"github.com/Ja7ad/amqp"
	"github.com/Ja7ad/amqp/types"
	"log"
	"time"
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
	rb, err := amqp.New("uri")
	if err != nil {
		log.Fatal(err)
	}

	pub, err := rb.Publisher(&types.Exchange{
		Name:       "test",
		Kind:       types.Topic,
		Declare:    true,
		Passive:    false,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Arguments:  nil,
	}, false)
	if err != nil {
		log.Fatal(err)
	}

	person := &Person{
		Name: "javad",
		Age:  30,
	}

	if err := pub.Publish(false, false, types.Publishing{
		DeliveryMode: types.Persistent,
		Body:         person,
	}, routingKeyPerson); err != nil {
		log.Fatal(err)
	}

	greeting := &Greeting{
		Msg: "foobar",
	}

	if err := pub.Publish(false, false, types.Publishing{
		DeliveryMode: types.Persistent,
		Body:         greeting,
	}, routingKeyGreeting); err != nil {
		log.Fatal(err)
	}
}
```