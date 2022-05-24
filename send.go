package main

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
)

type ms struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

func main() {
	ch := make(chan *nats.Msg, 64)
	conn, err := nats.Connect("0.0.0.0:4222")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	sub, err := conn.ChanSubscribe("send", ch)
	if err != nil {
		log.Fatal(err)
	}

	i := 0

	for msg := range ch {
		var m ms

		if err := json.Unmarshal(msg.Data, &m); err != nil {
			fmt.Println(err)
		}

		i++
		fmt.Printf("[SEND][%d] %+v\n", i, msg)
		fmt.Printf("[MESSAGE][%d]: %+v\n", i, m)
	}

	sub.Unsubscribe()
	sub.Drain()

	fmt.Println("fin")
}
