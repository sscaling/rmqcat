package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sscaling/rmqcat/options"
	"github.com/streadway/amqp"
)

const exchangeName string = "logs"
const queueName string = "page"
const routingKey string = "alert"

func exchangeDeclare(c *amqp.Channel) error {
	return c.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
}

func consume(conn *amqp.Connection) {

	log.Println("Waiting to consume")

	c, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel.open: %s", err)
	}

	// We declare our topology on both the publisher and consumer to ensure they
	// are the same.  This is part of AMQP being a programmable messaging model.
	//
	// See the Channel.Publish example for the complimentary declare.
	err = exchangeDeclare(c)
	if err != nil {
		log.Fatalf("exchange.declare: %s", err)
	}

	log.Println("Consume exchange declared")

	// Establish our queue topologies that we are responsible for
	type bind struct {
		queue string
		key   string
	}

	b := bind{queueName, routingKey}

	durable := true
	autoDelete := false
	exclusive := false
	nowait := false
	var amqpTable amqp.Table
	_, err = c.QueueDeclare(b.queue, durable, autoDelete, exclusive, nowait, amqpTable)
	if err != nil {
		log.Fatalf("queue.declare: %v", err)
	}

	exchange := exchangeName
	err = c.QueueBind(b.queue, b.key, exchange, nowait, amqpTable)
	if err != nil {
		log.Fatalf("queue.bind: %v", err)
	}

	log.Println("Bound to queue")

	// Set our quality of service.  Since we're sharing 3 consumers on the same
	// channel, we want at least 3 messages in flight.
	prefetchCount := 1
	prefetchSize := 0
	global := false
	err = c.Qos(prefetchCount, prefetchSize, global)
	if err != nil {
		log.Fatalf("basic.qos: %v", err)
	}

	// Establish our consumers that have different responsibilities.  Our first
	// two queues do not ack the messages on the server, so require to be acked
	// on the client.

	pages, err := c.Consume(queueName, "pager", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("basic.consume: %v", err)
	}

	go func() {
		for page := range pages {
			log.Printf("Page : %v\n", page)
			// ... this consumer is responsible for sending pages per log
			page.Ack(false)
		}
	}()

	// Wait until you're ready to finish, could be a signal handler here.
	time.Sleep(10 * time.Second)

	// Cancelling a consumer by name will finish the range and gracefully end the
	// goroutine
	err = c.Cancel("pager", false)
	if err != nil {
		log.Fatalf("basic.cancel: %v", err)
	}

	// deferred closing the Connection will also finish the consumer's ranges of
	// their delivery chans.  If you need every delivery to be processed, make
	// sure to wait for all consumers goroutines to finish before exiting your
	// process.
}

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

//
//type rmqSetupError struct {
//	stage string
//	msg   string
//}
//
//func NewSetupError(stage, msg string) {
//	return rmqSetupError{stage, msg}
//}
//
//func (r rmqSetupError) Error() string {
//	return fmt.Sprintf("%s: %s", r.stage, r.msg)
//}

type rmqLib struct {
	options options.Connection
	conn    *amqp.Connection
}

func New(opts options.Connection) *rmqLib {
	return &rmqLib{options: opts}
}

func (rc *rmqLib) Open() error {
	fmt.Printf("Opening connection with %#v\n", rc.options)

	conn, err := amqp.Dial(rc.options.ConnectionString)
	if err != nil {
		// un-recoverable
		log.Fatalf("amqp.connection.open: %s", err)
	}

	rc.conn = conn

	// Communicate everything over a channel (Socket abstraction)
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("amqp.channel.open: %s", err)
	}

	// need to make sure we have exchanges, queues and bindings
	err = rc.options.Exchange.Declare(channel)
	if err != nil {
		log.Fatalf("amqp.exchange.delcare: %s", err)
	}

	err = rc.options.Queue.Declare(channel)
	if err != nil {
		log.Fatalf("amqp.queue.declare: %s", err)
	}

	err = rc.options.Binding.Bind(channel, rc.options.Exchange, rc.options.Queue)
	if err != nil {
		log.Fatalf("amqp.binding.bind: %s", err)
	}

	return nil
}

func (rc *rmqLib) Close() {
	if rc.conn != nil {
		rc.conn.Close()
	}
}

func main() {

	exchange := options.Exchange{
		Name:    exchangeName,
		Kind:    "topic",
		Durable: true,
	}

	queue := options.Queue{
		Name:    queueName,
		Durable: true,
	}

	binding := options.Binding{
		RoutingKey: routingKey,
	}

	opts := options.Connection{
		Exchange:         exchange,
		Queue:            queue,
		Binding:          binding,
		Name:             "test-connection",
		ConnectionString: "amqp://guest:guest@localhost:5672/",
	}

	conn := New(opts)
	conn.Open()
	defer conn.Close()

	fmt.Printf("Done\n")

}
