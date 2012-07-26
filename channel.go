// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"reflect"
	"sync"
)

// 0      1         3             7                  size+7 size+8
// +------+---------+-------------+  +------------+  +-----------+
// | type | channel |     size    |  |  payload   |  | frame-end |
// +------+---------+-------------+  +------------+  +-----------+
//  octet   short         long         size octets       octet
const frameHeaderSize = 1 + 2 + 4 + 1

// Represents an AMQP channel. Used for concurrent, interleaved publishers and
// consumers on the same connection.
type Channel struct {
	// Mutex for notify listeners
	destructor sync.Once
	m          sync.Mutex

	connection *Connection

	rpc       chan message
	consumers consumers

	id    uint16
	state state

	// Writer lock for the notify slices
	notify sync.Mutex

	// Channel and Connection exceptions will be broadcast on these listeners.
	closes []chan *Error

	// Listeners for active=true flow control.  When true is sent to a listener,
	// publishing should pause until false is sent to listeners.
	flows []chan bool

	// Listeners for returned publishings for unroutable messages on mandatory
	// publishings or undeliverable messages on immediate publishings.
	returns []chan Return

	// Listeners for Acks/Nacks when the channel is in Confirm mode
	// the value is the sequentially increasing delivery tag
	// starting at 1 immediately after the Confirm
	acks  []chan uint64
	nacks []chan uint64

	// State machine that manages frame order, must only be mutated by the connection
	recv func(*Channel, frame) error

	// State that manages the send behavior after before and after shutdown, must
	// only be mutated in shutdown()
	send func(*Channel, message) error

	// Current state for frame re-assembly, only mutated from recv
	message messageWithContent
	header  *headerFrame
	body    []byte
}

// Constructs a new channel with the given framing rules
func newChannel(c *Connection, id uint16) (me *Channel, err error) {
	me = &Channel{
		connection: c,
		id:         id,
		rpc:        make(chan message),
		consumers:  makeConsumers(),
		recv:       (*Channel).recvMethod,
		send:       (*Channel).sendOpen,
	}

	return me, nil
}

func (me *Channel) shutdown(e *Error) {
	me.destructor.Do(func() {
		// Broadcast abnormal shutdown
		if e != nil {
			for _, c := range me.closes {
				c <- e
			}
		}

		delete(me.connection.channels, me.id)

		me.send = (*Channel).sendClosed

		close(me.rpc)

		me.consumers.closeAll()

		for _, c := range me.closes {
			close(c)
		}

		for _, c := range me.flows {
			close(c)
		}

		for _, c := range me.returns {
			close(c)
		}

		for _, c := range me.acks {
			close(c)
		}

		for _, c := range me.nacks {
			close(c)
		}
	})
}

func (me *Channel) open() (err error) {
	if err = me.call(&channelOpen{}, &channelOpenOk{}); err != nil {
		me.state = closed
	}

	return
}

// Performs a request/response call for when the message is not NoWait and is
// specified as Synchronous.
func (me *Channel) call(req message, res ...message) error {
	if err := me.send(me, req); err != nil {
		return err
	}

	if req.wait() {
		if msg, ok := <-me.rpc; ok {
			if closed, ok := msg.(*channelClose); ok {
				return newError(closed.ReplyCode, closed.ReplyText)
			} else {
				// Try to match one of the result types
				for _, try := range res {
					if reflect.TypeOf(msg) == reflect.TypeOf(try) {
						// *res = *msg
						vres := reflect.ValueOf(try).Elem()
						vmsg := reflect.ValueOf(msg).Elem()
						vres.Set(vmsg)
						return nil
					}
				}
				return ErrCommandInvalid
			}
		} else {
			// RPC channel has been closed, likely due to a hard error on the
			// Connection.  This indicates we have already been shutdown.
			return ErrClosed
		}
	}

	return nil
}

func (me *Channel) sendClosed(msg message) (err error) {
	// After a 'channel.close' is sent or received the only valid response is
	// channel.close-ok
	if _, ok := msg.(*channelCloseOk); ok {
		return me.connection.send(&methodFrame{
			ChannelId: me.id,
			Method:    msg,
		})
	}

	return ErrClosed
}

func (me *Channel) sendOpen(msg message) (err error) {
	me.m.Lock()
	defer me.m.Unlock()

	if content, ok := msg.(messageWithContent); ok {
		props, body := content.getContent()
		class, _ := content.id()
		size := me.connection.Config.MaxFrameSize - frameHeaderSize

		if err = me.connection.send(&methodFrame{
			ChannelId: me.id,
			Method:    content,
		}); err != nil {
			return
		}

		if err = me.connection.send(&headerFrame{
			ChannelId:  me.id,
			ClassId:    class,
			Size:       uint64(len(body)),
			Properties: props,
		}); err != nil {
			return
		}

		for i, j := 0, size; i < len(body); i, j = j, j+size {
			if j > len(body) {
				j = len(body)
			}

			if err = me.connection.send(&bodyFrame{
				ChannelId: me.id,
				Body:      body[i:j],
			}); err != nil {
				return
			}
		}
	} else {
		err = me.connection.send(&methodFrame{
			ChannelId: me.id,
			Method:    msg,
		})
	}

	return
}

// Eventually called via the state machine from the connection's reader
// goroutine, so assumes serialized access.
func (me *Channel) dispatch(msg message) {
	switch m := msg.(type) {
	case *channelClose:
		// Deliver this close message only when this is a synchronous response
		// (exception) to a waiting RPC call.  This is a touch racy when the call
		// goroutine is moving from send to receive.  If it doesn't meet the race,
		// the callee will get an ErrClosed instead of the text from this method.
		select {
		case me.rpc <- msg:
		default:
		}

		me.shutdown(newError(m.ReplyCode, m.ReplyText))
		me.send(me, &channelCloseOk{})

	case *channelFlow:
		for _, c := range me.flows {
			c <- m.Active
		}
		me.send(me, &channelFlowOk{Active: m.Active})

	case *basicReturn:
		ret := newReturn(*m)
		for _, c := range me.returns {
			c <- *ret
		}

	case *basicAck:
		for _, c := range me.acks {
			c <- m.DeliveryTag
		}

	case *basicNack:
		for _, c := range me.nacks {
			c <- m.DeliveryTag
		}

	case *basicDeliver:
		if me.state == open {
			me.consumers.send(m.ConsumerTag, newDelivery(me, m))
			// TODO log failed consumer and close channel, this can happen when
			// deliveries are in flight and a no-wait cancel has happened
		}

	default:
		me.rpc <- msg
	}
}

func (me *Channel) transition(f func(*Channel, frame) error) error {
	me.recv = f
	return nil
}

func (me *Channel) recvMethod(f frame) error {
	switch frame := f.(type) {
	case *methodFrame:
		if msg, ok := frame.Method.(messageWithContent); ok {
			me.body = make([]byte, 0)
			me.message = msg
			return me.transition((*Channel).recvHeader)
		}

		me.dispatch(frame.Method) // termination state
		return me.transition((*Channel).recvMethod)

	case *headerFrame:
		// drop
		return me.transition((*Channel).recvMethod)

	case *bodyFrame:
		// drop
		return me.transition((*Channel).recvMethod)
	}

	panic("unreachable")
}

func (me *Channel) recvHeader(f frame) error {
	switch frame := f.(type) {
	case *methodFrame:
		// interrupt content and handle method
		return me.recvMethod(f)

	case *headerFrame:
		// start collecting if we expect body frames
		me.header = frame

		if frame.Size == 0 {
			me.dispatch(me.message) // termination state
			return me.transition((*Channel).recvMethod)
		} else {
			return me.transition((*Channel).recvContent)
		}

	case *bodyFrame:
		// drop and reset
		return me.transition((*Channel).recvMethod)
	}

	panic("unreachable")
}

// state after method + header and before the length
// defined by the header has been reached
func (me *Channel) recvContent(f frame) error {
	switch frame := f.(type) {
	case *methodFrame:
		// interrupt content and handle method
		return me.recvMethod(f)

	case *headerFrame:
		// drop and reset
		return me.transition((*Channel).recvMethod)

	case *bodyFrame:
		me.body = append(me.body, frame.Body...)

		if uint64(len(me.body)) >= me.header.Size {
			me.message.setContent(me.header.Properties, me.body)
			me.dispatch(me.message) // termination state
			return me.transition((*Channel).recvMethod)
		}

		return me.transition((*Channel).recvContent)
	}

	panic("unreachable")
}

// Initiate a clean channel closure by sending a close message with the error
// code set to '200'
func (me *Channel) Close() error {
	err := me.call(&channelClose{ReplyCode: ReplySuccess}, &channelCloseOk{})
	me.shutdown(nil)
	return err
}

// Add a chan for when the server sends a channel or connection exception in
// the form of a connection.close or channel.close method.  Connection
// exceptions will be broadcast to all open channels and all channels will be
// closed, where channel exceptions will only be broadcast to listeners to this
// channel.
//
// The chan provided will be closed when the Channel is closed and on a
// graceful close, no error will be sent.
//
func (me *Channel) NotifyClose(c chan *Error) chan *Error {
	me.m.Lock()
	defer me.m.Unlock()
	me.closes = append(me.closes, c)
	return c
}

// Listens for basic.flow methods sent by the server.  When `true` is sent on
// one of the listener channels, all publishes should pause until a `false` is
// sent.
//
// Asks the producer to pause or restart the flow of content data sent by a
// consumer. This is a simple flow-control mechanism that a server can use to
// avoid overflowing its queues or otherwise finding itself receiving more
// messages than it can process. Note that this method is not intended for
// window control. It does not affect contents returned by basic.get-ok
// methods.
// 
// When a new channel is opened, it is active (flow is active). Some
// applications assume that channels are inactive until started. To emulate
// this behaviour a client MAY open the channel, then pause it.
// 
// Publishers should respond to a flow messages as rapidly as possible and the
// server may disconnect over producing channels that do not respect these
// messages.
//
// basic.flow-ok methods will always be returned to the server regardless of
// the number of listeners there are.
// 
// To control the flow of deliveries from the server.  Use the Channel.Flow()
// method instead.
// 
// Note: RabbitMQ will rather use TCP pushback on the network connection instead of
// sending basic.flow.  This means that if a single channel is producing too
// much on the same connection, all channels using that connection will suffer,
// including acknowlegements from deliveries.
//
func (me *Channel) NotifyFlow(c chan bool) chan bool {
	me.m.Lock()
	defer me.m.Unlock()
	me.flows = append(me.flows, c)
	return c
}

// Listens for basic.return methods.  These can be sent from the server when a
// publish with the `immediate` flag lands on a queue that does not have
// any ready consumers or when a publish with the `mandatory` flag does not
// have a route.
//
// A return struct has a copy of the Publishing along with some error
// information about why the publishing failed.
func (me *Channel) NotifyReturn(c chan Return) chan Return {
	me.m.Lock()
	defer me.m.Unlock()
	me.returns = append(me.returns, c)
	return c
}

// Intended for reliable publishing.  Add a listener for basic.ack and
// basic.nack messages.  These will be sent by the server for every publish
// after `channel.Confirm(...)` has been called.  The value sent on these
// channels are the sequence number of the publishing.  It is up to client of
// this channel to maintain the sequence number and handle resends.
//
// The order of acknowledgements is not bound to the order of deliveries.
//
// It's advisable to wait for all acks or nacks to arrive before closing the
// channel on completion.
func (me *Channel) NotifyConfirm(ack, nack chan uint64) (chan uint64, chan uint64) {
	me.m.Lock()
	defer me.m.Unlock()
	me.acks = append(me.acks, ack)
	me.nacks = append(me.nacks, nack)
	return ack, nack
}

func (me *Channel) Qos(prefetchCount uint16, prefetchSize uint32, global bool) (err error) {
	return me.call(
		&basicQos{
			PrefetchSize:  prefetchSize,
			PrefetchCount: prefetchCount,
			Global:        global,
		},
		&basicQosOk{},
	)
}

func (me *Channel) Cancel(consumerTag string, noWait bool) (err error) {
	req := &basicCancel{
		ConsumerTag: consumerTag,
		NoWait:      noWait,
	}
	res := &basicCancelOk{}

	if err = me.call(req, res); err != nil {
		return
	}

	if req.wait() {
		me.consumers.close(res.ConsumerTag)
	} else {
		// Potentially could drop deliveries in flight
		me.consumers.close(consumerTag)
	}

	return
}

func (me *Channel) QueueDeclare(name string, lifetime Lifetime, exclusive bool, noWait bool, arguments Table) (state QueueState, err error) {
	req := &queueDeclare{
		Queue:      name,
		Passive:    false,
		Durable:    lifetime.durable(),
		AutoDelete: lifetime.autoDelete(),
		Exclusive:  exclusive,
		NoWait:     noWait,
		Arguments:  arguments,
	}
	res := &queueDeclareOk{}

	if err = me.call(req, res); err != nil {
		return
	}

	if req.wait() {
		state = QueueState{
			Declared:      (res.Queue == name),
			MessageCount:  int(res.MessageCount),
			ConsumerCount: int(res.ConsumerCount),
		}
	} else {
		state = QueueState{Declared: true}
	}

	return
}

func (me *Channel) QueueInspect(name string) (QueueState, error) {
	req := &queueDeclare{
		Queue:   name,
		Passive: true,
	}
	res := &queueDeclareOk{}

	err := me.call(req, res)

	state := QueueState{
		Declared:      (res.Queue == name),
		MessageCount:  int(res.MessageCount),
		ConsumerCount: int(res.ConsumerCount),
	}

	return state, err
}

func (me *Channel) QueueBind(name string, routingKey string, sourceExchange string, noWait bool, arguments Table) error {
	return me.call(
		&queueBind{
			Queue:      name,
			Exchange:   sourceExchange,
			RoutingKey: routingKey,
			NoWait:     noWait,
			Arguments:  arguments,
		},
		&queueBindOk{},
	)
}

func (me *Channel) QueueUnbind(name string, routingKey string, sourceExchange string, arguments Table) error {
	return me.call(
		&queueUnbind{
			Queue:      name,
			Exchange:   sourceExchange,
			RoutingKey: routingKey,
			Arguments:  arguments,
		},
		&queueUnbindOk{},
	)
}

func (me *Channel) QueuePurge(name string, noWait bool) (messageCount int, err error) {
	req := &queuePurge{
		Queue:  name,
		NoWait: noWait,
	}
	res := &queuePurgeOk{}

	err = me.call(req, res)
	messageCount = int(res.MessageCount)

	return
}

func (me *Channel) QueueDelete(name string, ifUnused bool, ifEmpty bool, noWait bool) (err error) {
	return me.call(
		&queueDelete{
			Queue:    name,
			IfUnused: ifUnused,
			IfEmpty:  ifEmpty,
			NoWait:   noWait,
		},
		&queueDeleteOk{},
	)
}

// Will deliver on the chan passed or this will make a new chan. The
// delivery chan, either provided or created, will be returned.
//
// If a consumerTag is not provided, the server will generate one. If you do
// include noWait, a random tag will be generated by the client.
func (me *Channel) Consume(queueName string, noAck bool, exclusive bool, noLocal bool, noWait bool, consumerTag string, arguments Table, deliveries chan Delivery) (ch chan Delivery, err error) {
	ch = deliveries
	if ch == nil {
		ch = make(chan Delivery)
	}

	// when we won't wait for the server, add the consumer channel now
	if noWait {
		if consumerTag == "" {
			consumerTag = randomConsumerTag()
		}
		me.consumers.add(consumerTag, ch)
	}

	req := &basicConsume{
		Queue:       queueName,
		ConsumerTag: consumerTag,
		NoLocal:     noLocal,
		NoAck:       noAck,
		Exclusive:   exclusive,
		NoWait:      noWait,
		Arguments:   arguments,
	}
	res := &basicConsumeOk{}

	err = me.call(req, res)

	if !noWait {
		me.consumers.add(res.ConsumerTag, ch)
	}

	return
}

/*
Declares an exchange on the server. If the exchange does not already exist,
the server will create it.  If the exchange exists, the server verifies
that it is of the provided type and lifetime.

Errors returned from this method will close the channel.

	name string

Exchange names starting with "amq." are reserved for pre-declared and
standardized exchanges. The client MAY declare an exchange starting with
"amq." if the passive option is set, or the exchange already exists.

The exchange name can consists of a non-empty sequence of letters, digits, hyphen, underscore, period, or colon.

	exchangeType string

Each exchange belongs to one of a set of exchange types implemented by the
server. The exchange types define the functionality of the exchange - i.e. how
messages are routed through it. It is not valid or meaningful to attempt to
change the type of an existing exchange.

Exchanges cannot be redeclared with different types.  The client MUST not
attempt to redeclare an existing exchange with a different type than used
in the original Exchange.Declare method.

	lifetime Lifetime

One of the following values

		UntilDeleted
		UntilUnused
		UntilServerRestarted

The lifetime parameter contains the 'durable' and 'auto-delete' flags for this declaration and can be one of:

UntilDeleted - declares the exchange as durable and not auto-deleted will
survive server restarts and remain when there are no remaining bindings.  This
is the lifetime for long-lived exchange configurations like stable routes and
default exchanges.

UntilUnused - declares the exchange as non-durable and auto-deleted which will
not be redeclared on server restart and will be deleted when there are no
remaining bindings.  This lifetime is useful for temporary topologies that
should not pollute the virtual host on failure or completion.

UntilServerRestarted - declares the exchange as non-durable and not
auto-deleted which will remain as long as the server is running including when
there are no remaining bindings.  This is useful for temporary topologies that
may have long delays between bindings.

Note: RabbitMQ declares the default exchange types like 'amq.fanout' with the
equivalent Lifetime of UntilDeleted

	internal bool

When true, the exchange may not be used directly by publishers, but only when
bound to other exchanges. Internal exchanges are used to construct wiring that
is not visible to applications.

	noWait bool

Declare without waiting for a confirmation from the server.  The channel may
be closed as a result.  Add a NotifyClose listener to respond to any
exceptions.

	arguments amqp.Table

An amqp.Table of arguments that are specific to the server's implementation of
the exchange.  This can be nil when no arguments are required to be sent.
*/
func (me *Channel) ExchangeDeclare(name string, lifetime Lifetime, exchangeType string, internal, noWait bool, arguments Table) error {
	return me.call(
		&exchangeDeclare{
			Exchange:   name,
			Type:       exchangeType,
			Passive:    false,
			Durable:    lifetime.durable(),
			AutoDelete: lifetime.autoDelete(),
			Internal:   internal,
			NoWait:     noWait,
			Arguments:  arguments,
		},
		&exchangeDeclareOk{},
	)
}

/*
Deletes an exchange. When an exchange is deleted all queue bindings on the
exchange are also deleted.

	name string

The name of the exchange to delete.  If this exchange does not exist, the
channel will be closed with an error.

	ifUnused bool

When true, the server will only delete the exchange if it has no queue
bindings.  If the exchange has queue bindings the server does not delete it
but close the channel with an exception instead.

	noWait bool

When true, do not wait for a server confirmation that the exchange has been
deleted.  Failing to delete the channel could close the channel.  Add a
NotifyClose listener to respond to these channel exceptions.
*/
func (me *Channel) ExchangeDelete(name string, ifUnused, noWait bool) error {
	return me.call(
		&exchangeDelete{
			Exchange: name,
			IfUnused: ifUnused,
			NoWait:   noWait,
		},
		&exchangeDeleteOk{},
	)
}

/*
Binds an exchange to another exchange to create inter-exchange routing
topologies.  This can decouple the private topology from the public publishing
exchanges.

Binding two exchanges with identical arguments will not create duplicate
bindings.  The duplicates will be ignored by the server.

Binding one exchange to another with multiple bindings will only deliver a
message once.  For example if you bind your exchange to `amq.fanout` with two
different binding keys, only a single message will be delivered to your
exchange even though multiple bindings will match.

	name string

The source exchange of the binding.  Content published to this exchange may be
delivered to the `destinationExchange` when the routing key matches.  This
exchange must be declared, and if blank, will be interpreted as the default
exchange.

  routingKey string

For exchanges that use a routing key, like amq.direct, forward matched
deliveries from to the `destinationExchange`.

	destinationExchange string

The receiver exchange of messages that match the routingKey from the source
exchange.  This exchange must be declared and if blank, will be interpreted as
the default exchange.

	noWait bool

Do not wait for the server to confirm the binding.  If any error occurs the
channel will be closed.  Add a listener to NotifyClose to handle these errors.

	arguments amqp.Table

Arguments that are specific to the type of exchanges bound.

*/
func (me *Channel) ExchangeBind(name, routingKey, destinationExchange string, noWait bool, arguments Table) error {
	return me.call(
		&exchangeBind{
			Destination: destinationExchange,
			Source:      name,
			RoutingKey:  routingKey,
			NoWait:      noWait,
			Arguments:   arguments,
		},
		&exchangeBindOk{},
	)
}

/*
Unbinds one exchange from another that matches the same routingKey.

This is the inverse of ExchangeBind.

	name string

The name of the source exchange that matches the same parameter in
ExchangeBind.

	routingKey string

The routing key used to identify the binding between source and destination
exchanges.  Bindings that do not match this key will remain untouched.

	destinationExchange string

The destination exchange that will no longer receive deliveries that matched
the routing key delivered to the source exchange.

	noWait bool

Do not wait for the server to confirm the deletion of the binding.  If any
error occurs the channel will be closed.  Add a listener to NotifyClose to
handle these errors.

	arguments amqp.Table

Arguments that are specific to the type of exchanges bound.  These must match the same arguments specified in ExchangeBind to identify the binding.
*/
func (me *Channel) ExchangeUnbind(name, routingKey, destinationExchange string, noWait bool, arguments Table) error {
	return me.call(
		&exchangeUnbind{
			Destination: destinationExchange,
			Source:      name,
			RoutingKey:  routingKey,
			NoWait:      noWait,
			Arguments:   arguments,
		},
		&exchangeUnbindOk{},
	)
}

// Publishes a message to an exchange.
func (me *Channel) Publish(exchangeName string, routingKey string, mandatory bool, immediate bool, msg Publishing) (err error) {
	return me.send(me, &basicPublish{
		Exchange:   exchangeName,
		RoutingKey: routingKey,
		Mandatory:  mandatory,
		Immediate:  immediate,
		Body:       msg.Body,
		Properties: properties{
			Headers:         msg.Headers,
			ContentType:     msg.ContentType,
			ContentEncoding: msg.ContentEncoding,
			DeliveryMode:    msg.DeliveryMode,
			Priority:        msg.Priority,
			CorrelationId:   msg.CorrelationId,
			ReplyTo:         msg.ReplyTo,
			Expiration:      msg.Expiration,
			MessageId:       msg.MessageId,
			Timestamp:       msg.Timestamp,
			Type:            msg.Type,
			UserId:          msg.UserId,
			AppId:           msg.AppId,
		},
	})
}

// Synchronously fetch a single message from a queue.  In almost all cases,
// using `Consume` will be preferred.
//
// When a message is available on the queue, the 'ok bool' return parameter will be true.
func (me *Channel) Get(queueName string, noAck bool) (*Delivery, bool, error) {
	req := &basicGet{Queue: queueName, NoAck: noAck}
	res := &basicGetOk{}
	empty := &basicGetEmpty{}

	if err := me.call(req, res, empty); err != nil {
		return nil, false, err
	}

	if res.DeliveryTag > 0 {
		return newDelivery(me, res), true, nil
	}

	return nil, false, nil
}

func (me *Channel) TxSelect() (err error) {
	return me.call(
		&txSelect{},
		&txSelectOk{},
	)
}

func (me *Channel) TxCommit() (err error) {
	return me.call(
		&txCommit{},
		&txCommitOk{},
	)
}

func (me *Channel) TxRollback() (err error) {
	return me.call(
		&txRollback{},
		&txRollbackOk{},
	)
}

// On `true`, requests the server to pause delivery.  On `false` requests the
// server to resume delivery.
func (me *Channel) Flow(active bool) (err error) {
	return me.call(
		&channelFlow{Active: active},
		&channelFlowOk{},
	)
}

// Put this channel into confirm mode.  The server will then send either a
// basic.ack or basic.nack message with the deliver tag set to a 1 based index
// corresponding to every publishing received after the this method returns.
//
// The order of acknowledgements is not bound to the order of deliveries.
func (me *Channel) Confirm(noWait bool) (err error) {
	return me.call(
		&confirmSelect{Nowait: noWait},
		&confirmSelectOk{},
	)
}

//TODO func (me *Channel) Recover(requeue bool) error                                 { return nil }
