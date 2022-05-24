package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"net"
	"strconv"
	"time"
)

type mst struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

type customD struct {
	ctx             context.Context
	nc              *nats.Conn
	connectTimeout  time.Duration
	connectTimeWait time.Duration
}

func (cd *customD) Dial(network, address string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(cd.ctx, cd.connectTimeout)
	defer cancel()

	for {
		log.Println("Attempting to connect to", address)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		select {
		case <-cd.ctx.Done():
			return nil, cd.ctx.Err()
		default:
			d := &net.Dialer{}
			if conn, err := d.DialContext(ctx, network, address); err == nil {
				log.Println("Connected to NATS successfully")
				return conn, nil
			} else {
				time.Sleep(cd.connectTimeWait)
			}
		}
	}
}

func main() {

	var (
		id     = 0
		ctx    context.Context
		cancel context.CancelFunc
		conn   *nats.Conn
		err    error
	)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	cd := &customD{
		ctx:             ctx,
		connectTimeout:  10 * time.Second,
		connectTimeWait: 1 * time.Second,
	}

	opty := []nats.Option{
		nats.SetCustomDialer(cd),
		nats.ReconnectWait(2 * time.Second),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Println("Reconnected to", c.ConnectedUrl())
		}),
		nats.DisconnectHandler(func(c *nats.Conn) {
			log.Println("Disconnected from NATS")
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Println("NATS connection is closed.")
		}),
		nats.NoReconnect(),
	}

	conn, err = nats.Connect("0.0.0.0:4222", opty...)
	if err != nil {
		panic(err)
	}

	getStatusTxt := func(conn *nats.Conn) string {
		switch conn.Status() {
		case nats.CONNECTED:
			return "Connected"
		case nats.CLOSED:
			return "Closed"
		default:
			return "Other"
		}
	}
	log.Printf("The connection is %v\n", getStatusTxt(conn))

	defer func() {
		conn.Close()
		log.Printf("The connection is %v\n", getStatusTxt(conn))
	}()

	for {
		var (
			b []byte
		)

		<-time.After(time.Millisecond * 500)

		m := &mst{
			Id:      id,
			Message: "test message",
		}

		id++

		b, err = json.Marshal(&m)
		if err != nil {
			fmt.Println(err)
		}

		mft := &nats.Msg{
			Subject: "send",
			Data:    b,
			Sub:     nil,
		}

		if err = conn.Publish("command", []byte("test message id: "+strconv.Itoa(id))); err != nil {
			log.Printf("error Publish:%v", err)
			return
		}

		if err = conn.PublishMsg(mft); err != nil {
			log.Printf("error PublishMsg:%v", err)
			return
		}
	}
}
