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
