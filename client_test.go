// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"
)

type server struct {
	*testing.T
	r reader        // framer <- client
	w writer        // framer -> client
	S io.ReadWriter // Server IO
	C io.ReadWriter // Client IO
}

func defaultConfig() Config {
	return Config{SASL: []Authentication{&PlainAuth{"guest", "guest"}}, Vhost: "/"}
}

func newSession(t *testing.T) (io.ReadWriteCloser, *server) {
	rs, wc := io.Pipe()
	rc, ws := io.Pipe()

	rws := &logIO{t, "server", pipe{rs, ws}}
	rwc := &logIO{t, "client", pipe{rc, wc}}

	server := server{
		T: t,
		r: reader{rws},
		w: writer{rws},
		S: rws,
		C: rwc,
	}

	return rwc, &server
}

func (t *server) expectBytes(b []byte) {
	in := make([]byte, len(b))
	if _, err := io.ReadFull(t.S, in); err != nil {
		t.Fatalf("io error expecting bytes: %v", err)
	}

	if bytes.Compare(b, in) != 0 {
		t.Fatalf("failed bytes: expected: %s got: %s", string(b), string(in))
	}
}

func (t *server) send(channel int, m message) {
	defer time.AfterFunc(10*time.Millisecond, func() { panic("send deadlock") }).Stop()

	if err := t.w.WriteFrame(&methodFrame{
		ChannelId: uint16(channel),
		Method:    m,
	}); err != nil {
		t.Fatalf("frame err, write: %s", err)
	}
}

// Currently drops all but method frames expected on the given channel
func (t *server) recv(channel int, m message) message {
	defer time.AfterFunc(10*time.Millisecond, func() { panic("recv deadlock") }).Stop()

	var remaining int
	var header *headerFrame
	var body []byte

	for {
		frame, err := t.r.ReadFrame()
		if err != nil {
			t.Fatalf("frame err, read: %s", err)
		}

		if frame.channel() != uint16(channel) {
			t.Fatalf("expected frame on channel %d, got channel %d", channel, frame.channel())
		}

		switch f := frame.(type) {
		case *heartbeatFrame:
			// drop

		case *headerFrame:
			// start content state
			header = f
			remaining = int(header.Size)

		case *bodyFrame:
			// continue until terminated
			body = append(body, f.Body...)
			remaining -= len(f.Body)
			if remaining <= 0 {
				m.(messageWithContent).setContent(header.Properties, body)
				return m
			}

		case *methodFrame:
			if reflect.TypeOf(m) == reflect.TypeOf(f.Method) {
				wantv := reflect.ValueOf(m).Elem()
				havev := reflect.ValueOf(f.Method).Elem()
				wantv.Set(havev)
				if _, ok := m.(messageWithContent); !ok {
					return m
				}
			} else {
				t.Fatalf("expected method type: %T, got: %T", m, f.Method)
			}
		}
	}

	panic("unreachable")
}

func (t *server) expectAMQP() {
	t.expectBytes([]byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1})
}

func (t *server) connectionStart() {
	t.send(0, &connectionStart{
		VersionMajor: 0,
		VersionMinor: 9,
		Mechanisms:   "PLAIN",
		Locales:      "en-us",
	})

	t.recv(0, &connectionStartOk{})
}

func (t *server) connectionTune() {
	t.send(0, &connectionTune{
		ChannelMax: 11,
		FrameMax:   20000,
		Heartbeat:  10,
	})

	t.recv(0, &connectionTuneOk{})
}

func (t *server) connectionOpen() {
	t.expectAMQP()
	t.connectionStart()
	t.connectionTune()

	t.recv(0, &connectionOpen{})
	t.send(0, &connectionOpenOk{})
}

func (t *server) connectionClose() {
	t.recv(0, &connectionClose{})
	t.send(0, &connectionCloseOk{})
}

func (t *server) channelOpen(id int) {
	t.recv(id, &channelOpen{})
	t.send(id, &channelOpenOk{})
}

func TestOpen(t *testing.T) {
	rwc, srv := newSession(t)
	go func() {
		srv.connectionOpen()
		rwc.Close()
	}()

	if c, err := Open(rwc, defaultConfig()); err != nil {
		t.Fatalf("could not create connection: %s (%s)", c, err)
	}
}

func TestChannelOpen(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		rwc.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %s (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %s (%s)", ch, err)
	}
}

func TestOpenFailedSASLUnsupportedMechanisms(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.expectAMQP()
		srv.send(0, &connectionStart{
			VersionMajor: 0,
			VersionMinor: 9,
			Mechanisms:   "KERBEROS NTLM",
			Locales:      "en-us",
		})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != ErrSASL {
		t.Fatalf("expected ErrSASL got: %+v on %+v", err, c)
	}
}

func TestOpenFailedCredentials(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.expectAMQP()
		srv.connectionStart()
		// Now kill/timeout the connection indicating bad auth
		rwc.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != ErrCredentials {
		t.Fatalf("expected ErrCredentials got: %+v on %+v", err, c)
	}
}

func TestOpenFailedVhost(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.expectAMQP()
		srv.connectionStart()
		srv.connectionTune()
		srv.recv(0, &connectionOpen{})

		// Now kill/timeout the connection on bad Vhost
		rwc.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != ErrVhost {
		t.Fatalf("expected ErrVhost got: %+v on %+v", err, c)
	}
}

func TestConfirmMultiple(t *testing.T) {
	rwc, srv := newSession(t)
	defer rwc.Close()

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		srv.recv(1, &confirmSelect{})
		srv.send(1, &confirmSelectOk{})

		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})

		// Single tag, plus multiple, should produce
		// 2, 1, 3, 4
		srv.send(1, &basicAck{DeliveryTag: 2})
		srv.send(1, &basicAck{DeliveryTag: 4, Multiple: true})

		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})

		// And some more, but in reverse order, multiple then one
		// 5, 6, 7, 8
		srv.send(1, &basicAck{DeliveryTag: 6, Multiple: true})
		srv.send(1, &basicAck{DeliveryTag: 8})
		srv.send(1, &basicAck{DeliveryTag: 7})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %s (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %s (%s)", ch, err)
	}

	acks, _ := ch.NotifyConfirm(make(chan uint64), make(chan uint64))

	ch.Confirm(false)

	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 1")})
	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 2")})
	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 3")})
	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 4")})

	for i, tag := range []uint64{2, 1, 3, 4} {
		if ack := <-acks; tag != ack {
			t.Fatalf("failed ack, expected ack#%d to be %d, got %d", i, tag, ack)
		}
	}

	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 5")})
	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 6")})
	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 7")})
	ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 8")})

	for i, tag := range []uint64{5, 6, 8, 7} {
		if ack := <-acks; tag != ack {
			t.Fatalf("failed ack, expected ack#%d to be %d, got %d", i, tag, ack)
		}
	}

}

func TestNotifyClosesReusedPublisherConfirmChan(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		srv.recv(1, &confirmSelect{})
		srv.send(1, &confirmSelectOk{})

		srv.recv(0, &connectionClose{})
		srv.send(0, &connectionCloseOk{})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %s (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %s (%s)", ch, err)
	}

	ackAndNack := make(chan uint64)
	ch.NotifyConfirm(ackAndNack, ackAndNack)

	if err := ch.Confirm(false); err != nil {
		t.Fatalf("expected to enter confirm mode: %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("could not close connection: %s (%s)", c, err)
	}
}

func TestNotifyClosesAllChansAfterConnectionClose(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		srv.recv(0, &connectionClose{})
		srv.send(0, &connectionCloseOk{})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %s (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %s (%s)", ch, err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("could not close connection: %s (%s)", c, err)
	}

	select {
	case <-c.NotifyClose(make(chan *Error)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close NotifyClose chan after Connection.Close")
	}

	select {
	case <-ch.NotifyClose(make(chan *Error)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close Connection.NotifyClose chan after Connection.Close")
	}

	select {
	case <-ch.NotifyFlow(make(chan bool)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close Channel.NotifyFlow chan after Connection.Close")
	}

	select {
	case <-ch.NotifyReturn(make(chan Return)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close Channel.NotifyReturn chan after Connection.Close")
	}

	ack, nack := ch.NotifyConfirm(make(chan uint64), make(chan uint64))

	select {
	case <-ack:
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close acks on Channel.NotifyConfirm chan after Connection.Close")
	}

	select {
	case <-nack:
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close nacks Channel.NotifyConfirm chan after Connection.Close")
	}
}
