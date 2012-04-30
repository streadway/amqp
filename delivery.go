package amqp

// Contains a delivered message on a single channel as a result
// from a Consume or Get.  The methods invokable on this struct
// control the behavior of the channel
//
// The fields in Delivery relate to how this content arrived to this
// channel, and the fields in Properties relate to the content itself.
type Delivery struct {
	channel     *Channel
	method      *BasicDeliver
	Exchange    string
	RoutingKey  string
	Redelivered bool
	Properties  ContentProperties
	Body        []byte
}

// Cancels the asynchronous consumer that received this message, this is
// similar to "unsubscribing" from a queue.  Any messages in flight will still
// be delivered to the consumer until the consumer channel is closed.
func (me *Delivery) Cancel() error {
	return me.channel.BasicCancel(me.method.ConsumerTag)
}

// Acknowledges the client has recevied and processed this message on this
// channel This should be called if you have a begun a consumer with
// "ConsumeReliable"
func (me *Delivery) Ack() {
	me.channel.BasicAck(me.method.DeliveryTag, false)
}

// Acknowledges the client has recevied and processed this message on this channel
// and all previous messages delivered on this channel
// This should be called if you have a begun a consumer with "ConsumeReliable"
// and prefer to periodically batch your acknowledgements.
func (me *Delivery) AckAll() {
	me.channel.BasicAck(me.method.DeliveryTag, true)
}
