package rmqlib

type Connection struct {
	Exchange
	Queue
	Binding
	Name             string
	ConnectionString string
}
