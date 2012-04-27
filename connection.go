package amqp

import (
	"amqp/wire"
	"fmt"
)

// The logical AMQP Connection class that handles opening and closing channels
type connection struct {
	framing *Framing

	s2c chan wire.Frame

	major      int
	minor      int
	properties Table

	maxChannels       int
	maxFrameSize      int
	heartbeatInterval int

	seq      chan uint16
	channels map[uint16]*channel
}

// XXX(ST) how/when does this get GC'd?  Better solutions must exist.
func sequence(i uint16, c chan uint16) {
	for {
		c <- i
		i++
	}
}

func newConnection(s2c, c2s chan wire.Frame) (me *connection) {
	seq := make(chan uint16)
	me = &connection{
		framing:  newFraming(0, wire.FrameMinSize, make(chan wire.Frame), c2s),
		s2c:      s2c, // shared chan across all channels including the connection
		seq:      seq,
		channels: make(map[uint16]*channel),
	}
	go sequence(1, seq)
	go me.demux()

	return me
}

func (me *connection) demux() {
	// TODO termination condition
	for {
		frame, ok := <-me.s2c
		if ok {
			if frame.ChannelID() == 0 {
				me.framing.s2c <- frame
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
		} else {
			// TODO handle closed connection
			return
		}
	}

}

// Only needs to be called if you want to interleave multiple publishers or
// multiple consumers on the same network connection.  Each client comes with a
// opened embedded channel and exposes all channel related interfaces directly.
func (me *connection) OpenChannel() (channel *channel, err error) {
	id := uint16(<-me.seq)
	framing := newFraming(id, me.maxFrameSize, make(chan wire.Frame), me.framing.c2s)

	if channel, err = newChannel(framing); err != nil {
		return
	}

	me.channels[id] = channel
	return channel, channel.open()
}

//    connection          = open-connection *use-connection close-connection
//    open-connection     = C:protocol-header
//                          S:START C:START-OK
//                          *challenge
//                          S:TUNE C:TUNE-OK
//                          C:OPEN S:OPEN-OK
//    challenge           = S:SECURE C:SECURE-OK
//    use-connection      = *channel
//    close-connection    = C:CLOSE S:CLOSE-OK
//                        / S:CLOSE C:CLOSE-OK
func (me *connection) open(username, password, vhost string) error {
	switch start := me.framing.Recv().Method.(type) {
	case wire.ConnectionStart:
		me.major = int(start.VersionMajor)
		me.minor = int(start.VersionMinor)
		me.properties = Table(start.ServerProperties)

		me.framing.Send(Message{
			Method: wire.ConnectionStartOk{
				Mechanism: "PLAIN",
				Response:  fmt.Sprintf("\000%s\000%s", username, password),
			},
		})

		switch tune := me.framing.Recv().Method.(type) {
		// TODO SECURE HANDSHAKE
		case wire.ConnectionTune:
			me.maxChannels = int(tune.ChannelMax)
			me.heartbeatInterval = int(tune.Heartbeat)
			me.maxFrameSize = int(tune.FrameMax)

			me.framing.Send(Message{
				Method: wire.ConnectionTuneOk{
					ChannelMax: 10,
					FrameMax:   wire.FrameMinSize,
					Heartbeat:  0,
				},
			})

			me.framing.Send(Message{
				Method: wire.ConnectionOpen{
					VirtualHost: vhost,
				},
			})

			switch me.framing.Recv().Method.(type) {
			case wire.ConnectionOpenOk:
				return nil
			}
		}
	}
	return ErrBadProtocol
}
