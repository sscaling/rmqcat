package options

type Connection struct {
	Exchange
	Queue
	Binding
	Name             string
	ConnectionString string
}
