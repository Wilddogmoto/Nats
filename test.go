package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"strconv"
	"time"
)

func main() {
	conn, err := nats.Connect("0.0.0.0:4222")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	id := 0

	go func() {
		for {
			<-time.After(time.Millisecond * 500)
			id++
			if err := conn.Publish("command", []byte("number: "+strconv.Itoa(id))); err != nil {
				panic(err)
			}
		}

	}()

	ch := make(chan *nats.Msg, 64)
	sub, err := conn.ChanSubscribe("command", ch)
	if err != nil {
		log.Fatal(err)
	}

	i := 0

	for msg := range ch {
		fmt.Printf("[%d] %+v\n", i, string(msg.Data))
		i++
	}

	sub.Unsubscribe()
	sub.Drain()

	fmt.Println("fin")
}
