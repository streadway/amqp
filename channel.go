package amqp

import (
//	"fmt"
)

const (
	Direct  = "direct"
	Topic   = "topic"
	Fanout  = "fanout"
	Headers = "headers"
)

// Represents an AMQP channel, used for concurrent, interleaved publishers and
// consumers on the same connection.
type Channel struct {
	sync  chan Message
	async chan Message

	out chan chan Frame

	c2s chan Frame
	s2c chan Frame

	id      uint16
	maxSize int

	// State machine that manages frame order
	recv func(*Channel, Frame) error

	// Current state for frame re-assembly
	message MessageWithContent
	header  *HeaderFrame
	body    []byte
}

// Constructs and opens a new channel with the given framing rules
func newChannel(id uint16, maxFrameSize int, in chan Frame, out chan Frame) (me *Channel, err error) {
	me = &Channel{
		id:      id,
		maxSize: maxFrameSize,
		c2s:     out,
		s2c:     in,
		recv:    (*Channel).recvMethod,
	}

	return me, nil
}

//    channel             = open-channel *use-channel close-channel
//    open-channel        = C:OPEN S:OPEN-OK
//    use-channel         = C:FLOW S:FLOW-OK
//                        / S:FLOW C:FLOW-OK
//                        / functional-class
//    close-channel       = C:CLOSE S:CLOSE-OK
//                        / S:CLOSE C:CLOSE-OK
//func (me *Channel) open() error {
//	me.framing.SendMethod(ChannelOpen{})
//
//	switch me.framing.Recv().Method.(type) {
//	case ChannelOpenOk:
//		return nil
//	}
//
//	// TODO handle channel open errors (like already opened on this ID)
//	return ErrBadProtocol
//}
//
//// Manages the multiplexing and demultiplexing of frames and follows the rules
//// about interleaving method/header/content frames on the same channel.
//type Channel struct {
//}
//
//func newChannel(channel uint16, maxSize int, s2c, c2s chan Frame) *Channel {
//	me := &Channel{
//		sync:    make(chan Message),
//		async:   make(chan Message),
//		out:     make(chan chan Frame),
//		c2s:     c2s,
//		s2c:     s2c,
//		recv:    (*Channel).recvMethod,
//		channel: channel,
//		maxSize: maxSize,
//	}
//	go me.loop()
//	return me
//}
//
//func (me *Channel) loop() {
//	for {
//		select {
//		case set := <-me.out:
//			for f := range set {
//				me.c2s <- f
//			}
//		case f := <-me.s2c:
//			switch frame := f.(type) {
//			case MethodFrame, HeaderFrame, BodyFrame:
//				// run the state machine
//				me.recv(me, frame)
//			case HeartbeatFrame:
//				// meh drop for now
//			default:
//				// protocol error
//				panic("TODO protocol error")
//			}
//		}
//	}
//}
//
//// Frames and sends a method that should not have a payload
//func (me *Channel) SendMethod(method interface{}) {
//	me.Send(Message{Method: method})
//}
//
//// Frames and sends a method that may or may not have payload
//func (me *Channel) Send(msg Message) {
//	set := make(chan Frame)
//	me.out <- set
//
//	set <- MethodFrame{
//		Channel: me.channel,
//		Method:  msg.Method,
//	}
//
//	if msg.Method.HasContent() {
//		set <- HeaderFrame{
//			Channel: me.channel,
//			Header: ContentHeader{
//				Class:      msg.Method.Class(),
//				Size:       uint64(len(msg.Body)),
//				Properties: msg.Properties,
//			},
//		}
//
//		for i := 0; i < len(msg.Body); i += me.maxSize {
//			j := i + me.maxSize
//			if j > len(msg.Body) {
//				j = len(msg.Body)
//			}
//
//			set <- BodyFrame{
//				Channel: me.channel,
//				Payload: msg.Body[i:j],
//			}
//		}
//	}
//
//	close(set)
//}
//
//func (me *Channel) Recv() Message {
//	return <-me.sync
//}
//
func (me *Channel) transition(f func(*Channel, Frame) error) error {
	me.recv = f
	return nil
}

// readMethod
// hasContent

func (me *Channel) recvMethod(f Frame) error {
	switch frame := f.(type) {
	case *MethodFrame:
		if msg, ok := frame.Method.(MessageWithContent); ok {
			// XXX(ST) body grows until the largest body size
			me.body = me.body[0:0]
			me.message = msg
			return me.transition((*Channel).recvHeader)
		}

		if frame.Method.wait() {
			me.sync <- frame.Method
		} else {
			me.async <- frame.Method
		}
		return me.transition((*Channel).recvMethod)

	case *HeaderFrame:
		// drop
		return me.transition((*Channel).recvMethod)

	case *BodyFrame:
		// drop
		return me.transition((*Channel).recvMethod)
	}

	panic("unreachable")
}

func (me *Channel) recvHeader(f Frame) error {
	switch frame := f.(type) {
	case *MethodFrame:
		// interrupt content and handle method
		return me.recvMethod(f)

	case *HeaderFrame:
		// start collecting
		me.header = frame
		return me.transition((*Channel).recvContent)

	case *BodyFrame:
		// drop and reset
		return me.transition((*Channel).recvMethod)
	}

	panic("unreachable")
}

// state after method + header and before the length
// defined by the header has been reached
func (me *Channel) recvContent(f Frame) error {
	switch frame := f.(type) {
	case *MethodFrame:
		// interrupt content and handle method
		return me.recvMethod(f)

	case *HeaderFrame:
		// drop and reset
		return me.transition((*Channel).recvMethod)

	case *BodyFrame:
		me.body = append(me.body, frame.Body...)

		if uint64(len(me.body)) >= me.header.Size {
			me.message.SetContent(me.header.Properties, me.body)

			if me.message.wait() {
				me.sync <- me.message
			} else {
				me.async <- me.message
			}

			return me.transition((*Channel).recvMethod)
		}

		return me.transition((*Channel).recvContent)
	}

	panic("unreachable")
}
