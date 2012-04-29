package amqp

import (
	"amqp/wire"
	"fmt"
	"io"
)

type Client struct {
	io io.ReadWriter
	*connection
	*Channel
}

// When this returns, an AMPQ connection and AMQP Channel will be opened,
// authenticated and ready to use.
//
// XXX(ST) NewClient will block until the wire handshake completes, is it expected in Go to have blocking constructors?
func NewClient(conn io.ReadWriter, username, password, vhost string) (me *Client, err error) {
	me = &Client{
		io:         conn,
		connection: newConnection(make(chan wire.Frame), make(chan wire.Frame)),
	}

	go me.reader()
	go me.writer()

	if _, err = me.io.Write(wire.ProtocolHeader); err != nil {
		return
	}

	if err = me.connection.open(username, password, vhost); err != nil {
		return
	}

	channel, err := me.connection.OpenChannel()
	if err != nil {
		return
	}

	me.Channel = channel

	return
}

// Reads each frame off the IO and hand off to the connection object that
// will demux the streams and dispatch to one of the opened channels or
// handle on channel 0 (the connection channel).
func (me *Client) reader() {
	reader := wire.NewFrameReader(me.io)

	for {
		frame, err := reader.NextFrame()

		if err != nil {
			return
			panic(fmt.Sprintf("TODO process io error by initiating a shutdown/reconnect", err))
		}

		me.connection.s2c <- frame
	}
}

func (me *Client) writer() {
	for {
		frame := <-me.connection.framing.c2s
		// TODO handle when the chan closes
		_, err := frame.WriteTo(me.io)
		if err != nil {
			// TODO handle write failure to cleanly shutdown the connection
		}
	}
}
