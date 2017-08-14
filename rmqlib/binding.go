package rmqlib

import "github.com/streadway/amqp"

type Binding struct {
	RoutingKey string
}

func (b Binding) Bind(channel *amqp.Channel, exchange Exchange, queue Queue) error {
	return channel.QueueBind(queue.Name, b.RoutingKey, exchange.Name, exchange.NoWait, exchange.Args)
}
