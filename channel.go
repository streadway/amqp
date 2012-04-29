package amqp

import (
	"amqp/wire"
	"fmt"
)

type Lifetime int

const (
	UntilDeleted         Lifetime = iota // durable
	UntilServerRestarted                 // not durable, not auto-delete
	UntilUnused                          // auto-delete
)

const (
	Direct  = "direct"
	Topic   = "topic"
	Fanout  = "fanout"
	Headers = "headers"
)

type QueueState struct {
	Name      string
	Consumers int
	Messages  int
}

// Represents an AMQP channel, used for concurrent, interleaved publishers and
// consumers on the same connection.
type Channel struct {
	framing   *Framing
	noWait    bool
	consumers map[string]chan Message
}

// Constructs and opens a new channel with the given framing rules
func newChannel(framing *Framing) (me *Channel, err error) {
	me = &Channel{
		framing:   framing,
		consumers: make(map[string]chan Message),
	}

	go me.handleAsync()

	return me, nil
}

// Can be one of the following methods:
func (me *Channel) handleAsync() {
	for {
		msg, ok := <-me.framing.async
		if !ok {
			// TODO close all consumer channels
			return
		}
		switch method := msg.Method.(type) {
		case wire.BasicDeliver:
			consumer, ok := me.consumers[method.ConsumerTag]
			if !ok {
				// TODO handle missing consumer
			} else {
				consumer <- msg
			}
		default:
			fmt.Println("Unhandled async method:", method)
		}
	}
}

//    channel             = open-channel *use-channel close-channel
//    open-channel        = C:OPEN S:OPEN-OK
//    use-channel         = C:FLOW S:FLOW-OK
//                        / S:FLOW C:FLOW-OK
//                        / functional-class
//    close-channel       = C:CLOSE S:CLOSE-OK
//                        / S:CLOSE C:CLOSE-OK
func (me *Channel) open() error {
	me.framing.SendMethod(wire.ChannelOpen{})

	switch me.framing.Recv().Method.(type) {
	case wire.ChannelOpenOk:
		return nil
	}

	// TODO handle channel open errors (like already opened on this ID)
	return ErrBadProtocol
}

func newQueueState(msg *wire.QueueDeclareOk) *QueueState {
	return &QueueState{
		Name:      msg.Queue,
		Consumers: int(msg.ConsumerCount),
		Messages:  int(msg.MessageCount),
	}
}

func (me *Channel) unhandled(msg wire.Method) error {
	// TODO CLOSE/CLOSE-OK/ERROR
	fmt.Println("UNHANDLED", msg)
	panic("UNHANDLED")
	return nil
}

func lifetimeParams(l Lifetime) (durable bool, autoDelete bool) {
	switch l {
	case UntilDeleted:
		return true, false
	case UntilServerRestarted:
		return false, false
	case UntilUnused:
		return false, true
	}

	panic("unreachable")
}

func (me *Channel) DeclareExchange(lifetime Lifetime, typ, name string) error {
	durable, autoDelete := lifetimeParams(lifetime)
	return me.CustomDeclareExchange(typ, name, durable, autoDelete, false, nil)
}

func (me *Channel) CustomDeclareExchange(typ string, name string, durable bool, autoDelete bool, internal bool, args Table) error {
	msg := wire.ExchangeDeclare{
		Exchange:   name,
		Type:       typ,
		Passive:    false,
		Durable:    durable,
		AutoDelete: autoDelete,
		Internal:   internal,
		NoWait:     me.noWait,
		Arguments:  wire.Table(args),
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.ExchangeDeclareOk:
			return nil
		default:
			return me.unhandled(res)
		}
		return ErrBadProtocol
	}

	return nil
}

func (me *Channel) InspectExchange(name string) (bool, error) {
	msg := wire.ExchangeDeclare{
		Exchange: name,
		Passive:  true,
	}

	me.framing.SendMethod(msg)

	// XXX maybe a select over the in and err channels is better?
	switch res := me.framing.Recv().Method.(type) {
	case wire.ExchangeDeclareOk:
		return true, nil
	case wire.ChannelClose:
		return false, nil
	default:
		return false, me.unhandled(res)
	}

	panic("unreachable")
}

func (me *Channel) DeleteExchange(name string, ifUnused bool) error {
	msg := wire.ExchangeDelete{
		Exchange: name,
		IfUnused: ifUnused,
		NoWait:   me.noWait,
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.ExchangeDeleteOk:
			return nil
		default:
			return me.unhandled(res)
		}
	}

	return nil
}

func (me *Channel) BindExchange(destination, source, routingKey string) error {
	return me.CustomBindExchange(destination, source, routingKey, nil)
}

func (me *Channel) CustomBindExchange(destination string, source string, routingKey string, arguments Table) error {
	msg := wire.ExchangeBind{
		Destination: destination,
		Source:      source,
		RoutingKey:  routingKey,
		NoWait:      me.noWait,
		Arguments:   wire.Table(arguments),
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.ExchangeBindOk:
			return nil
		default:
			return me.unhandled(res)
		}
	}

	return nil
}

func (me *Channel) UnbindExchange(destination string, source string, routingKey string) error {
	return me.CustomUnbindExchange(destination, source, routingKey, nil)
}

func (me *Channel) CustomUnbindExchange(destination string, source string, routingKey string, arguments Table) error {
	msg := wire.ExchangeUnbind{
		Destination: destination,
		Source:      source,
		RoutingKey:  routingKey,
		NoWait:      me.noWait,
		Arguments:   wire.Table(arguments),
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.ExchangeUnbindOk:
			return nil
		default:
			return me.unhandled(res)
		}
	}

	return nil
}

func (me *Channel) DeclareQueue(lifetime Lifetime, name string) (*QueueState, error) {
	durable, autoDelete := lifetimeParams(lifetime)
	return me.CustomDeclareQueue(name, durable, autoDelete, false, nil)
}

func (me *Channel) CustomDeclareQueue(name string, durable bool, autoDelete bool, exclusive bool, arguments Table) (*QueueState, error) {
	msg := wire.QueueDeclare{
		Queue:      name,
		Passive:    false,
		Durable:    durable,
		Exclusive:  exclusive,
		AutoDelete: autoDelete,
		NoWait:     me.noWait,
		Arguments:  wire.Table(arguments),
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.QueueDeclareOk:
			return newQueueState(&res), nil
		default:
			return nil, me.unhandled(res)
		}
	}

	return nil, nil
}

func (me *Channel) InspectQueue(name string) (*QueueState, error) {
	msg := wire.QueueDeclare{
		Queue:   name,
		Passive: true,
	}

	me.framing.SendMethod(msg)

	switch res := me.framing.Recv().Method.(type) {
	case wire.QueueDeclareOk:
		return newQueueState(&res), nil
	default:
		return nil, me.unhandled(res)
	}

	panic("unreachable")
}

func (me *Channel) BindQueue(exchange string, queue string, routingKey string) error {
	return me.CustomBindQueue(exchange, queue, routingKey, nil)
}

func (me *Channel) CustomBindQueue(exchange string, queue string, routingKey string, arguments Table) error {
	msg := wire.QueueBind{
		Queue:      queue,
		Exchange:   exchange,
		RoutingKey: routingKey,
		NoWait:     me.noWait,
		Arguments:  wire.Table(arguments),
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.QueueBindOk:
			return nil
		default:
			return me.unhandled(res)
		}
	}

	return nil
}

func (me *Channel) UnbindQueue(exchange string, queue string, routingKey string) error {
	return me.CustomUnbindQueue(exchange, queue, routingKey, nil)
}

func (me *Channel) CustomUnbindQueue(exchange string, queue string, routingKey string, arguments Table) error {
	msg := wire.QueueUnbind{
		Queue:      queue,
		Exchange:   exchange,
		RoutingKey: routingKey,
		Arguments:  wire.Table(arguments),
	}

	me.framing.SendMethod(msg)

	switch res := me.framing.Recv().Method.(type) {
	case wire.QueueUnbindOk:
		return nil
	default:
		return me.unhandled(res)
	}

	panic("unreachable")
}

func (me *Channel) PurgeQueue(name string) error {
	msg := wire.QueuePurge{
		Queue:  name,
		NoWait: me.noWait,
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.QueuePurgeOk:
			return nil
		default:
			return me.unhandled(res)
		}
	}

	return nil
}

func (me *Channel) DeleteQueue(name string, ifUnused bool, ifEmpty bool) error {
	msg := wire.QueueDelete{
		Queue:    name,
		IfUnused: ifUnused,
		IfEmpty:  ifEmpty,
		NoWait:   me.noWait,
	}

	me.framing.SendMethod(msg)

	if !msg.NoWait {
		switch res := me.framing.Recv().Method.(type) {
		case wire.QueueDeleteOk:
			return nil
		default:
			return me.unhandled(res)
		}
	}

	return nil
}

// Only applies to this Channel
func (me *Channel) Qos(prefetchMessageCount int, prefetchWindowByteSize int) error {
	msg := wire.BasicQos{
		PrefetchSize:  uint32(prefetchWindowByteSize),
		PrefetchCount: uint16(prefetchMessageCount),
		Global:        false, // connection global change from a channel message, durr...
	}

	me.framing.SendMethod(msg)

	switch res := me.framing.Recv().Method.(type) {
	case wire.BasicQosOk:
		return nil
	default:
		return me.unhandled(res)
	}

	panic("unreachable")
}

func (me *Channel) Publish(exchange string, routingKey string, body []byte, headers Table) {
	me.CustomPublish(exchange, routingKey, false, false, int64(len(body)), body, Properties(wire.ContentProperties{
		Headers:      wire.Table(headers),
		DeliveryMode: wire.TransientDelivery,
	}))
}

func (me *Channel) PublishPersistent(exchange string, routingKey string, body []byte, headers Table) {
	me.CustomPublish(exchange, routingKey, false, false, int64(len(body)), body, Properties(wire.ContentProperties{
		Headers:      wire.Table(headers),
		DeliveryMode: wire.PersistentDelivery,
	}))
}

func (me *Channel) CustomPublish(exchange string, routingKey string, mandatory bool, immediate bool, size int64, body []byte, properties Properties) {
	me.framing.Send(Message{
		Method: wire.BasicPublish{
			Exchange:   exchange,
			RoutingKey: routingKey,
			Mandatory:  mandatory,
			Immediate:  immediate,
		},
		Properties: wire.ContentProperties(properties),
		Body:       body,
	})
}

func (me *Channel) Consume(queue string) (chan Message, error) {
	msg := wire.BasicConsume{
		Queue:       queue,
		ConsumerTag: "",
		NoLocal:     false,
		NoAck:       false,
		Exclusive:   false,
		NoWait:      me.noWait,
		//Arguments
	}

	me.framing.SendMethod(msg)

	switch res := me.framing.Recv().Method.(type) {
	case wire.BasicConsumeOk:
		messages := make(chan Message)
		me.consumers[res.ConsumerTag] = messages
		return messages, nil
	default:
		return nil, me.unhandled(res)
	}

	panic("unreachable")
}

//func (me *Channel) Consume(buffersize) -> Consumer.(Cancel|messages -> Message(queue, exchange, key, tag, chan).(Reject|Ack)) {
//
//
//
//func (me *Channel) OnReturn(chan Return.Content*)) {
//}
//
//func (me *Channel) Get() -> (Message, ok) {
//}
