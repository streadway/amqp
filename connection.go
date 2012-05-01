package amqp

import (
	"fmt"
	"io"
	"sync"
)

// Manages the serialization and deserialization of frames from IO and dispatches the frames to the appropriate channel.
type Connection struct {
	conn io.ReadWriteCloser

	messages chan Message

	c2s chan Frame

	major      int
	minor      int
	properties Table

	maxChannels       int
	maxFrameSize      int
	heartbeatInterval int

	m   sync.Mutex
	seq uint16

	channels map[uint16]*Channel
}

func NewConnection(conn io.ReadWriteCloser, auth *PlainAuth, vhost string) (me *Connection, err error) {
	me = &Connection{
		conn: conn,
		//msgs:  make(chan Message), // incoming synchronous connection methods (Channel == 0)
		//out:      make(chan Frame),  // shared chan that muxes all frames
		channels: make(map[uint16]*Channel),
	}

	go me.reader()
	go me.writer()

	return me, me.open(auth.Username, auth.Password, vhost)
}

func (me *Connection) nextSeq() uint16 {
	me.m.Lock()
	defer me.m.Unlock()
	me.seq++
	return me.seq
}

// All methods sent to the connection channel should be synchronous so we
// can handle them directly without a framing component
func (me *Connection) demux(frame Frame) {
	if frame.channel() == 0 {
		// TODO send hard error if any content frames/async frames are sent here
		switch mf := frame.(type) {
		case *MethodFrame:
			println(mf)
		default:
			panic("TODO close with hard-error")
		}
	} else {
		channel, ok := me.channels[frame.channel()]
		if ok {
			channel.s2c <- frame
		} else {
			// TODO handle unknown channel for now drop
			println("XXX unknown channel", frame.channel())
			panic("XXX unknown channel")
		}
	}
}

// Reads each frame off the IO and hand off to the connection object that
// will demux the streams and dispatch to one of the opened channels or
// handle on channel 0 (the connection channel).
func (me *Connection) reader() {
	frames := NewFramer(me.conn, me.conn)

	for {
		frame, err := frames.ReadFrame()

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

		err := frame.write(me.conn)

		if err != nil {
			// TODO handle write failure to cleanly shutdown the connection
		}
	}
}

// Only needs to be called if you want to interleave multiple publishers or
// multiple consumers on the same network Connection.  Each client comes with a
// opened embedded channel and exposes all channel related interfaces directly.
func (me *Connection) OpenChannel() (channel *Channel, err error) {
	id := me.nextSeq()
	channel, err = newChannel(id, me.maxFrameSize, make(chan Frame), me.c2s)
	me.channels[id] = channel
	//return channel, channel.open()
	return channel, nil
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
	if _, err = me.conn.Write(ProtocolHeader); err != nil {
		return
	}

	switch start := (<-me.messages).(type) {
	case *ConnectionStart:
		me.major = int(start.VersionMajor)
		me.minor = int(start.VersionMinor)
		me.properties = Table(start.ServerProperties)

		me.c2s <- &MethodFrame{
			ChannelId: 0,
			Method: &ConnectionStartOk{
				Mechanism: "PLAIN",
				Response:  fmt.Sprintf("\000%s\000%s", username, password),
			},
		}

		switch tune := (<-me.messages).(type) {
		// TODO SECURE HANDSHAKE
		case *ConnectionTune:
			me.maxChannels = int(tune.ChannelMax)
			me.heartbeatInterval = int(tune.Heartbeat)
			me.maxFrameSize = int(tune.FrameMax)

			me.c2s <- &MethodFrame{
				ChannelId: 0,
				Method: &ConnectionTuneOk{
					ChannelMax: 10,
					FrameMax:   FrameMinSize,
					Heartbeat:  0,
				},
			}

			me.c2s <- &MethodFrame{
				ChannelId: 0,
				Method: &ConnectionOpen{
					VirtualHost: vhost,
				},
			}

			switch (<-me.messages).(type) {
			case *ConnectionOpenOk:
				return nil
			}
		}
	}
	return ErrBadProtocol
}
