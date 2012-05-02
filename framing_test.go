package amqp

import (
	"bytes"
	"reflect"
	"testing"
)

func newTestFraming() (*Framing, chan frame, chan frame) {
	c2s := make(chan frame)
	s2c := make(chan frame)

	return newFraming(1, 1024, s2c, c2s), s2c, c2s
}

func TestRecvMethod(t *testing.T) {
	f, s2c, _ := newTestFraming()
	frame := methodFrame{
		Channel: 1,
		Method:  ConnectionOpenOk{},
	}

	go func() {
		s2c <- frame
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: ConnectionOpenOk{}})
}

func TestRecvMethodMethod(t *testing.T) {
	f, s2c, _ := newTestFraming()

	go func() {
		s2c <- methodFrame{Channel: 1, Method: ConnectionOpenOk{}}
		s2c <- methodFrame{Channel: 1, Method: ConnectionTuneOk{}}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: ConnectionOpenOk{}})
	assertEqualMessage(t, f.Recv(), Message{Method: ConnectionTuneOk{}})
}

func TestRecvContentMethodContent(t *testing.T) {
	f, s2c, _ := newTestFraming()

	go func() {
		s2c <- methodFrame{Channel: 1, Method: ConnectionOpenOk{}}

		s2c <- methodFrame{Channel: 1, Method: BasicDeliver{}}
		s2c <- headerFrame{Channel: 1, Header: ContentHeader{Size: 4, Class: 60}}
		s2c <- bodyFrame{Channel: 1, Payload: []byte("oh")}
		s2c <- bodyFrame{Channel: 1, Payload: []byte("ai")}

		s2c <- methodFrame{Channel: 1, Method: ConnectionTuneOk{}}

		s2c <- methodFrame{Channel: 1, Method: BasicDeliver{}}
		s2c <- headerFrame{Channel: 1, Header: ContentHeader{Size: 3, Class: 60}}
		s2c <- bodyFrame{Channel: 1, Payload: []byte("lol")}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: ConnectionOpenOk{}})
	assertEqualMessage(t, <-f.async, Message{Method: BasicDeliver{}, Body: []byte("ohai")})
	assertEqualMessage(t, f.Recv(), Message{Method: ConnectionTuneOk{}})
	assertEqualMessage(t, <-f.async, Message{Method: BasicDeliver{}, Body: []byte("lol")})
}

func TestRecvContent(t *testing.T) {
	f, s2c, _ := newTestFraming()

	go func() {
		s2c <- methodFrame{Channel: 1, Method: BasicDeliver{}}
		s2c <- headerFrame{Channel: 1, Header: ContentHeader{Size: 4}}
		s2c <- bodyFrame{Channel: 1, Payload: []byte("ohai")}
	}()

	assertEqualMessage(t, <-f.async, Message{Method: BasicDeliver{}, Body: []byte("ohai")})
}

func TestRecvInterruptedContent(t *testing.T) {
	f, s2c, _ := newTestFraming()

	go func() {
		s2c <- methodFrame{Channel: 1, Method: BasicDeliver{}}
		s2c <- headerFrame{Channel: 1, Header: ContentHeader{Size: 4}}
		s2c <- bodyFrame{Channel: 1, Payload: []byte("oh")}
		s2c <- methodFrame{Channel: 1, Method: BasicCancelOk{}}
		s2c <- bodyFrame{Channel: 1, Payload: []byte("ai")}
	}()

	assertEqualMessage(t, f.Recv(), Message{Method: BasicCancelOk{}})
}

func TestRecvMethodContentMethod(t *testing.T) {
	f, s2c, _ := newTestFraming()

	go func() {
		s2c <- methodFrame{Channel: 1, Method: BasicDeliver{}}
		s2c <- headerFrame{Channel: 1, Header: ContentHeader{Size: 4}}
		s2c <- bodyFrame{Channel: 1, Payload: []byte("ohai")}
		s2c <- methodFrame{Channel: 1, Method: BasicCancelOk{}}
	}()

	assertEqualMessage(t, <-f.async, Message{Method: BasicDeliver{}, Body: []byte("ohai")})
	assertEqualMessage(t, f.Recv(), Message{Method: BasicCancelOk{}})
}

func TestSendMethod(t *testing.T) {
	f, _, c2s := newTestFraming()

	go func() {
		f.Send(Message{Method: BasicCancel{}})
		f.Send(Message{Method: BasicConsume{}})
	}()

	assertEqualFrame(t, <-c2s, methodFrame{Channel: 1, Method: BasicCancel{}})
	assertEqualFrame(t, <-c2s, methodFrame{Channel: 1, Method: BasicConsume{}})
}

func TestSendMethodThenContent(t *testing.T) {
	f, _, c2s := newTestFraming()

	go func() {
		f.Send(Message{Method: BasicConsumeOk{}})
		f.Send(Message{Method: BasicPublish{}, Body: []byte("ohai")})
	}()

	assertEqualFrame(t, <-c2s, methodFrame{Channel: 1, Method: BasicConsumeOk{}})
	assertEqualFrame(t, <-c2s, methodFrame{Channel: 1, Method: BasicPublish{}})
	assertEqualFrame(t, <-c2s, headerFrame{Channel: 1, Header: ContentHeader{Size: 4, Class: 60}})
	assertEqualFrame(t, <-c2s, bodyFrame{Channel: 1, Payload: []byte("ohai")})
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

func assertEqualFrame(t *testing.T, received frame, expected frame) {
	if !reflect.DeepEqual(received, expected) {
		t.Errorf("different frame method expected: %v received: %v", received, expected)
	}
}
