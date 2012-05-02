package amqp

func (me *Queue) Declare(lifetime Lifetime, exclusive bool, noWait bool, arguments Table) (QueueState, error) {
	return QueueState{}, nil
}

func (me *Queue) Inspect() (QueueState, error) {
	return QueueState{}, nil
}

func (me *Queue) Bind(routingKey string, sourceExchange string, noWait bool, arguments Table) error {
	return nil
}

func (me *Queue) Unbind(routingKey string, sourceExchange string, arguments Table) error {
	return nil
}

func (me *Queue) Purge(noWait bool) error { return nil }

func (me *Queue) Delete(ifUnused bool, ifEmpty bool, noWait bool) error {
	return nil
}

func (me *Queue) Consume(noAck bool, exclusive bool, noLocal bool, noWait bool, arguments Table) (chan Delivery, error) {
	return make(chan Delivery), nil
}

func (me *Queue) Publish(mandatory bool, immediate bool, body []byte, properties Properties) error {
	return nil
}
