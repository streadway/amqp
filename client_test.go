package amqp_test

import (
	"amqp"
	"amqp/wire"
	"bytes"
	"testing"
)

func TestNewClientHandshake(t *testing.T) {
	done := make(chan bool)
	server, client := interPipes()

	go func() {
		wire.MethodFrame{
			Channel: 0,
			Method: wire.ConnectionStart{
				VersionMajor: 0,
				VersionMinor: 9,
				Mechanisms:   "PLAIN",
				Locales:      "en-us",
			},
		}.WriteTo(server)

		wire.MethodFrame{
			Channel: 0,
			Method: wire.ConnectionTune{
				ChannelMax: 11,
				FrameMax:   20000,
				Heartbeat:  10,
			},
		}.WriteTo(server)

		wire.MethodFrame{
			Channel: 0,
			Method:  wire.ConnectionOpenOk{},
		}.WriteTo(server)

		wire.MethodFrame{
			Channel: 1,
			Method:  wire.ChannelOpenOk{},
		}.WriteTo(server)
	}()

	go func() {
		handshake := make([]byte, 8)
		server.Read(handshake)
		if bytes.Compare(handshake, []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}) != 0 {
			t.Error("bad protocol handshake", handshake)
		}

		r := wire.NewFrameReader(server)

		var f wire.Frame
		var err error
		var ok bool

		if f, err = r.NextFrame(); err != nil {
			t.Error("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionStartOk); !ok {
			t.Error("expected ConnectionStartOk")
		}

		if f, err = r.NextFrame(); err != nil {
			t.Error("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionTuneOk); !ok {
			t.Error("expected ConnectionTuneOk")
		}

		if f, err = r.NextFrame(); err != nil {
			t.Error("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionOpen); !ok {
			t.Error("expected ConnectionOpen")

		}

		if f, err = r.NextFrame(); err != nil {
			t.Error("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ChannelOpen); !ok {
			t.Error("expected ChannelOpen")
		}

		if f.ChannelID() != 1 {
			t.Error("expected ChannelOpen on channel 1")
		}

		server.Close()
		done <- true
	}()

	c, err := amqp.NewClient(client, "guest", "guest", "/")
	if err != nil {
		t.Error("could not create client:", c, err)
	}

	println("X X X X X X X X X X X X X X X X X X X X Ohai")
	<-done
}
