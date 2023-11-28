package amqp

import (
	"fmt"
	"log"
)

func ExampleNew() {
	rb, err := New("uri")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rb)
}
