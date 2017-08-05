package options

import "github.com/streadway/amqp"

type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Table      amqp.Table
}
