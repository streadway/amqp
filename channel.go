package amqp

import (
	"fmt"
)

// Represents an AMQP channel, used for concurrent, interleaved publishers and
// consumers on the same connection.
type Channel struct {
	Closed *Closed

	connection *Connection

	// Either asynchronous 
	rpc chan message

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
	// TODO Mutex
	if tag == "" {
		tag = randomTag()
	}

	me.consumers[tag] = ch

	return tag
}

func (me *Channel) shutdown() {
	fmt.Println("channel shutdown", me.state)
	me.state = closing

	delete(me.connection.channels, me.id)

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
		size := me.connection.MaxFrameSize

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

	me.state = closing

	me.send(&channelClose{ReplyCode: ReplySuccess})

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

func (me *Channel) deliver(msg messageWithContent, c chan Delivery) {
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

	c <- delivery
}

// Eventually called via the state machine from the connection's reader goroutine so
// assumes serialized access
func (me *Channel) dispatch(msg message) {
	switch m := msg.(type) {
	case *channelClose:
		me.send(&channelCloseOk{})
		if me.state == open {
			me.Close()
		}

	case *channelFlow:
		// unhandled

	case *basicDeliver:
		if me.state == open {
			if c, ok := me.consumers[m.ConsumerTag]; ok {
				me.deliver(m, c)
			}
			// TODO log failed consumer
		}

	default:
		if me.state != closed {
			me.rpc <- m
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
			me.body = me.body[0:0]
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

// RPC Implementation

func (me *Channel) E(name string) *Exchange {
	return &Exchange{channel: me, name: name}
}

func (me *Channel) Q(name string) *Queue {
	return &Queue{channel: me, name: name}
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

//TODO func (me *Channel) Get(queueName string, noAck bool) error                     { return nil }
//TODO func (me *Channel) Recover(requeue bool) error                                 { return nil }
//TODO func (me *Channel) Nack(deliveryTag uint64, requeue bool, multiple bool) error { return nil }

//TODO func (me *Channel) Confirm(noWait bool) error                                  { return nil }

//TODO func (me *Channel) TxSelect() error   { return nil }
//TODO func (me *Channel) TxCommit() error   { return nil }
//TODO func (me *Channel) TxRollback() error { return nil }
//TODO func (me *Channel) Flow(active bool) error { return nil }
