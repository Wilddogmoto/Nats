package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
)

func main() {

	var (
		err  error
		conn *nats.Conn
		jet  nats.JetStreamContext
		sub  *nats.Subscription
		ch   = make(chan *nats.Msg, 100)
	)

	opty := []nats.Option{
		nats.UserInfo("1234", "777"),
	}

	conn, err = nats.Connect("0.0.0.0:4222", opty...)
	if err != nil {
		log.Printf("error Connect:%v", err)
		return
	}

	log.Println("connect listener to", conn.ConnectedUrl())

	jet, err = conn.JetStream()
	if err != nil {
		log.Printf("error JetStream:%v", err)
		return
	}

	log.Println("create jet stream")

	sub, err = jet.ChanQueueSubscribe("XXX.test", "CORE", ch)
	if err != nil {
		log.Printf("error ChanQueueSubscribe:%v", err)
		return
	}

	i := 0
	for msg := range ch {
		fmt.Printf("[Listener got meesage][id: %d] %+v\n", i, string(msg.Data))
		i++
	}

	////go func() {
	//sub, err = jet.ChanSubscribe("XXX.test", ch)
	//if err != nil {
	//	log.Printf("error ChanSubscribe:%v", err)
	//	return
	//}
	//ii := 0
	//for msg := range ch {
	//	fmt.Printf("[ALL Listener got meesage][id: %d] %+v\n", ii, string(msg.Data))
	//	ii++
	//}
	////}()

	defer func() {
		conn.Close()
		sub.Unsubscribe()
		sub.Drain()
	}()

}
