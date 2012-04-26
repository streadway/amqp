package amqp

import (
	"amqp/wire"
	"fmt"
)

// Struct that contains the method and any content if the method HasContent
// intended to be used by higher level APIs like Client/Channel/Connection
type Message struct {
	Method     wire.Method
	Properties wire.ContentProperties
	Body       []byte
}

// Manages the multiplexing and demultiplexing of frames and follows the rules
// about interleaving method/header/content frames on the same channel.
//
// XXX(ST) Framing needs a better name, this handles the framing rules but
// also handles routing to a shared channel.
type Framing struct {
	sync  chan Message
	async chan Message

	out chan chan wire.Frame

	c2s chan wire.Frame
	s2c chan wire.Frame

	channel uint16
	maxSize int

	// State machine that manages frame order
	recv func(*Framing, wire.Frame) error

	method *wire.MethodFrame
	header *wire.HeaderFrame
	body   []byte
}

func newFraming(channel uint16, maxSize int, s2c, c2s chan wire.Frame) *Framing {
	me := &Framing{
		sync:    make(chan Message),
		async:   make(chan Message),
		out:     make(chan chan wire.Frame),
		c2s:     c2s,
		s2c:     s2c,
		recv:    (*Framing).recvMethod,
		channel: channel,
		maxSize: maxSize,
	}
	go me.loop()
	return me
}

func (me *Framing) loop() {
	for {
		select {
		case set := <-me.out:
			for f := range set {
				me.c2s <- f
			}
		case f := <-me.s2c:
			switch frame := f.(type) {
			case wire.MethodFrame, wire.HeaderFrame, wire.BodyFrame:
				// run the state machine
				me.recv(me, frame)
			case wire.HeartbeatFrame:
				// meh drop for now
			default:
				// protocol error
				panic("TODO protocol error")
			}
		}
	}
}

func (me *Framing) Send(msg Message) {
	fmt.Println("Send: msg", msg)

	set := make(chan wire.Frame)
	me.out <- set
	set <- wire.MethodFrame{
		Channel: me.channel,
		Method:  msg.Method,
	}

	if msg.Method.HasContent() {
		set <- wire.HeaderFrame{
			Channel: me.channel,
			Header: wire.ContentHeader{
				Class:      msg.Method.Class(),
				Size:       uint64(len(msg.Body)),
				Properties: msg.Properties,
			},
		}

		for i := 0; i < len(msg.Body); i += me.maxSize {
			j := i + me.maxSize
			if j > len(msg.Body) {
				j = len(msg.Body)
			}

			set <- wire.BodyFrame{
				Channel: me.channel,
				Payload: msg.Body[i:j],
			}
		}
	}

	close(set)
}

func (me *Framing) Recv() Message {
	return <-me.sync
}

func (me *Framing) RecvAsynchronous() Message {
	return <-me.async
}

func (me *Framing) transition(f func(*Framing, wire.Frame) error) error {
	me.recv = f
	return nil
}

// readMethod
// hasContent

func (me *Framing) recvMethod(f wire.Frame) error {
	fmt.Println("recvMethod: f,c: ", f, me.method, me.header)

	switch frame := f.(type) {
	case wire.MethodFrame:
		me.method = &frame

		if frame.Method.HasContent() {
			// XXX(ST) body grows until the largest body size
			me.body = me.body[0:0]
			return me.transition((*Framing).recvHeader)
		} else {
			msg := Message{
				Method: me.method.Method,
			}

			if frame.Method.IsSynchronous() {
				me.sync <- msg
			} else {
				me.async <- msg
			}
			return me.transition((*Framing).recvMethod)
		}

	case wire.HeaderFrame:
		// drop
		return me.transition((*Framing).recvMethod)

	case wire.BodyFrame:
		// drop
		return me.transition((*Framing).recvMethod)
	}

	panic("unreachable")
}

func (me *Framing) recvHeader(f wire.Frame) error {
	fmt.Println("recvHeader: f,c: ", f, me.method, me.header)

	switch frame := f.(type) {
	case wire.MethodFrame:
		// interrupt content and handle method
		return me.recvMethod(f)

	case wire.HeaderFrame:
		// start collecting
		me.header = &frame
		return me.transition((*Framing).recvContent)

	case wire.BodyFrame:
		// drop and reset
		return me.transition((*Framing).recvMethod)
	}

	panic("unreachable")
}

// state after method + header and before the length
// defined by the header has been reached
func (me *Framing) recvContent(f wire.Frame) error {
	fmt.Println("recvContent: f: ", f)

	switch frame := f.(type) {
	case wire.MethodFrame:
		// interrupt content and handle method
		return me.recvMethod(f)

	case wire.HeaderFrame:
		// drop and reset
		return me.transition((*Framing).recvMethod)

	case wire.BodyFrame:
		me.body = append(me.body, frame.Payload...)

		if uint64(len(me.body)) >= me.header.Header.Size {
			msg := Message{
				Method:     me.method.Method,
				Properties: me.header.Header.Properties,
				Body:       me.body,
			}

			if me.method.Method.IsSynchronous() {
				me.sync <- msg
			} else {
				me.async <- msg
			}

			return me.transition((*Framing).recvMethod)
		}

		return me.transition((*Framing).recvContent)
	}

	panic("unreachable")
}
