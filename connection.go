package amqp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
)

// Manages the serialization and deserialization of frames from IO and dispatches the frames to the appropriate channel.
type Connection struct {
	conn io.ReadWriteCloser

	in chan message

	state state

	writer *writer
	muw    sync.Mutex

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

// Accepts a string in the AMQP URI format specificer found at:
// http://www.rabbitmq.com/uri-spec.html and returns a new Connection
// to the peer over tcp using PlainAuth
// 
// Fallback values for the fields are:
//
// host:			localhost
// port:			5672
// usenrame:	guest
// password:	guest
// vhost:			/
func Dial(amqp string) (me *Connection, err error) {
	host, port, username, password, vhost := "localhost", 5672, "guest", "guest", "/"

	u, err := url.Parse(amqp)
	if err != nil {
		return
	}

	fmt.Sscanf(u.Host, "%s:%d", &host, &port)

	hostport := fmt.Sprintf("%s:%d", host, port)

	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	if u.Path != "" {
		vhost = u.Path
	}

	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		return
	}

	me, err = NewConnection(conn, &PlainAuth{username, password}, vhost)

	return
}

func NewConnection(conn io.ReadWriteCloser, auth *PlainAuth, vhost string) (me *Connection, err error) {
	me = &Connection{
		conn:     conn,
		writer:   &writer{bufio.NewWriter(conn)},
		in:       make(chan message),
		channels: make(map[uint16]*Channel),
	}

	go me.reader()

	return me, me.open(auth.Username, auth.Password, vhost)
}

func (me *Connection) nextChannelId() uint16 {
	me.increment.Lock()
	defer me.increment.Unlock()
	me.sequence++
	return me.sequence
}

func (me *Connection) Close() (err error) {
	if me.state != closed {
		if err = me.send(&methodFrame{
			ChannelId: 0,
			Method:    &connectionClose{ReplyCode: ReplySuccess, ReplyText: "bye"},
		}); err != nil {
			return
		}

		switch (<-me.in).(type) {
		case *connectionCloseOk:
			me.shutdown()
			return nil
		default:
			return ErrBadProtocol
		}
	}

	return ErrAlreadyClosed
}

func (me *Connection) dispatch() {
	for {
		switch (<-me.in).(type) {
		// handle the 4 way shutdown
		case *connectionClose:
			me.send(&methodFrame{
				ChannelId: 0,
				Method:    &connectionCloseOk{},
			})
			me.Close()
		case nil:
			// closed
			return
		}
	}
}

func (me *Connection) send(f frame) (err error) {
	//fmt.Println("send:", f)

	me.muw.Lock()
	defer me.muw.Unlock()

	if me.state > open {
		return ErrAlreadyClosed
	}

	if err = me.writer.WriteFrame(f); err != nil {
		// TODO handle write failure to cleanly shutdown the connection
		me.shutdown()
	}

	return nil
}

func (me *Connection) shutdown() {
	fmt.Println("connection shutdown", me.state)
	if me.state != closed {
		me.state = closing
		for i, c := range me.channels {
			delete(me.channels, i)
			c.shutdown()
		}
		close(me.in)
		me.conn.Close()
		me.state = closed
	}
}

// All methods sent to the connection channel should be synchronous so we
// can handle them directly without a framing component
func (me *Connection) demux(f frame) {
	if me.state != closed {
		if f.channel() == 0 {
			// TODO send hard error if any content frames/async frames are sent here
			switch mf := f.(type) {
			case *methodFrame:
				me.in <- mf.Method
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
}

// Reads each frame off the IO and hand off to the connection object that
// will demux the streams and dispatch to one of the opened channels or
// handle on channel 0 (the connection channel).
func (me *Connection) reader() {
	buf := bufio.NewReader(me.conn)
	frames := &reader{buf}

	for {
		frame, err := frames.ReadFrame()

		if err != nil {
			fmt.Println("err in ReadFrame:", frame, err)
			me.shutdown()
			return
		}

		me.demux(frame)
	}
}

// Constructs and opens a unique channel for concurrent operations
func (me *Connection) Channel() (channel *Channel, err error) {
	id := me.nextChannelId()
	channel, err = newChannel(me, id)
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
	me.state = handshaking

	if _, err = me.conn.Write(protocolHeader); err != nil {
		return
	}

	switch start := (<-me.in).(type) {
	case *connectionStart:
		me.VersionMajor = int(start.VersionMajor)
		me.VersionMinor = int(start.VersionMinor)
		me.Properties = Table(start.ServerProperties)

		if err = me.send(&methodFrame{
			ChannelId: 0,
			Method: &connectionStartOk{
				Mechanism: "PLAIN",
				Response:  fmt.Sprintf("\000%s\000%s", username, password),
			},
		}); err != nil {
			return
		}

		switch tune := (<-me.in).(type) {
		// TODO SECURE HANDSHAKE
		case *connectionTune:
			me.MaxChannels = int(tune.ChannelMax)
			me.HeartbeatInterval = int(tune.Heartbeat)
			me.MaxFrameSize = int(tune.FrameMax)

			if err = me.send(&methodFrame{
				ChannelId: 0,
				Method: &connectionTuneOk{
					ChannelMax: 10,
					FrameMax:   FrameMinSize,
					Heartbeat:  0,
				},
			}); err != nil {
				return
			}

			if err = me.send(&methodFrame{
				ChannelId: 0,
				Method: &connectionOpen{
					VirtualHost: vhost,
				},
			}); err != nil {
				return
			}

			switch (<-me.in).(type) {
			case *connectionOpenOk:
				me.state = open
				go me.dispatch()
				return nil
			case nil:
				return ErrBadVhost
			}
		case nil:
			return ErrBadCredentials
		}
	}
	return ErrBadProtocol
}
