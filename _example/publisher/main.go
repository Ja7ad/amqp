package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Ja7ad/amqp"
	"github.com/Ja7ad/amqp/types"
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
