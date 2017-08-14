package rmqlib

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type rmqLib struct {
	options Connection
	conn    *amqp.Connection
}

func New(opts Connection) *rmqLib {
	return &rmqLib{options: opts}
}

func (rl *rmqLib) Open() error {
	fmt.Printf("Opening connection with %#v\n", rl.options)

	conn, err := amqp.Dial(rl.options.ConnectionString)
	if err != nil {
		// un-recoverable
		log.Fatalf("amqp.connection.open: %s", err)
	}

	onError := func(msg string, err error) {
		if err != nil {
			defer conn.Close()
			log.Fatalf("Open::onError:%s: %s", msg, err)
		}
	}

	// Communicate everything over a channel (Socket abstraction)
	channel, err := conn.Channel()
	onError("ampq.channel.open", err)
	defer channel.Close()

	// need to make sure we have exchanges, queues and bindings
	err = rl.options.Exchange.Declare(channel)
	onError("amqp.exchange.declare", err)

	err = rl.options.Queue.Declare(channel)
	onError("amqp.queue.declare", err)

	err = rl.options.Binding.Bind(channel, rl.options.Exchange, rl.options.Queue)
	onError("amqp.binding.bind", err)

	log.Printf("Connected: %s\n", conn)

	// only if everything is successful, save connection
	rl.conn = conn

	return nil
}

func (rl *rmqLib) Close() {
	log.Printf("Closing connection: %s\n", rl.conn)
	if rl.conn != nil {
		rl.conn.Close()
	}
}

func (rl *rmqLib) Channel() *amqp.Channel {
	// TODO: Qos
	channel, err := rl.conn.Channel()
	if err != nil {
		log.Fatalf("Cannot create channel : %v\n", err)
	}

	return channel
}

func (rl *rmqLib) Publish(payload []byte, contentType string) error {
	channel := rl.Channel()
	defer channel.Close()

	mandatory := false
	immedate := false
	err := channel.Publish(
		rl.options.Exchange.Name,
		rl.options.Binding.RoutingKey,
		mandatory,
		immedate,
		amqp.Publishing{
			ContentType: contentType,
			Body:        payload,
		},
	)

	return err
}

func (rl *rmqLib) Consume() ([]byte, error) {
	channel := rl.Channel()
	defer channel.Close()

	// FIXME: configurable / programatic
	prefetchCount := 1
	prefetchSize := 0
	global := false
	err := channel.Qos(prefetchCount, prefetchSize, global)
	if err != nil {
		log.Fatalf("basic.qos: %v", err)
	}

	// FIXME: configurable / programatic
	consumer := "Consumer-id"
	autoAck := true
	noLocal := false // not supported by RMQ
	messages, err := channel.Consume(
		rl.options.Queue.Name,
		consumer,
		autoAck,
		rl.options.Queue.Exclusive,
		noLocal,
		rl.options.Exchange.NoWait,
		nil,
	)

	if err != nil {
		log.Fatalf("basic.consume: %v", err)
	}

	buff := new(bytes.Buffer)
	go func() {
		for m := range messages {
			log.Printf("Message : %v\n", m)
			buff.Write(m.Body)
		}
	}()

	time.Sleep(2 * time.Second)

	err = channel.Cancel(consumer, false)
	if err != nil {
		log.Fatalf("basic.cancel: %v", err)
	}

	return buff.Bytes(), nil
}
