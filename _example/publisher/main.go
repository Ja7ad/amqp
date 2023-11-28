package main

import (
	"log"

	"github.com/Ja7ad/amqp"
	"github.com/Ja7ad/amqp/types"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

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
	}, "foo"); err != nil {
		log.Fatal(err)
	}
}
