package main

import (
	"fmt"
	"log"

	"github.com/sscaling/rmqcat/rmqlib"
)

const (
	exchangeName = "e" // default
	queueName    = "q" // default
	routingKey   = ""  // default
)

func main() {

	exchange := rmqlib.Exchange{
		Name:    exchangeName,
		Kind:    "topic",
		Durable: true,
	}

	queue := rmqlib.Queue{
		Name:    queueName,
		Durable: true,
	}

	binding := rmqlib.Binding{
		RoutingKey: routingKey,
	}

	opts := rmqlib.Connection{
		Exchange:         exchange,
		Queue:            queue,
		Binding:          binding,
		Name:             "test-connection",
		ConnectionString: "amqp://guest:guest@localhost:5672/",
	}

	rmq := rmqlib.New(opts)
	rmq.Open()
	defer rmq.Close()

	err := rmq.Publish([]byte{'a', 'b', 'c'}, "text/plain")
	if err != nil {
		log.Fatalf("Publish error: %v\n", err)
	}

	payload, err := rmq.Consume()
	if err != nil {
		log.Fatalf("Consume error: %v\n", err)
	}

	fmt.Printf("Payload %s\n", payload)

	fmt.Printf("Done\n")
}

/*
func consumeForever(done chan bool) {
	// Connects opens an AMQP connection from the credentials in the URL.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}
	defer conn.Close()

	for {
		go consume(conn)

		select {
		case <-done:
			return
		case <-time.After(1 * time.Second):
			log.Print("Waiting for done")
		}
	}
}

func oldmain() {

	// deal with consumption
	done := make(chan bool)
	go consumeForever(done)

	log.Println("Waiting for consumption to start")
	time.Sleep(1 * time.Second)

	// Connects opens an AMQP connection from the credentials in the URL.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}

	// This waits for a server acknowledgment which means the sockets will have
	// flushed all outbound publishings prior to returning.  It's important to
	// block on Close to not lose any publishings.
	defer conn.Close()

	c, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel.open: %s", err)
	}

	// We declare our topology on both the publisher and consumer to ensure they
	// are the same.  This is part of AMQP being a programmable messaging model.
	//
	// See the Channel.Consume example for the complimentary declare.
	err = exchangeDeclare(c)
	if err != nil {
		log.Fatalf("exchange.declare: %v", err)
	}

	// Prepare this message to be persistent.  Your publishing requirements may
	// be different.
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         []byte("Go Go AMQP!"),
	}

	// This is not a mandatory delivery, so it will be dropped if there are no
	// queues bound to the logs exchange.
	mandatory := false
	immediate := false
	err = c.Publish(exchangeName, routingKey, mandatory, immediate, msg)
	if err != nil {
		// Since publish is asynchronous this can happen if the network connection
		// is reset or if the server has run out of resources.
		log.Fatalf("basic.publish: %v", err)
	}

	log.Println("Success")

	time.Sleep(4 * time.Second)

	done <- true

	log.Println("shutdown")
}
*/
