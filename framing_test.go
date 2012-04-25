package amqp

import (
	"amqp/wire"
	"bytes"
	"reflect"
	"testing"
)

func newFraming() (*Framing, chan wire.Frame, chan wire.Frame) {
	c2s := make(chan wire.Frame)
	s2c := make(chan wire.Frame)

	return NewFraming(1, 1024, s2c, c2s), s2c, c2s
}

func TestRecvMethod(t *testing.T) {
	f, s2c, _ := newFraming()
	frame := wire.MethodFrame{
		Channel: 1,
		Method:  wire.ConnectionOpenOk{},
	}

	go func() {
		s2c <- frame
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: wire.ConnectionOpenOk{}})
}

func TestRecvMethodMethod(t *testing.T) {
	f, s2c, _ := newFraming()

	go func() {
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.ConnectionOpenOk{}}
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.ConnectionTuneOk{}}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: wire.ConnectionOpenOk{}})
	assertEqualMessage(t, f.Recv(), Message{Method: wire.ConnectionTuneOk{}})
}

func TestRecvContentMethodContent(t *testing.T) {
	f, s2c, _ := newFraming()

	go func() {
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.ConnectionOpenOk{}}

		s2c <- wire.MethodFrame{Channel: 1, Method: wire.BasicDeliver{}}
		s2c <- wire.HeaderFrame{Channel: 1, Header: wire.ContentHeader{Size: 4, Class: 60}}
		s2c <- wire.BodyFrame{Channel: 1, Payload: []byte("oh")}
		s2c <- wire.BodyFrame{Channel: 1, Payload: []byte("ai")}

		s2c <- wire.MethodFrame{Channel: 1, Method: wire.ConnectionTuneOk{}}

		s2c <- wire.MethodFrame{Channel: 1, Method: wire.BasicDeliver{}}
		s2c <- wire.HeaderFrame{Channel: 1, Header: wire.ContentHeader{Size: 3, Class: 60}}
		s2c <- wire.BodyFrame{Channel: 1, Payload: []byte("lol")}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: wire.ConnectionOpenOk{}})
	assertEqualMessage(t, f.Recv(), Message{Method: wire.BasicDeliver{}, Body: []byte("ohai")})
	assertEqualMessage(t, f.Recv(), Message{Method: wire.ConnectionTuneOk{}})
	assertEqualMessage(t, f.Recv(), Message{Method: wire.BasicDeliver{}, Body: []byte("lol")})
}

func TestRecvContent(t *testing.T) {
	f, s2c, _ := newFraming()

	go func() {
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.BasicDeliver{}}
		s2c <- wire.HeaderFrame{Channel: 1, Header: wire.ContentHeader{Size: 4}}
		s2c <- wire.BodyFrame{Channel: 1, Payload: []byte("ohai")}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: wire.BasicDeliver{}, Body: []byte("ohai")})
}

func TestRecvInterruptedContent(t *testing.T) {
	f, s2c, _ := newFraming()

	go func() {
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.BasicDeliver{}}
		s2c <- wire.HeaderFrame{Channel: 1, Header: wire.ContentHeader{Size: 4}}
		s2c <- wire.BodyFrame{Channel: 1, Payload: []byte("oh")}
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.BasicCancelOk{}}
		s2c <- wire.BodyFrame{Channel: 1, Payload: []byte("ai")}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: wire.BasicCancelOk{}})
}

func TestRecvMethodContentMethod(t *testing.T) {
	f, s2c, _ := newFraming()

	go func() {
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.BasicDeliver{}}
		s2c <- wire.HeaderFrame{Channel: 1, Header: wire.ContentHeader{Size: 4}}
		s2c <- wire.BodyFrame{Channel: 1, Payload: []byte("ohai")}
		s2c <- wire.MethodFrame{Channel: 1, Method: wire.BasicCancelOk{}}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: wire.BasicDeliver{}, Body: []byte("ohai")})
	assertEqualMessage(t, f.Recv(), Message{Method: wire.BasicCancelOk{}})
}

func TestSendMethod(t *testing.T) {
	f, _, c2s := newFraming()

	go func() {
		f.Send(Message{Method: wire.BasicCancel{}})
		f.Send(Message{Method: wire.BasicConsume{}})
	}()

	assertEqualFrame(t, <-c2s, wire.MethodFrame{Channel: 1, Method: wire.BasicCancel{}})
	assertEqualFrame(t, <-c2s, wire.MethodFrame{Channel: 1, Method: wire.BasicConsume{}})
}

func TestSendMethodThenContent(t *testing.T) {
	f, _, c2s := newFraming()

	go func() {
		f.Send(Message{Method: wire.BasicConsumeOk{}})
		f.Send(Message{Method: wire.BasicPublish{}, Body: []byte("ohai")})
	}()

	assertEqualFrame(t, <-c2s, wire.MethodFrame{Channel: 1, Method: wire.BasicConsumeOk{}})
	assertEqualFrame(t, <-c2s, wire.MethodFrame{Channel: 1, Method: wire.BasicPublish{}})
	assertEqualFrame(t, <-c2s, wire.HeaderFrame{Channel: 1, Header: wire.ContentHeader{Size: 4, Class: 60}})
	assertEqualFrame(t, <-c2s, wire.BodyFrame{Channel: 1, Payload: []byte("ohai")})
}

func assertEqualMessage(t *testing.T, received Message, expected Message) {
	if received.Method != expected.Method {
		t.Errorf("different frame method expected: %v received: %v", received, expected)
	}

	if !reflect.DeepEqual(received.Properties, expected.Properties) {
		t.Errorf("different frame properties expected: %v received: %v", received, expected)
	}

	if bytes.Compare(received.Body, expected.Body) != 0 {
		t.Errorf("different frame body expected : %v received: %v", received, expected)
	}
}

func assertEqualFrame(t *testing.T, received wire.Frame, expected wire.Frame) {
	if !reflect.DeepEqual(received, expected) {
		t.Errorf("different frame method expected: %v received: %v", received, expected)
	}
}
