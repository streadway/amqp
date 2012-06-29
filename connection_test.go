package amqp

import (
	"bytes"
	"io"
	"testing"
)

func driveConnectionOpen(t *testing.T, server io.ReadWriter) {
	var f frame
	var err error
	var ok bool

	handshake := make([]byte, 8)
	server.Read(handshake)
	if bytes.Compare(handshake, []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}) != 0 {
		t.Fatalf("bad protocol handshake: %s", handshake)
	}

	r := reader{server}
	w := writer{server}

	if err = w.WriteFrame(&methodFrame{
		ChannelId: 0,
		Method: &connectionStart{
			VersionMajor: 0,
			VersionMinor: 9,
			Mechanisms:   "PLAIN",
			Locales:      "en-us",
		},
	}); err != nil {
		t.Fatalf("bad write")
	}

	if f, err = r.ReadFrame(); err != nil {
		t.Fatalf("bad read: %s", err)
	}

	if _, ok = f.(*methodFrame).Method.(*connectionStartOk); !ok {
		t.Fatalf("expected ConnectionStartOk")
	}

	if err = w.WriteFrame(&methodFrame{
		ChannelId: 0,
		Method: &connectionTune{
			ChannelMax: 11,
			FrameMax:   20000,
			Heartbeat:  10,
		},
	}); err != nil {
		t.Fatalf("bad write: %s", err)
	}

	if f, err = r.ReadFrame(); err != nil {
		t.Fatalf("bad read: %s", err)
	}

	if _, ok = f.(*methodFrame).Method.(*connectionTuneOk); !ok {
		t.Fatalf("expected ConnectionTuneOk")
	}

	if f, err = r.ReadFrame(); err != nil {
		t.Fatalf("bad read: %s", err)
	}

	if _, ok = f.(*methodFrame).Method.(*connectionOpen); !ok {
		t.Fatalf("expected ConnectionOpen")
	}

	if err = w.WriteFrame(&methodFrame{
		ChannelId: 0,
		Method:    &connectionOpenOk{},
	}); err != nil {
		t.Fatalf("bad write: %s", err)
	}
}

func driveChannelOpen(t *testing.T, server io.ReadWriteCloser) {
	var f frame
	var err error
	var ok bool

	r := reader{server}
	w := writer{server}

	if f, err = r.ReadFrame(); err != nil {
		t.Fatalf("bad read: %s", err)
	}

	if _, ok = f.(*methodFrame).Method.(*channelOpen); !ok {
		t.Fatalf("expected channelOpen")
	}

	if err = w.WriteFrame(&methodFrame{
		ChannelId: f.channel(),
		Method:    &channelOpenOk{},
	}); err != nil {
		t.Fatalf("bad write")
	}
}

func TestNewConnectionOpen(t *testing.T) {
	server, client := interPipes(t)

	go driveConnectionOpen(t, server)

	c, err := NewConnection(client, &PlainAuth{"guest", "guest"}, "/")
	if err != nil {
		t.Fatalf("could not create connection: %s (%s)", c, err)
	}
}

func TestNewConnectionChannelOpen(t *testing.T) {
	server, client := interPipes(t)

	go driveConnectionOpen(t, server)

	c, err := NewConnection(client, &PlainAuth{"guest", "guest"}, "/")
	if err != nil {
		t.Fatalf("could not create connection: %s (%s)", c, err)
	}

	go driveChannelOpen(t, server)

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %s (%s) %s %s", ch, err, c.state, ch.state)
	}
}
