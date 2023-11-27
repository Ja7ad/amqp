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
	"github.com/Ja7ad/amqp/logger"
	"github.com/Ja7ad/amqp/types"
	"log"
)

func main() {
	done := make(chan struct{})
	lg, err := logger.New(logger.CONSOLE_HANDLER, logger.Options{
		Development:  true,
		Debug:        false,
		EnableCaller: true,
		SkipCaller:   3,
	})
	if err != nil {
		log.Fatal(err)
	}

	b, err := amqp.New("uri", lg)
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
```

## Example Publisher

```go
package main

import (
	"fmt"
	"github.com/Ja7ad/amqp"
	"github.com/Ja7ad/amqp/logger"
	"github.com/Ja7ad/amqp/types"
	"log"
	"time"
)

func main() {
	lg, err := logger.New(logger.CONSOLE_HANDLER, logger.Options{
		Development:  true,
		Debug:        false,
		EnableCaller: true,
		SkipCaller:   3,
	})
	if err != nil {
		log.Fatal(err)
	}

	rb, err := amqp.New("uri", lg)
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
	}, lg, false)
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	for {
		if err := pub.Publish(false, false, types.Publishing{
			DeliveryMode: types.Persistent,
			Body:         []byte(fmt.Sprintf("msg %d", i)),
		}, "foo"); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("message %d publised\n", i)

		i++
		time.Sleep(5 * time.Second)
	}
}
```