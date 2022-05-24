package main

import (
	"context"
	"encoding/base64"
	"github.com/nats-io/nats.go"
	uuid "github.com/satori/go.uuid"
	"log"
	"net"
	"strconv"
	"time"
)

//Durable (Name)
//MaxAckPending

type customDialer struct {
	ctx             context.Context
	nc              *nats.Conn
	connectTimeout  time.Duration
	connectTimeWait time.Duration
}

func (cd *customDialer) Dial(network, address string) (net.Conn, error) {
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
		err  error
		conn *nats.Conn
		jet  nats.JetStreamContext
		//sub  *nats.Subscription
		id = 0
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cd := &customDialer{
		ctx:             ctx,
		connectTimeout:  10 * time.Second,
		connectTimeWait: 1 * time.Second,
	}

	opty := []nats.Option{
		nats.Token("All4Stream2@22!"),
		nats.SetCustomDialer(cd),
		nats.ReconnectWait(2 * time.Second),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Println("Reconnected to", c.ConnectedUrl())
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Println("NATS connection is closed.")
		}),
		//nats.NoReconnect(),
	}

	conn, err = nats.Connect("10.9.9.16:4222", opty...)
	if err != nil {
		log.Printf("error Connect:%v", err)
		return
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

	jet, err = conn.JetStream()
	if err != nil {
		log.Printf("error JetStream:%v", err)
		return
	}

	stream := &nats.StreamConfig{
		NoAck:        false,
		Name:         "TEST",
		Subjects:     []string{"XXX.*"},
		Retention:    nats.RetentionPolicy(2),
		MaxConsumers: 10,
		Discard:      nats.DiscardPolicy(1),
		Storage:      nats.StorageType(0),
		MaxMsgs:      100,
		MaxAge:       time.Second * 10,
	}

	if _, err := jet.AddStream(stream); err != nil {
		log.Printf("error AddStream:%v", err)
		return
	}

	//cf := &nats.ConsumerConfig{
	//	Durable:         "T",
	//	DeliverSubject:  "test",
	//	DeliverPolicy:   5,
	//	AckPolicy:       2,
	//	AckWait:         10000,
	//	MaxDeliver:      1,
	//	SampleFrequency: "100",
	//	Heartbeat:       time.Second * 5,
	//}

	//_, err = jet.AddConsumer("TEST", cf)
	//if err != nil {
	//	log.Printf("error AddConsumer:%v", err)
	//	return
	//}

	//msgId := createId()
	//_, err = jet.Publish("TEST.test", []byte(" Send MESSAGE id: "+strconv.Itoa(id)+" / MsgId:"+msgId), nats.MsgId(msgId))
	//if err != nil {
	//	log.Printf("error Publish:%v", err)
	//	return
	//}

	for {

		msgId := createId()

		<-time.After(time.Millisecond * 300)
		id++
		_, err = jet.Publish("XXX.test", []byte(" Send MESSAGE id: "+strconv.Itoa(id)+" / MsgId:"+msgId), nats.MsgId(msgId))
		if err != nil {
			log.Printf("error Publish:%v", err)
			return
		}

		log.Printf("Send message id: %v / MsgId: %v", strconv.Itoa(id), msgId)
	}
}

func createId() string {
	var err error
	return base64.URLEncoding.EncodeToString(uuid.Must(uuid.NewV4(), err).Bytes()[:6])
}
