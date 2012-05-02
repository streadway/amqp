package amqp

import (
	"fmt"
	"io"
	"sync"
)

// Manages the serialization and deserialization of frames from IO and dispatches the frames to the appropriate channel.
type Connection struct {
	conn io.ReadWriteCloser

	messages chan message

	out chan frame

	VersionMajor int
	VersionMinor int
	Properties   Table

	MaxChannels       int
	MaxFrameSize      int
	HeartbeatInterval int

	increment sync.Mutex
	sequence  uint16

	channels map[uint16]*Channel
}

func NewConnection(conn io.ReadWriteCloser, auth *PlainAuth, vhost string) (me *Connection, err error) {
	me = &Connection{
		conn: conn,
		//msgs:  make(chan message), // incoming synchronous connection methods (Channel == 0)
		//out:      make(chan frame),  // shared chan that muxes all frames
		channels: make(map[uint16]*Channel),
	}

	go me.reader()
	go me.writer()

	return me, me.open(auth.Username, auth.Password, vhost)
}

func (me *Connection) nextChannelId() uint16 {
	me.increment.Lock()
	defer me.increment.Unlock()
	me.sequence++
	return me.sequence
}

func (me *Connection) send(f frame) error {
	me.out <- f
	return nil
}

// All methods sent to the connection channel should be synchronous so we
// can handle them directly without a framing component
func (me *Connection) demux(f frame) {
	if f.channel() == 0 {
		// TODO send hard error if any content frames/async frames are sent here
		switch mf := f.(type) {
		case *methodFrame:
			me.messages <- mf.Method
		default:
			panic("TODO close with hard-error")
		}
	} else {
		channel, ok := me.channels[f.channel()]
		if ok {
			channel.recv(channel, f)
		} else {
			// TODO handle unknown channel for now drop
			println("XXX unknown channel", f.channel())
			panic("XXX unknown channel")
		}
	}
}

// Reads each frame off the IO and hand off to the connection object that
// will demux the streams and dispatch to one of the opened channels or
// handle on channel 0 (the connection channel).
func (me *Connection) reader() {
	frames := &reader{me.conn}

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
		frame := <-me.out
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
	id := me.nextChannelId()
	channel, err = newChannel(me, id)
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
	case *connectionStart:
		me.VersionMajor = int(start.VersionMajor)
		me.VersionMinor = int(start.VersionMinor)
		me.Properties = Table(start.ServerProperties)

		me.out <- &methodFrame{
			ChannelId: 0,
			Method: &connectionStartOk{
				Mechanism: "PLAIN",
				Response:  fmt.Sprintf("\000%s\000%s", username, password),
			},
		}

		switch tune := (<-me.messages).(type) {
		// TODO SECURE HANDSHAKE
		case *connectionTune:
			me.MaxChannels = int(tune.ChannelMax)
			me.HeartbeatInterval = int(tune.Heartbeat)
			me.MaxFrameSize = int(tune.FrameMax)

			me.out <- &methodFrame{
				ChannelId: 0,
				Method: &connectionTuneOk{
					ChannelMax: 10,
					FrameMax:   FrameMinSize,
					Heartbeat:  0,
				},
			}

			me.out <- &methodFrame{
				ChannelId: 0,
				Method: &connectionOpen{
					VirtualHost: vhost,
				},
			}

			switch (<-me.messages).(type) {
			case *connectionOpenOk:
				return nil
			}
		}
	}
	return ErrBadProtocol
}
