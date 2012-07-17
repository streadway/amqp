// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

// Used along side NewConnection to specify the desired tuning parameters used
// during a Connection handshake.  The negotiated tuning will be stored in the
// resultant connection.
type Config struct {
	// The SASL mechanisms to try in the client request, and the successful
	// mechanism used on the Connection object
	SASL []Authentication

	Vhost string // The Vhost the Auth credentials are permitted to open

	MaxChannels       int // Maximum number of channels the client intends to open - defaults to 0
	MaxFrameSize      int // Maximum frame size the client intends to send - default to 0
	HeartbeatInterval int // Frequency in seconds the client wishes the server to send heartbeats - defaults to 0 (no heartbeats)
}

// Manages the serialization and deserialization of frames from IO and dispatches the frames to the appropriate channel.
type Connection struct {
	conn io.ReadWriteCloser

	in            chan message
	shutdownMutex sync.Mutex

	state state // TODO not goroutine-safe; refactor with stateMutex?

	writer      *writer
	writerMutex sync.Mutex

	increment sync.Mutex
	sequence  uint16

	channels map[uint16]*Channel

	Config Config // The negotiated Config after connection.open

	VersionMajor int   // The server's major version
	VersionMinor int   // The server's minor version
	Properties   Table // Server properties
}

// Dial accepts a string in the AMQP URI format, and returns a new Connection
// over TCP using PlainAuth.
func Dial(amqp string) (*Connection, error) {
	uri, err := ParseURI(amqp)
	if err != nil {
		return nil, err
	}

	addr := net.JoinHostPort(uri.Host, strconv.FormatInt(int64(uri.Port), 10))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return NewConnection(conn, Config{
		SASL:  []Authentication{uri.PlainAuth()},
		Vhost: uri.Vhost,
	})
}

func NewConnection(conn io.ReadWriteCloser, config Config) (me *Connection, err error) {
	me = &Connection{
		conn:     conn,
		writer:   &writer{bufio.NewWriter(conn)},
		in:       make(chan message),
		channels: make(map[uint16]*Channel),
	}

	go me.reader()

	return me, me.open(config)
}

func (me *Connection) nextChannelId() uint16 {
	me.increment.Lock()
	defer me.increment.Unlock()
	me.sequence++
	return me.sequence
}

func (me *Connection) Close() (err error) {
	if me.state != closed {
		if err = me.send(
			&methodFrame{
				ChannelId: 0,
				Method: &connectionClose{
					ReplyCode: ReplySuccess,
					ReplyText: "bye",
				},
			},
		); err != nil {
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

		switch msg := <-me.in; msg.(type) {
		// handle the 4 way shutdown
		case *connectionClose: // request from server
			me.send(&methodFrame{
				ChannelId: 0,
				Method:    &connectionCloseOk{},
			})
			me.Close()
		case *connectionCloseOk: // response to our Close() request
			me.in <- msg // forward to Close() method
			return
		case nil: // closed
			return
		}
	}
}

func (me *Connection) send(f frame) (err error) {
	//fmt.Println("send:", f)

	me.writerMutex.Lock()
	defer me.writerMutex.Unlock()

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
	me.shutdownMutex.Lock()
	defer me.shutdownMutex.Unlock()

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
			me.shutdown()
			return
		}

		me.demux(frame)
	}
}

// Convienence method to inspect the Connection.Properties["capabilities"]
// Table for server identified capabilities like "basic.ack" or
// "confirm.select".
func (me *Connection) IsCapable(featureName string) bool {
	if me.Properties != nil {
		if v, ok := me.Properties["capabilities"]; ok {
			if capabilities, ok := v.(Table); ok {
				if feature, ok := capabilities[featureName]; ok {
					if has, ok := feature.(bool); ok && has {
						return true
					}
				}
			}
		}
	}
	return false
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
func (me *Connection) open(config Config) (err error) {
	me.state = handshaking

	if _, err = me.conn.Write(protocolHeader); err != nil {
		return
	}

	return me.openStart(config)
}

func (me *Connection) openStart(config Config) (err error) {
	switch start := (<-me.in).(type) {
	case *connectionStart:
		me.VersionMajor = int(start.VersionMajor)
		me.VersionMinor = int(start.VersionMinor)
		me.Properties = Table(start.ServerProperties)

		// TODO support challenge/response
		auth, ok := pickSASLMechanism(config.SASL, strings.Split(start.Mechanisms, " "))
		if !ok {
			return ErrUnsupportedMechanism
		}

		// Save this mechanism off as the one we chose
		me.Config.SASL = []Authentication{auth}

		if err = me.send(&methodFrame{
			ChannelId: 0,
			Method: &connectionStartOk{
				Mechanism: auth.Mechanism(),
				Response:  auth.Response(),
			},
		}); err != nil {
			return
		}

		return me.openTune(config)
	case nil:
		return ErrBadProtocol
	}
	return ErrBadProtocol
}

func (me *Connection) openTune(config Config) (err error) {
	switch tune := (<-me.in).(type) {
	// TODO SECURE HANDSHAKE
	case *connectionTune:
		// When this is bounded, share the bound.  We're effectively only bounded
		// by MaxUint16.  If you hit a wrap around bug, throw a small party then
		// make an github issue.
		if int(tune.ChannelMax) > 0 {
			me.Config.MaxChannels = int(tune.ChannelMax)
		}

		// Frame size includes headers and end byte (len(payload)+8), even if
		// this is less than FrameMinSize, use what the server sends because the
		// alternative is to stop the handshake here.
		if int(tune.FrameMax) > 0 {
			if config.MaxFrameSize <= 0 || int(tune.FrameMax) < config.MaxFrameSize {
				me.Config.MaxFrameSize = int(tune.FrameMax)
			} else {
				me.Config.MaxFrameSize = config.MaxFrameSize
			}
		} else {
			if config.MaxFrameSize > 0 {
				me.Config.MaxFrameSize = config.MaxFrameSize
			} else {
				// Client and Server are unlimited.  We'll bound the frame size to
				// 256KB to fit within the bandwidth delay product of a satellite
				// uplink on a scaled TCP window up to 2Mb/s
				me.Config.MaxFrameSize = 256 * 1024
			}
		}

		// This the interval the server wishes us to send at.  config.Heartbeat
		// is the interval what the client wishes the server to send at.
		me.Config.HeartbeatInterval = int(tune.Heartbeat)

		if err = me.send(&methodFrame{
			ChannelId: 0,
			Method: &connectionTuneOk{
				ChannelMax: uint16(me.Config.MaxChannels),
				FrameMax:   uint32(me.Config.MaxFrameSize),
				Heartbeat:  uint16(config.HeartbeatInterval),
			},
		}); err != nil {
			return
		}

		return me.openVhost(config)
	case nil:
		return ErrBadCredentials
	}
	return ErrBadProtocol
}

func (me *Connection) openVhost(config Config) (err error) {
	me.Config.Vhost = config.Vhost

	if err = me.send(&methodFrame{
		ChannelId: 0,
		Method: &connectionOpen{
			VirtualHost: me.Config.Vhost,
		},
	}); err != nil {
		return
	}

	switch (<-me.in).(type) {
	case *connectionOpenOk:
		me.state = open
		go me.dispatch()
		return nil // connection.open handshake finished
	case nil:
		return ErrBadVhost
	}

	return ErrBadProtocol
}
