package amqp

func (me *Delivery) Ack() error                { return nil }
func (me *Delivery) AckMultiple() error        { return nil }
func (me *Delivery) Reject(requeue bool) error { return nil }
