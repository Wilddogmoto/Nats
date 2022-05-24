package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
)

func main() {

	conn, err := nats.Connect("0.0.0.0:4222")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch := make(chan *nats.Msg, 64)
	sub, err := conn.ChanSubscribe("command", ch)
	if err != nil {
		log.Fatal(err)
	}

	i := 0

	for msg := range ch {
		fmt.Printf("[SUB-1][%d] %+v\n", i, string(msg.Data))
		i++
	}

	sub.Unsubscribe()
	sub.Drain()

	fmt.Println("fin")
}
