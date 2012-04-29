package amqp_test

import (
	"amqp"
	"amqp/wire"
	"bytes"
	"testing"
)

func TestNewClientHandshake(t *testing.T) {
	server, client := interPipes(t)

	go func() {
		var f wire.Frame
		var err error
		var ok bool

		handshake := make([]byte, 8)
		server.Read(handshake)
		if bytes.Compare(handshake, []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}) != 0 {
			t.Fatal("bad protocol handshake", handshake)
		}

		r := wire.NewFrameReader(server)

		wire.MethodFrame{
			Channel: 0,
			Method: wire.ConnectionStart{
				VersionMajor: 0,
				VersionMinor: 9,
				Mechanisms:   "PLAIN",
				Locales:      "en-us",
			},
		}.WriteTo(server)

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionStartOk); !ok {
			t.Fatal("expected ConnectionStartOk")
		}

		wire.MethodFrame{
			Channel: 0,
			Method: wire.ConnectionTune{
				ChannelMax: 11,
				FrameMax:   20000,
				Heartbeat:  10,
			},
		}.WriteTo(server)

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionTuneOk); !ok {
			t.Fatal("expected ConnectionTuneOk")
		}

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionOpen); !ok {
			t.Fatal("expected ConnectionOpen")
		}

		wire.MethodFrame{
			Channel: 0,
			Method:  wire.ConnectionOpenOk{},
		}.WriteTo(server)

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ChannelOpen); !ok {
			t.Fatal("expected ChannelOpen")
		}

		if f.ChannelID() != 1 {
			t.Fatal("expected ChannelOpen on channel 1")
		}

		wire.MethodFrame{
			Channel: 1,
			Method:  wire.ChannelOpenOk{},
		}.WriteTo(server)

		server.Close()
	}()

	c, err := amqp.NewConnection(client, &amqp.PlainAuth{"guest", "guest"}, "/")
	if err != nil {
		t.Fatal("could not create connection:", c, err)
	}

	ch, err := c.OpenChannel()
	if err != nil {
		t.Fatal("could not open channel:", ch, err)
	}
}
