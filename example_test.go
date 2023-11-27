package amqp

import (
	"fmt"
	"github.com/Ja7ad/amqp/logger"
	"log"
)

func ExampleNew() {
	lg, err := logger.New(logger.CONSOLE_HANDLER, logger.Options{
		Development:  true,
		Debug:        false,
		EnableCaller: true,
		SkipCaller:   3,
	})
	if err != nil {
		log.Fatal(err)
	}

	rb, err := New("uri", lg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rb)
}
