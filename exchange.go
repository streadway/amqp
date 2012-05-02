package amqp

func (me *Exchange) Declare(lifetime Lifetime, exchangeType string, internal bool, noWait bool, arguments Table) error {
	return nil
}

func (me *Exchange) Exists() (bool, error) {
	return false, nil
}

func (me *Exchange) Delete(ifUnused bool, noWait bool) error { return nil }

func (me *Exchange) Bind(routingKey string, destinationExchange string, noWait bool, arguments Table) error {
	return nil
}

func (me *Exchange) Unbind(routingKey string, destinationExchange string, noWait bool, arguments Table) error {
	return nil
}

func (me *Exchange) Publish(routingKey string, mandatory bool, immediate bool, body []byte, properties Properties) error {
	return nil
}
