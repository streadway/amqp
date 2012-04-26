package amqp

import (
	"amqp/wire"
	"fmt"
	"io"
)

type client struct {
	io  io.ReadWriter
	s2c chan wire.Frame
	c2s chan wire.Frame
	*connection
}

// When this returns, an AMPQ Connection and AMQP Channel will be opened,
// authenticated and ready to use.
//
// XXX(ST) NewClient will block until the wire handshake completes, is it expected in Go to have blocking constructors?
func NewClient(conn io.ReadWriter, username, password, vhost string) (me *client, err error) {
	me = &client{
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

	return
}

// Reads each frame off the IO and dispatches it to the right channel/connection
func (me *client) reader() {
	reader := wire.NewFrameReader(me.io)

	for {
		frame, err := reader.Read()
		fmt.Println("frame read: ", frame)

		if err != nil {
			return
			panic(fmt.Sprintf("TODO process io error by initiating a shutdown/reconnect", err))
		}

		switch frame.ChannelID() {
		case 0:
			me.connection.framing.s2c <- frame
		default:
			// TODO lookup and dispatch to the right Channel
		}
	}
}

func (me *client) writer() {
	for {
		frame := <-me.connection.framing.c2s
		// TODO handle when the chan closes
		fmt.Println("frame write: ", frame)
		_, err := frame.WriteTo(me.io)
		if err != nil {
			// TODO handle write failure to cleanly shutdown the connection
		}
	}
}
