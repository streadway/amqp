package amqp_test

import (
	"amqp"
	"bytes"
	"testing"
)

func TestNewClientHandshake(t *testing.T) {
	server, client := interPipes(t)

	go func() {
		var f amqp.Frame
		var err error
		var ok bool

		handshake := make([]byte, 8)
		server.Read(handshake)
		if bytes.Compare(handshake, []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}) != 0 {
			t.Fatal("bad protocol handshake", handshake)
		}

		r := amqp.NewFrameReader(server)

		amqp.MethodFrame{
			Channel: 0,
			Method: amqp.ConnectionStart{
				VersionMajor: 0,
				VersionMinor: 9,
				Mechanisms:   "PLAIN",
				Locales:      "en-us",
			},
		}.WriteTo(server)

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(amqp.MethodFrame).Method.(amqp.ConnectionStartOk); !ok {
			t.Fatal("expected ConnectionStartOk")
		}

		amqp.MethodFrame{
			Channel: 0,
			Method: amqp.ConnectionTune{
				ChannelMax: 11,
				FrameMax:   20000,
				Heartbeat:  10,
			},
		}.WriteTo(server)

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(amqp.MethodFrame).Method.(amqp.ConnectionTuneOk); !ok {
			t.Fatal("expected ConnectionTuneOk")
		}

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(amqp.MethodFrame).Method.(amqp.ConnectionOpen); !ok {
			t.Fatal("expected ConnectionOpen")
		}

		amqp.MethodFrame{
			Channel: 0,
			Method:  amqp.ConnectionOpenOk{},
		}.WriteTo(server)

		if f, err = r.NextFrame(); err != nil {
			t.Fatal("bad read", err)
		}

		if _, ok = f.(amqp.MethodFrame).Method.(amqp.ChannelOpen); !ok {
			t.Fatal("expected ChannelOpen")
		}

		if f.ChannelID() != 1 {
			t.Fatal("expected ChannelOpen on channel 1")
		}

		amqp.MethodFrame{
			Channel: 1,
			Method:  amqp.ChannelOpenOk{},
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
