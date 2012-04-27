package amqp

import (
	"amqp/wire"
)

// Represents an AMQP channel, used for concurrent, interleaved publishers and
// consumers on the same connection.
type channel struct {
	framing *Framing
}

// Constructs and opens a new channel with the given framing rules
func newChannel(framing *Framing) (me *channel, err error) {
	return &channel{
		framing: framing,
	}, nil
}

//    channel             = open-channel *use-channel close-channel
//    open-channel        = C:OPEN S:OPEN-OK
//    use-channel         = C:FLOW S:FLOW-OK
//                        / S:FLOW C:FLOW-OK
//                        / functional-class
//    close-channel       = C:CLOSE S:CLOSE-OK
//                        / S:CLOSE C:CLOSE-OK
func (me *channel) open() error {
	me.framing.Send(Message{
		Method: wire.ChannelOpen{},
	})

	switch me.framing.Recv().Method.(type) {
	case wire.ChannelOpenOk:
		return nil
	}

	// TODO handle channel open errors (like already opened on this ID)
	return ErrBadProtocol
}
