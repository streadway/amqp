package amqp

import (
	//	"fmt"
	"sync"
)

// Represents an AMQP channel, used for concurrent, interleaved publishers and
// consumers on the same connection.
type Channel struct {
	Closed *Closed

	connection *Connection

	// Either asynchronous 
	rpc chan message

	consumers map[string]chan Delivery

	id     uint16
	closed bool

	// Sequences sending content frames against to the connection,
	// can be controlled by flow messages
	flow sync.Mutex

	// State machine that manages frame order
	recv func(*Channel, frame) error

	// Current state for frame re-assembly
	message messageWithContent
	header  *headerFrame
	body    []byte
}

// Constructs and opens a new channel with the given framing rules
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

func (me *Channel) send(msg message) {
	if content, ok := msg.(messageWithContent); ok {
		me.flow.Lock()
		defer me.flow.Unlock()

		props, body := content.getContent()
		class, _ := content.id()
		size := me.connection.MaxFrameSize

		me.connection.send(&methodFrame{
			ChannelId: me.id,
			Method:    content,
		})

		me.connection.send(&headerFrame{
			ChannelId:  me.id,
			ClassId:    class,
			Size:       uint64(len(body)),
			Properties: props,
		})

		for i, j := 0, size; i < len(body); i, j = j, j+size {
			if j > len(body) {
				j = len(body)
			}

			me.connection.send(&bodyFrame{
				ChannelId: me.id,
				Body:      body[i:j],
			})
		}
	} else {
		me.connection.send(&methodFrame{
			ChannelId: me.id,
			Method:    msg,
		})
	}
}

// Initiate a clean channel closure by sending a close message with the error code set to '200'
func (me *Channel) Close() (err error) {
	if me.closed {
		return ErrAlreadyClosed
	}
	me.send(&channelClose{ReplyCode: ReplySuccess})
	return nil
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

func (me *Channel) terminate() {
	close(me.rpc)

	for k, c := range me.consumers {
		delete(me.consumers, k)
		close(c)
	}
}

// Eventually called via the state machine from the connection's reader goroutine so
// assumes serialized access
func (me *Channel) dispatch(msg message) {
	switch m := msg.(type) {
	case *channelClose:
		me.send(&channelCloseOk{})
		if !me.closed {
			// when the connection is closed, we'll get an error here
			me.send(&channelClose{ReplyCode: ReplySuccess})
		}

	case *channelCloseOk:
		if !me.closed {
			me.closed = true
			me.terminate()
		}

	case *channelFlow:
		// unhandled

	case *basicDeliver:
		if !me.closed {
			if c, ok := me.consumers[m.ConsumerTag]; ok {
				me.deliver(m, c)
			}
			// TODO log failed consumer
		}

	default:
		if !me.closed {
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

		me.dispatch(me.message) // termination state
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
		// start collecting
		me.header = frame
		return me.transition((*Channel).recvContent)

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

func (me *Channel) E(name string) Exchange { return Exchange{channel: me, name: name} }
func (me *Channel) Q(name string) Queue    { return Queue{channel: me, name: name} }

func (me *Channel) Flow(active bool) error { return nil }

func (me *Channel) Qos(prefetchCount uint16, prefetchSize uint32, global bool) error { return nil }

func (me *Channel) Cancel(consumerTag string, noWait bool) error { return nil }

func (me *Channel) Get(queueName string, noAck bool) error                     { return nil }
func (me *Channel) Recover(requeue bool) error                                 { return nil }
func (me *Channel) Nack(deliveryTag uint64, requeue bool, multiple bool) error { return nil }
func (me *Channel) Confirm(noWait bool) error                                  { return nil }

func (me *Channel) TxSelect() error   { return nil }
func (me *Channel) TxCommit() error   { return nil }
func (me *Channel) TxRollback() error { return nil }
