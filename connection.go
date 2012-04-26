package amqp

import (
	"amqp/wire"
	"fmt"
)

// The logical AMQP Connection class that handles opening and closing channels
type connection struct {
	framing *Framing

	major      int
	minor      int
	properties Table

	maxChannels       int
	maxFrameSize      int
	heartbeatInterval int
}

func newConnection(s2c, c2s chan wire.Frame) (me *connection) {
	return &connection{
		framing: newFraming(0, wire.FrameMinSize, s2c, c2s),
	}
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
			me.maxFrameSize = int(tune.FrameMax)
			me.heartbeatInterval = int(tune.Heartbeat)

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
