package amqp_test

import (
	"amqp"
	"amqp/wire"
	"bytes"
	"testing"
)

func TestNewClientHandshake(t *testing.T) {
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

		if f, err = r.Read(); err != nil {
			t.Error("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionStartOk); !ok {
			t.Error("expected ConnectionStartOk")
		}

		if f, err = r.Read(); err != nil {
			t.Error("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionTuneOk); !ok {
			t.Error("expected ConnectionTuneOk")
		}

		if f, err = r.Read(); err != nil {
			t.Error("bad read", err)
		}

		if _, ok = f.(wire.MethodFrame).Method.(wire.ConnectionOpen); !ok {
			t.Error("expected ConnectionOpen")

		}

		server.Close()
	}()

	amqp.NewClient(client, "guest", "guest", "/")

}
