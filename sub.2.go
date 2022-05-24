package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"time"
)

func main() {

	opty := []nats.Option{
		nats.UserInfo("1234", ""),
		nats.ReconnectWait(2 * time.Second),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Println("Reconnected to", c.ConnectedUrl())
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Println("NATS connection is closed.")
		}),
		nats.NoReconnect(),
	}

	conn, err := nats.Connect("0.0.0.0:4222", opty...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch := make(chan *nats.Msg, 64)
	sub, err := conn.ChanSubscribe("TEST", ch)
	if err != nil {
		log.Fatal(err)
	}

	i := 0

	for msg := range ch {
		fmt.Printf("[SUB-2][%d] %+v\n", i, string(msg.Data))
		i++
	}

	sub.Unsubscribe()
	sub.Drain()

	fmt.Println("fin")
}
