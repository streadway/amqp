// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"reflect"
)

// 0      1         3             7                  size+7 size+8
// +------+---------+-------------+  +------------+  +-----------+
// | type | channel |     size    |  |  payload   |  | frame-end |
// +------+---------+-------------+  +------------+  +-----------+
//  octet   short         long         size octets       octet
const frameHeaderOverhead = 1 + 2 + 4 + 1

// Represents an AMQP channel. Used for concurrent, interleaved publishers and
// consumers on the same connection.
type Channel struct {
	Closed *Closed

	connection *Connection

	rpc       chan message
	consumers consumers

	id    uint16
	state state

	// State machine that manages frame order
	recv func(*Channel, frame) error

	// Current state for frame re-assembly
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
	}

	return me, nil
}

func (me *Channel) shutdown() {
	me.state = closing

	delete(me.connection.channels, me.id)
	close(me.rpc)
	me.consumers.closeAll()
	me.state = closed
}

func (me *Channel) open() (err error) {
	req := &channelOpen{}
	res := &channelOpenOk{}

	me.state = handshaking

	if err = me.call(req, res); err != nil {
		return
	}

	me.state = open

	return
}

// Performs a request/response call for when the message is not NoWait and is
// specified as Synchronous.
func (me *Channel) call(req, res message) (err error) {
	if err = me.send(req); err != nil {
		return
	}

	if req.wait() {
		msg, ok := <-me.rpc
		if !ok {
			me.shutdown()
			err = ErrAlreadyClosed
		} else if reflect.TypeOf(msg) == reflect.TypeOf(res) {
			// *res = *msg
			vres := reflect.ValueOf(res).Elem()
			vmsg := reflect.ValueOf(msg).Elem()
			vres.Set(vmsg)
		} else {
			// Unexpected message in the response
			// TODO close this channel with 503
			err = ErrBadProtocol
		}
	}

	return
}

func (me *Channel) send(msg message) (err error) {
	if me.state != open && me.state != handshaking && me.state != closing {
		return ErrAlreadyClosed
	}

	if content, ok := msg.(messageWithContent); ok {
		props, body := content.getContent()
		class, _ := content.id()
		size := me.connection.Config.MaxFrameSize - frameHeaderOverhead

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

// Initiate a clean channel closure by sending a close message with the error code set to '200'
func (me *Channel) Close() (err error) {
	if me.state != open {
		return ErrAlreadyClosed
	}

	err = me.call(&channelClose{ReplyCode: ReplySuccess}, &channelCloseOk{})
	me.shutdown()

	return
}

func (me *Channel) newDelivery(msg messageWithContent) *Delivery {
	props, body := msg.getContent()

	delivery := Delivery{
		channel: me,

		Headers:         props.Headers,
		ContentType:     props.ContentType,
		ContentEncoding: props.ContentEncoding,
		DeliveryMode:    props.DeliveryMode,
		Priority:        props.Priority,
		CorrelationId:   props.CorrelationId,
		ReplyTo:         props.ReplyTo,
		Expiration:      props.Expiration,
		MessageId:       props.MessageId,
		Timestamp:       props.Timestamp,
		Type:            props.Type,
		UserId:          props.UserId,
		AppId:           props.AppId,

		Body: body,
	}

	// Properties for the delivery types
	switch m := msg.(type) {
	case *basicDeliver:
		delivery.ConsumerTag = m.ConsumerTag
		delivery.DeliveryTag = m.DeliveryTag
		delivery.Redelivered = m.Redelivered
		delivery.Exchange = m.Exchange
		delivery.RoutingKey = m.RoutingKey

	case *basicGetOk:
		delivery.MessageCount = m.MessageCount
		delivery.DeliveryTag = m.DeliveryTag
		delivery.Redelivered = m.Redelivered
		delivery.Exchange = m.Exchange
		delivery.RoutingKey = m.RoutingKey
	}

	return &delivery
}

// Eventually called via the state machine from the connection's reader
// goroutine, so assumes serialized access.
func (me *Channel) dispatch(msg message) {
	switch m := msg.(type) {
	case *channelClose:
		if me.state != closed {
			me.rpc <- msg
		}
		me.send(&channelCloseOk{})
		if me.state == open {
			me.shutdown()
		}

	case *channelFlow:
		// unhandled

	case *basicDeliver:
		if me.state == open {
			me.consumers.send(m.ConsumerTag, me.newDelivery(m))
			// TODO log failed consumer and close channel, this can happen when
			// deliveries are in flight and a no-wait cancel has happened
		}

	default:
		// TODO only deliver on the RPC channel if there the message is
		// synchronouse (wait()==true)
		if me.state != closed {
			me.rpc <- msg
		}
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

func (me *Channel) QueueInspect(name string) (state QueueState, err error) {
	req := &queueDeclare{
		Queue:   name,
		Passive: true,
	}
	res := &queueDeclareOk{}

	err = me.call(req, res)

	state = QueueState{
		Declared:      (res.Queue == name),
		MessageCount:  int(res.MessageCount),
		ConsumerCount: int(res.ConsumerCount),
	}

	return
}

func (me *Channel) QueueBind(name string, routingKey string, sourceExchange string, noWait bool, arguments Table) (err error) {
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

func (me *Channel) QueueUnbind(name string, routingKey string, sourceExchange string, arguments Table) (err error) {
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

func (me *Channel) ExchangeDeclare(name string, lifetime Lifetime, exchangeType string, internal bool, noWait bool, arguments Table) (err error) {
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

func (me *Channel) ExchangeDelete(name string, ifUnused bool, noWait bool) (err error) {
	return me.call(
		&exchangeDelete{
			Exchange: name,
			IfUnused: ifUnused,
			NoWait:   noWait,
		},
		&exchangeDeleteOk{},
	)
}

func (me *Channel) ExchangeBind(name string, routingKey string, destinationExchange string, noWait bool, arguments Table) (err error) {
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

func (me *Channel) ExchangeUnbind(name string, routingKey string, destinationExchange string, noWait bool, arguments Table) (err error) {
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
	return me.send(&basicPublish{
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
func (me *Channel) Get(queueName string, noAck bool) (msg *Delivery, ok bool, err error) {
	// Special case where 'call' cannot be used because of the variant return type
	// *** needs to mirror any changes to .call
	if err = me.send(&basicGet{Queue: queueName, NoAck: noAck}); err != nil {
		return
	}

	switch m := (<-me.rpc).(type) {
	case *basicGetOk:
		return me.newDelivery(m), true, nil
	case *basicGetEmpty:
		return nil, false, nil
	case nil:
		me.shutdown()
		return nil, false, ErrAlreadyClosed
	}

	return nil, false, ErrBadProtocol
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

//TODO func (me *Channel) Recover(requeue bool) error                                 { return nil }

//TODO func (me *Channel) Confirm(noWait bool) error                                  { return nil }

//TODO func (me *Channel) Flow(active bool) error { return nil }
