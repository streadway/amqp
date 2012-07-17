// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"sync"
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

	rpc chan message

	mutex     sync.Mutex
	consumers map[string]chan Delivery

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
		consumers:  make(map[string]chan Delivery),
		recv:       (*Channel).recvMethod,
	}

	return me, nil
}

func (me *Channel) addConsumer(tag string, ch chan Delivery) string {
	if tag == "" {
		tag = randomTag()
	}

	me.mutex.Lock()
	defer me.mutex.Unlock()
	me.consumers[tag] = ch

	return tag
}

func (me *Channel) shutdown() {
	me.state = closing

	delete(me.connection.channels, me.id)

	me.mutex.Lock()
	defer me.mutex.Unlock()
	for tag, ch := range me.consumers {
		delete(me.consumers, tag)
		me.Cancel(tag, false)
		close(ch)
	}

	close(me.rpc)
	me.state = closed
}

func (me *Channel) open() (err error) {
	me.state = handshaking

	if err = me.send(&channelOpen{}); err != nil {
		return
	}

	switch (<-me.rpc).(type) {
	case *channelOpenOk:
		me.state = open
		return
	default:
		return ErrBadProtocol
	}

	panic("unreachable")
}

func (me *Channel) send(msg message) (err error) {
	if me.state != open && me.state != handshaking {
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

	me.send(&channelClose{ReplyCode: ReplySuccess})

	me.state = closing

	switch (<-me.rpc).(type) {
	case *channelCloseOk:
		return
	case nil:
		return
	default:
		return ErrBadProtocol
	}

	panic("unreachable")
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

	// Properties for the delivery
	if deliver, ok := msg.(*basicDeliver); ok {
		delivery.ConsumerTag = deliver.ConsumerTag
		delivery.DeliveryTag = deliver.DeliveryTag
		delivery.Redelivered = deliver.Redelivered
		delivery.Exchange = deliver.Exchange
		delivery.RoutingKey = deliver.RoutingKey

	}

	if get, ok := msg.(*basicGetOk); ok {
		delivery.MessageCount = get.MessageCount
		delivery.DeliveryTag = get.DeliveryTag
		delivery.Redelivered = get.Redelivered
		delivery.Exchange = get.Exchange
		delivery.RoutingKey = get.RoutingKey
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
			me.Close()
		}

	case *channelFlow:
		// unhandled

	case *basicDeliver:
		if me.state == open {
			me.mutex.Lock()
			defer me.mutex.Unlock()
			if c, ok := me.consumers[m.ConsumerTag]; ok {
				c <- *me.newDelivery(m)
			}
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
	if err = me.send(&basicQos{
		PrefetchSize:  prefetchSize,
		PrefetchCount: prefetchCount,
		Global:        global,
	}); err != nil {
		return
	}

	switch (<-me.rpc).(type) {
	case *basicQosOk:
		return
	default:
		return ErrBadProtocol
	}

	panic("unreachable")
}

func (me *Channel) Cancel(consumerTag string, noWait bool) (err error) {
	if err = me.send(&basicCancel{
		ConsumerTag: consumerTag,
		NoWait:      noWait,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.rpc).(type) {
		case *basicCancelOk:
			return
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Channel) QueueDeclare(name string, lifetime Lifetime, exclusive bool, noWait bool, arguments Table) (state QueueState, err error) {
	if err = me.send(&queueDeclare{
		Queue:      name,
		Passive:    false,
		Durable:    lifetime.durable(),
		AutoDelete: lifetime.autoDelete(),
		Exclusive:  exclusive,
		NoWait:     noWait,
		Arguments:  arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch ok := (<-me.rpc).(type) {
		case *queueDeclareOk:
			return QueueState{
				Declared:      true,
				MessageCount:  int(ok.MessageCount),
				ConsumerCount: int(ok.ConsumerCount),
			}, nil
		case nil:
			return QueueState{Declared: false}, me.Close()
		default:
			return QueueState{Declared: false}, ErrBadProtocol
		}
	}

	return QueueState{Declared: true}, nil
}

func (me *Channel) QueueInspect(name string) (state QueueState, err error) {
	if err = me.send(&queueDeclare{
		Queue:   name,
		Passive: true,
	}); err != nil {
		return
	}

	switch ok := (<-me.rpc).(type) {
	case *queueDeclareOk:
		return QueueState{
			Declared:      true,
			MessageCount:  int(ok.MessageCount),
			ConsumerCount: int(ok.ConsumerCount),
		}, nil
		// TODO: handle when it's not declared?
	case nil:
		return QueueState{Declared: false}, me.Close()
	default:
		return QueueState{Declared: false}, ErrBadProtocol
	}

	panic("unreachable")
}

func (me *Channel) QueueBind(name string, routingKey string, sourceExchange string, noWait bool, arguments Table) (err error) {
	if err = me.send(&queueBind{
		Queue:      name,
		Exchange:   sourceExchange,
		RoutingKey: routingKey,
		NoWait:     noWait,
		Arguments:  arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.rpc).(type) {
		case *queueBindOk:
			return
		case nil:
			return me.Close()
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Channel) QueueUnbind(name string, routingKey string, sourceExchange string, arguments Table) (err error) {
	if err = me.send(&queueUnbind{
		Queue:      name,
		Exchange:   sourceExchange,
		RoutingKey: routingKey,
		Arguments:  arguments,
	}); err != nil {
		return
	}

	switch (<-me.rpc).(type) {
	case *queueUnbindOk:
		return
	case nil:
		return me.Close()
	default:
		return ErrBadProtocol
	}

	panic("unreachable")
}

func (me *Channel) QueuePurge(name string, noWait bool) (messageCount int, err error) {
	if err = me.send(&queuePurge{
		Queue:  name,
		NoWait: noWait,
	}); err != nil {
		return
	}

	if !noWait {
		switch ok := (<-me.rpc).(type) {
		case *queuePurgeOk:
			return int(ok.MessageCount), nil
		case nil:
			return 0, me.Close()
		default:
			return 0, ErrBadProtocol
		}
	}

	return
}

func (me *Channel) QueueDelete(name string, ifUnused bool, ifEmpty bool, noWait bool) (err error) {
	if err = me.send(&queueDelete{
		Queue:    name,
		IfUnused: ifUnused,
		IfEmpty:  ifEmpty,
		NoWait:   noWait,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.rpc).(type) {
		case *queueDeleteOk:
			return
		case nil:
			return me.Close()
		default:
			return ErrBadProtocol
		}
	}

	return
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
		consumerTag = me.addConsumer(consumerTag, ch)
	}

	if err = me.send(&basicConsume{
		Queue:       queueName,
		ConsumerTag: consumerTag,
		NoLocal:     noLocal,
		NoAck:       noAck,
		Exclusive:   exclusive,
		NoWait:      noWait,
		Arguments:   arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch ok := (<-me.rpc).(type) {
		case *basicConsumeOk:
			me.addConsumer(ok.ConsumerTag, ch)
			return
		case nil:
			return ch, me.Close()
		default:
			return ch, ErrBadProtocol
		}
	}

	return
}

func (me *Channel) ExchangeDeclare(name string, lifetime Lifetime, exchangeType string, internal bool, noWait bool, arguments Table) (err error) {
	if err = me.send(&exchangeDeclare{
		Exchange:   name,
		Type:       exchangeType,
		Passive:    false,
		Durable:    lifetime.durable(),
		AutoDelete: lifetime.autoDelete(),
		Internal:   internal,
		NoWait:     noWait,
		Arguments:  arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.rpc).(type) {
		case *exchangeDeclareOk:
			return
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Channel) ExchangeDelete(name string, ifUnused bool, noWait bool) (err error) {
	if err = me.send(&exchangeDelete{
		Exchange: name,
		IfUnused: ifUnused,
		NoWait:   noWait,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.rpc).(type) {
		case *exchangeDeleteOk:
			return
		case nil:
			return me.Close()
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Channel) ExchangeBind(name string, routingKey string, destinationExchange string, noWait bool, arguments Table) (err error) {
	if err = me.send(&exchangeBind{
		Destination: destinationExchange,
		Source:      name,
		RoutingKey:  routingKey,
		NoWait:      noWait,
		Arguments:   arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.rpc).(type) {
		case *exchangeBindOk:
			return
		case nil:
			return me.Close()
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Channel) ExchangeUnbind(name string, routingKey string, destinationExchange string, noWait bool, arguments Table) (err error) {
	if err = me.send(&exchangeUnbind{
		Destination: destinationExchange,
		Source:      name,
		RoutingKey:  routingKey,
		NoWait:      noWait,
		Arguments:   arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.rpc).(type) {
		case *exchangeUnbindOk:
			return
		case nil:
			return me.Close()
		default:
			return ErrBadProtocol
		}
	}
	return
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
	if err = me.send(&basicGet{
		Queue: queueName,
		NoAck: noAck,
	}); err != nil {
		return
	}

	switch m := (<-me.rpc).(type) {
	case *basicGetOk:
		return me.newDelivery(m), true, nil
	case *basicGetEmpty:
		return nil, false, nil
	case nil:
		return nil, false, me.Close()
	}

	return nil, false, ErrBadProtocol
}

// RabbitMQ extension - Negatively acknowledge the delivery of message(s)
// identified by the deliveryTag.  When multiple, nack messages up to and
// including delivered messages up until the deliveryTag.
//
// This method must not be used to select or requeue messages the client wishes
// not to handle.
func (me *Channel) Nack(deliveryTag uint64, multiple bool, requeue bool) (err error) {
	return me.send(&basicNack{
		DeliveryTag: deliveryTag,
		Multiple:    multiple,
		Requeue:     requeue,
	})
}

//TODO func (me *Channel) Recover(requeue bool) error                                 { return nil }

//TODO func (me *Channel) Confirm(noWait bool) error                                  { return nil }

//TODO func (me *Channel) TxSelect() error   { return nil }
//TODO func (me *Channel) TxCommit() error   { return nil }
//TODO func (me *Channel) TxRollback() error { return nil }
//TODO func (me *Channel) Flow(active bool) error { return nil }
