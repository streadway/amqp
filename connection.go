package amqp

import (
	"amqp/wire"
	"fmt"
	"io"
)

// Manages the serialization and deserialization of frames from IO and dispatches the frames to the appropriate channel.
type Connection struct {
	conn    io.ReadWriteCloser
	c2s     chan wire.Frame
	methods chan wire.Method

	major      int
	minor      int
	properties Table

	maxChannels       int
	maxFrameSize      int
	heartbeatInterval int

	seq chan uint16

	channels map[uint16]*Channel
}

// XXX(ST) how/when does this get GC'd?  Better solutions must exist.
func sequence(i uint16, c chan uint16) {
	for {
		c <- i
		i++
	}
}

func NewConnection(conn io.ReadWriteCloser, auth *PlainAuth, vhost string) (me *Connection, err error) {
	seq := make(chan uint16)

	me = &Connection{
		conn:     conn,
		methods:  make(chan wire.Method), // incoming synchronous connection methods (Channel == 0)
		c2s:      make(chan wire.Frame),  // shared chan that muxes all frames
		seq:      seq,
		channels: make(map[uint16]*Channel),
	}

	go sequence(1, seq)

	go me.reader()
	go me.writer()

	return me, me.open(auth.Username, auth.Password, vhost)
}

// All methods sent to the connection channel should be synchronous so we
// can handle them directly without a framing component
func (me *Connection) demux(frame wire.Frame) {
	if frame.ChannelID() == 0 {
		// TODO send hard error if any content frames/async frames are sent here
		switch mf := frame.(type) {
		case wire.MethodFrame:
			me.methods <- mf.Method
		default:
			panic("TODO close with hard-error")
		}
	} else {
		channel, ok := me.channels[frame.ChannelID()]
		if ok {
			channel.framing.s2c <- frame
		} else {
			// TODO handle unknown channel for now drop
			println("XXX unknown channel", frame.ChannelID())
			panic("XXX unknown channel")
		}
	}
}

// Reads each frame off the IO and hand off to the connection object that
// will demux the streams and dispatch to one of the opened channels or
// handle on channel 0 (the connection channel).
func (me *Connection) reader() {
	frames := wire.NewFrameReader(me.conn)

	for {
		frame, err := frames.NextFrame()

		if err != nil {
			return
			panic(fmt.Sprintf("TODO process io error by initiating a shutdown/reconnect", err))
		}

		me.demux(frame)
	}
}

func (me *Connection) writer() {
	for {
		frame := <-me.c2s
		if frame == nil {
			// TODO handle when the chan closes
			return
		}

		_, err := frame.WriteTo(me.conn)
		if err != nil {
			// TODO handle write failure to cleanly shutdown the connection
		}
	}
}

// Only needs to be called if you want to interleave multiple publishers or
// multiple consumers on the same network Connection.  Each client comes with a
// opened embedded channel and exposes all channel related interfaces directly.
func (me *Connection) OpenChannel() (channel *Channel, err error) {
	id := uint16(<-me.seq)

	framing := newFraming(id, me.maxFrameSize, make(chan wire.Frame), me.c2s)

	if channel, err = newChannel(framing); err != nil {
		return
	}

	me.channels[id] = channel

	return channel, channel.open()
}

//    Connection          = open-Connection *use-Connection close-Connection
//    open-Connection     = C:protocol-header
//                          S:START C:START-OK
//                          *challenge
//                          S:TUNE C:TUNE-OK
//                          C:OPEN S:OPEN-OK
//    challenge           = S:SECURE C:SECURE-OK
//    use-Connection      = *channel
//    close-Connection    = C:CLOSE S:CLOSE-OK
//                        / S:CLOSE C:CLOSE-OK
func (me *Connection) open(username, password, vhost string) (err error) {
	if _, err = me.conn.Write(wire.ProtocolHeader); err != nil {
		return
	}

	switch start := (<-me.methods).(type) {
	case wire.ConnectionStart:
		me.major = int(start.VersionMajor)
		me.minor = int(start.VersionMinor)
		me.properties = Table(start.ServerProperties)

		me.c2s <- wire.MethodFrame{
			Channel: 0,
			Method: wire.ConnectionStartOk{
				Mechanism: "PLAIN",
				Response:  fmt.Sprintf("\000%s\000%s", username, password),
			},
		}

		switch tune := (<-me.methods).(type) {
		// TODO SECURE HANDSHAKE
		case wire.ConnectionTune:
			me.maxChannels = int(tune.ChannelMax)
			me.heartbeatInterval = int(tune.Heartbeat)
			me.maxFrameSize = int(tune.FrameMax)

			me.c2s <- wire.MethodFrame{
				Channel: 0,
				Method: wire.ConnectionTuneOk{
					ChannelMax: 10,
					FrameMax:   wire.FrameMinSize,
					Heartbeat:  0,
				},
			}

			me.c2s <- wire.MethodFrame{
				Channel: 0,
				Method: wire.ConnectionOpen{
					VirtualHost: vhost,
				},
			}

			switch (<-me.methods).(type) {
			case wire.ConnectionOpenOk:
				return nil
			}
		}
	}
	return ErrBadProtocol
}
