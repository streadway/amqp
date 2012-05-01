/* GENERATED FILE - DO NOT EDIT */
/* Rebuild from the protocol/gen.go tool */

package amqp

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	FrameMethod    = 1
	FrameHeader    = 2
	FrameBody      = 3
	FrameHeartbeat = 8
	FrameMinSize   = 4096
	FrameEnd       = 206

	/* 
	   Indicates that the method completed successfully. This reply code is
	   reserved for future use - the current protocol design does not use positive
	   confirmation and reply codes are sent only in case of an error.
	*/
	ReplySuccess = 200

	/* 
	   The client attempted to transfer content larger than the server could accept
	   at the present time. The client may retry at a later time.
	*/
	ContentTooLarge = 311

	/* 
	   When the exchange cannot deliver to a consumer when the immediate flag is
	   set. As a result of pending data on the queue or the absence of any
	   consumers of the queue.
	*/
	NoConsumers = 313

	/* 
	   An operator intervened to close the connection for some reason. The client
	   may retry at some later date.
	*/
	ConnectionForced = 320

	/* 
	   The client tried to work with an unknown virtual host.
	*/
	InvalidPath = 402

	/* 
	   The client attempted to work with a server entity to which it has no
	   access due to security settings.
	*/
	AccessRefused = 403

	/* 
	   The client attempted to work with a server entity that does not exist.
	*/
	NotFound = 404

	/* 
	   The client attempted to work with a server entity to which it has no
	   access because another client is working with it.
	*/
	ResourceLocked = 405

	/* 
	   The client requested a method that was not allowed because some precondition
	   failed.
	*/
	PreconditionFailed = 406

	/* 
	   The sender sent a malformed frame that the recipient could not decode.
	   This strongly implies a programming error in the sending peer.
	*/
	FrameError = 501

	/* 
	   The sender sent a frame that contained illegal values for one or more
	   fields. This strongly implies a programming error in the sending peer.
	*/
	SyntaxError = 502

	/* 
	   The client sent an invalid sequence of frames, attempting to perform an
	   operation that was considered invalid by the server. This usually implies
	   a programming error in the client.
	*/
	CommandInvalid = 503

	/* 
	   The client attempted to work with a channel that had not been correctly
	   opened. This most likely indicates a fault in the client layer.
	*/
	ChannelError = 504

	/* 
	   The peer sent a frame that was not expected, usually in the context of
	   a content header and body.  This strongly indicates a fault in the peer's
	   content processing.
	*/
	UnexpectedFrame = 505

	/* 
	   The server could not complete the method because it lacked sufficient
	   resources. This may be due to the client creating too many of some type
	   of entity.
	*/
	ResourceError = 506

	/* 
	   The client tried to work with some entity in a manner that is prohibited
	   by the server, due to security settings or by some other criteria.
	*/
	NotAllowed = 530

	/* 
	   The client tried to use functionality that is not implemented in the
	   server.
	*/
	NotImplemented = 540

	/* 
	   The server could not complete the method because of an internal error.
	   The server may require intervention by an operator in order to resume
	   normal operations.
	*/
	InternalError = 541
)

/*  
   This method starts the connection negotiation process by telling the client the
   protocol version that the server proposes, along with a list of security mechanisms
   which the client can use for authentication.
*/
type ConnectionStart struct {
	VersionMajor     byte   // protocol major version
	VersionMinor     byte   // protocol minor version
	ServerProperties Table  // server properties
	Mechanisms       string // available security mechanisms
	Locales          string // available message locales

}

func (me *ConnectionStart) id() (uint16, uint16) {
	return 10, 10
}

func (me *ConnectionStart) wait() bool {
	return true
}

func (me *ConnectionStart) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.VersionMajor); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, me.VersionMinor); err != nil {
		return
	}

	if err = writeTable(w, me.ServerProperties); err != nil {
		return
	}

	return
}

func (me *ConnectionStart) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.VersionMajor); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &me.VersionMinor); err != nil {
		return
	}

	if me.ServerProperties, err = readTable(r); err != nil {
		return
	}

	return
}

/*  
   This method selects a SASL security mechanism.
*/
type ConnectionStartOk struct {
	ClientProperties Table  // client properties
	Mechanism        string // selected security mechanism
	Response         string // security response data
	Locale           string // selected message locale

}

func (me *ConnectionStartOk) id() (uint16, uint16) {
	return 10, 11
}

func (me *ConnectionStartOk) wait() bool {
	return true
}

func (me *ConnectionStartOk) write(w io.Writer) (err error) {

	if err = writeTable(w, me.ClientProperties); err != nil {
		return
	}

	if err = writeShortstr(w, me.Mechanism); err != nil {
		return
	}

	if err = writeLongstr(w, me.Response); err != nil {
		return
	}

	return
}

func (me *ConnectionStartOk) read(r io.Reader) (err error) {

	if me.ClientProperties, err = readTable(r); err != nil {
		return
	}

	if me.Mechanism, err = readShortstr(r); err != nil {
		return
	}

	if me.Response, err = readLongstr(r); err != nil {
		return
	}

	return
}

/*  
   The SASL protocol works by exchanging challenges and responses until both peers have
   received sufficient information to authenticate each other. This method challenges
   the client to provide more information.
*/
type ConnectionSecure struct {
	Challenge string // security challenge data

}

func (me *ConnectionSecure) id() (uint16, uint16) {
	return 10, 20
}

func (me *ConnectionSecure) wait() bool {
	return true
}

func (me *ConnectionSecure) write(w io.Writer) (err error) {

	return
}

func (me *ConnectionSecure) read(r io.Reader) (err error) {

	return
}

/*  
   This method attempts to authenticate, passing a block of SASL data for the security
   mechanism at the server side.
*/
type ConnectionSecureOk struct {
	Response string // security response data

}

func (me *ConnectionSecureOk) id() (uint16, uint16) {
	return 10, 21
}

func (me *ConnectionSecureOk) wait() bool {
	return true
}

func (me *ConnectionSecureOk) write(w io.Writer) (err error) {

	return
}

func (me *ConnectionSecureOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method proposes a set of connection configuration values to the client. The
   client can accept and/or adjust these.
*/
type ConnectionTune struct {
	ChannelMax uint16 // proposed maximum channels
	FrameMax   uint32 // proposed maximum frame size
	Heartbeat  uint16 // desired heartbeat delay

}

func (me *ConnectionTune) id() (uint16, uint16) {
	return 10, 30
}

func (me *ConnectionTune) wait() bool {
	return true
}

func (me *ConnectionTune) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, me.FrameMax); err != nil {
		return
	}

	return
}

func (me *ConnectionTune) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &me.FrameMax); err != nil {
		return
	}

	return
}

/*  
   This method sends the client's connection tuning parameters to the server.
   Certain fields are negotiated, others provide capability information.
*/
type ConnectionTuneOk struct {
	ChannelMax uint16 // negotiated maximum channels
	FrameMax   uint32 // negotiated maximum frame size
	Heartbeat  uint16 // desired heartbeat delay

}

func (me *ConnectionTuneOk) id() (uint16, uint16) {
	return 10, 31
}

func (me *ConnectionTuneOk) wait() bool {
	return true
}

func (me *ConnectionTuneOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, me.FrameMax); err != nil {
		return
	}

	return
}

func (me *ConnectionTuneOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &me.FrameMax); err != nil {
		return
	}

	return
}

/*  
   This method opens a connection to a virtual host, which is a collection of
   resources, and acts to separate multiple application domains within a server.
   The server may apply arbitrary limits per virtual host, such as the number
   of each type of entity that may be used, per connection and/or in total.
*/
type ConnectionOpen struct {
	VirtualHost string // virtual host name
	reserved1   string
	reserved2   bool
}

func (me *ConnectionOpen) id() (uint16, uint16) {
	return 10, 40
}

func (me *ConnectionOpen) wait() bool {
	return true
}

func (me *ConnectionOpen) write(w io.Writer) (err error) {

	if err = writeShortstr(w, me.VirtualHost); err != nil {
		return
	}
	if err = writeShortstr(w, me.reserved1); err != nil {
		return
	}

	return
}

func (me *ConnectionOpen) read(r io.Reader) (err error) {

	if me.VirtualHost, err = readShortstr(r); err != nil {
		return
	}
	if me.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method signals to the client that the connection is ready for use.
*/
type ConnectionOpenOk struct {
	reserved1 string
}

func (me *ConnectionOpenOk) id() (uint16, uint16) {
	return 10, 41
}

func (me *ConnectionOpenOk) wait() bool {
	return true
}

func (me *ConnectionOpenOk) write(w io.Writer) (err error) {

	return
}

func (me *ConnectionOpenOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method indicates that the sender wants to close the connection. This may be
   due to internal conditions (e.g. a forced shut-down) or due to an error handling
   a specific method, i.e. an exception. When a close is due to an exception, the
   sender provides the class and method id of the method which caused the exception.
*/
type ConnectionClose struct {
	ReplyCode uint16
	ReplyText string
	ClassId   uint16 // failing method class
	MethodId  uint16 // failing method ID

}

func (me *ConnectionClose) id() (uint16, uint16) {
	return 10, 50
}

func (me *ConnectionClose) wait() bool {
	return true
}

func (me *ConnectionClose) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, me.ReplyText); err != nil {
		return
	}

	return
}

func (me *ConnectionClose) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.ReplyCode); err != nil {
		return
	}

	if me.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method confirms a Connection.Close method and tells the recipient that it is
   safe to release resources for the connection and close the socket.
*/
type ConnectionCloseOk struct {
}

func (me *ConnectionCloseOk) id() (uint16, uint16) {
	return 10, 51
}

func (me *ConnectionCloseOk) wait() bool {
	return true
}

func (me *ConnectionCloseOk) write(w io.Writer) (err error) {

	return
}

func (me *ConnectionCloseOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method opens a channel to the server.
*/
type ChannelOpen struct {
	reserved1 string
}

func (me *ChannelOpen) id() (uint16, uint16) {
	return 20, 10
}

func (me *ChannelOpen) wait() bool {
	return true
}

func (me *ChannelOpen) write(w io.Writer) (err error) {

	return
}

func (me *ChannelOpen) read(r io.Reader) (err error) {

	return
}

/*  
   This method signals to the client that the channel is ready for use.
*/
type ChannelOpenOk struct {
	reserved1 string
}

func (me *ChannelOpenOk) id() (uint16, uint16) {
	return 20, 11
}

func (me *ChannelOpenOk) wait() bool {
	return true
}

func (me *ChannelOpenOk) write(w io.Writer) (err error) {

	return
}

func (me *ChannelOpenOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method asks the peer to pause or restart the flow of content data sent by
   a consumer. This is a simple flow-control mechanism that a peer can use to avoid
   overflowing its queues or otherwise finding itself receiving more messages than
   it can process. Note that this method is not intended for window control. It does
   not affect contents returned by Basic.Get-Ok methods.
*/
type ChannelFlow struct {
	Active bool // start/stop content frames

}

func (me *ChannelFlow) id() (uint16, uint16) {
	return 20, 20
}

func (me *ChannelFlow) wait() bool {
	return true
}

func (me *ChannelFlow) write(w io.Writer) (err error) {

	return
}

func (me *ChannelFlow) read(r io.Reader) (err error) {

	return
}

/*  
   Confirms to the peer that a flow command was received and processed.
*/
type ChannelFlowOk struct {
	Active bool // current flow setting

}

func (me *ChannelFlowOk) id() (uint16, uint16) {
	return 20, 21
}

func (me *ChannelFlowOk) wait() bool {
	return false
}

func (me *ChannelFlowOk) write(w io.Writer) (err error) {

	return
}

func (me *ChannelFlowOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method indicates that the sender wants to close the channel. This may be due to
   internal conditions (e.g. a forced shut-down) or due to an error handling a specific
   method, i.e. an exception. When a close is due to an exception, the sender provides
   the class and method id of the method which caused the exception.
*/
type ChannelClose struct {
	ReplyCode uint16
	ReplyText string
	ClassId   uint16 // failing method class
	MethodId  uint16 // failing method ID

}

func (me *ChannelClose) id() (uint16, uint16) {
	return 20, 40
}

func (me *ChannelClose) wait() bool {
	return true
}

func (me *ChannelClose) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, me.ReplyText); err != nil {
		return
	}

	return
}

func (me *ChannelClose) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.ReplyCode); err != nil {
		return
	}

	if me.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method confirms a Channel.Close method and tells the recipient that it is safe
   to release resources for the channel.
*/
type ChannelCloseOk struct {
}

func (me *ChannelCloseOk) id() (uint16, uint16) {
	return 20, 41
}

func (me *ChannelCloseOk) wait() bool {
	return true
}

func (me *ChannelCloseOk) write(w io.Writer) (err error) {

	return
}

func (me *ChannelCloseOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method creates an exchange if it does not already exist, and if the exchange
   exists, verifies that it is of the correct and expected class.
*/
type ExchangeDeclare struct {
	reserved1  uint16
	Exchange   string
	Type       string // exchange type
	Passive    bool   // do not create exchange
	Durable    bool   // request a durable exchange
	AutoDelete bool   // auto-delete when unused
	Internal   bool   // create internal exchange
	NoWait     bool
	Arguments  Table // arguments for declaration

}

func (me *ExchangeDeclare) id() (uint16, uint16) {
	return 40, 10
}

func (me *ExchangeDeclare) wait() bool {
	return true && !me.NoWait
}

func (me *ExchangeDeclare) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, me.Type); err != nil {
		return
	}

	var bits byte

	if me.Passive {
		bits |= 1 << 0
	}

	if me.Durable {
		bits |= 1 << 1
	}

	if me.AutoDelete {
		bits |= 1 << 2
	}

	if me.Internal {
		bits |= 1 << 3
	}

	if me.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (me *ExchangeDeclare) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if me.Type, err = readShortstr(r); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.Passive = (bits&(1<<0) > 0)
	me.Durable = (bits&(1<<1) > 0)
	me.AutoDelete = (bits&(1<<2) > 0)
	me.Internal = (bits&(1<<3) > 0)
	me.NoWait = (bits&(1<<4) > 0)

	return
}

/*  
   This method confirms a Declare method and confirms the name of the exchange,
   essential for automatically-named exchanges.
*/
type ExchangeDeclareOk struct {
}

func (me *ExchangeDeclareOk) id() (uint16, uint16) {
	return 40, 11
}

func (me *ExchangeDeclareOk) wait() bool {
	return true
}

func (me *ExchangeDeclareOk) write(w io.Writer) (err error) {

	return
}

func (me *ExchangeDeclareOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method deletes an exchange. When an exchange is deleted all queue bindings on
   the exchange are cancelled.
*/
type ExchangeDelete struct {
	reserved1 uint16
	Exchange  string
	IfUnused  bool // delete only if unused
	NoWait    bool
}

func (me *ExchangeDelete) id() (uint16, uint16) {
	return 40, 20
}

func (me *ExchangeDelete) wait() bool {
	return true && !me.NoWait
}

func (me *ExchangeDelete) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Exchange); err != nil {
		return
	}

	return
}

func (me *ExchangeDelete) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Exchange, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  This method confirms the deletion of an exchange.  */
type ExchangeDeleteOk struct {
}

func (me *ExchangeDeleteOk) id() (uint16, uint16) {
	return 40, 21
}

func (me *ExchangeDeleteOk) wait() bool {
	return true
}

func (me *ExchangeDeleteOk) write(w io.Writer) (err error) {

	return
}

func (me *ExchangeDeleteOk) read(r io.Reader) (err error) {

	return
}

/*  This method binds an exchange to an exchange.  */
type ExchangeBind struct {
	reserved1   uint16
	Destination string // name of the destination exchange to bind to
	Source      string // name of the source exchange to bind to
	RoutingKey  string // message routing key
	NoWait      bool
	Arguments   Table // arguments for binding

}

func (me *ExchangeBind) id() (uint16, uint16) {
	return 40, 30
}

func (me *ExchangeBind) wait() bool {
	return true && !me.NoWait
}

func (me *ExchangeBind) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, me.Source); err != nil {
		return
	}
	if err = writeShortstr(w, me.RoutingKey); err != nil {
		return
	}

	var bits byte

	if me.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (me *ExchangeBind) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Destination, err = readShortstr(r); err != nil {
		return
	}
	if me.Source, err = readShortstr(r); err != nil {
		return
	}
	if me.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.NoWait = (bits&(1<<0) > 0)

	return
}

/*  This method confirms that the bind was successful.  */
type ExchangeBindOk struct {
}

func (me *ExchangeBindOk) id() (uint16, uint16) {
	return 40, 31
}

func (me *ExchangeBindOk) wait() bool {
	return true
}

func (me *ExchangeBindOk) write(w io.Writer) (err error) {

	return
}

func (me *ExchangeBindOk) read(r io.Reader) (err error) {

	return
}

/*  This method unbinds an exchange from an exchange.  */
type ExchangeUnbind struct {
	reserved1   uint16
	Destination string
	Source      string
	RoutingKey  string // routing key of binding
	NoWait      bool
	Arguments   Table // arguments of binding

}

func (me *ExchangeUnbind) id() (uint16, uint16) {
	return 40, 40
}

func (me *ExchangeUnbind) wait() bool {
	return true && !me.NoWait
}

func (me *ExchangeUnbind) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, me.Source); err != nil {
		return
	}
	if err = writeShortstr(w, me.RoutingKey); err != nil {
		return
	}

	var bits byte

	if me.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (me *ExchangeUnbind) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Destination, err = readShortstr(r); err != nil {
		return
	}
	if me.Source, err = readShortstr(r); err != nil {
		return
	}
	if me.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.NoWait = (bits&(1<<0) > 0)

	return
}

/*  This method confirms that the unbind was successful.  */
type ExchangeUnbindOk struct {
}

func (me *ExchangeUnbindOk) id() (uint16, uint16) {
	return 40, 51
}

func (me *ExchangeUnbindOk) wait() bool {
	return true
}

func (me *ExchangeUnbindOk) write(w io.Writer) (err error) {

	return
}

func (me *ExchangeUnbindOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method creates or checks a queue. When creating a new queue the client can
   specify various properties that control the durability of the queue and its
   contents, and the level of sharing for the queue.
*/
type QueueDeclare struct {
	reserved1  uint16
	Queue      string
	Passive    bool // do not create queue
	Durable    bool // request a durable queue
	Exclusive  bool // request an exclusive queue
	AutoDelete bool // auto-delete queue when unused
	NoWait     bool
	Arguments  Table // arguments for declaration

}

func (me *QueueDeclare) id() (uint16, uint16) {
	return 50, 10
}

func (me *QueueDeclare) wait() bool {
	return true && !me.NoWait
}

func (me *QueueDeclare) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}

	var bits byte

	if me.Passive {
		bits |= 1 << 0
	}

	if me.Durable {
		bits |= 1 << 1
	}

	if me.Exclusive {
		bits |= 1 << 2
	}

	if me.AutoDelete {
		bits |= 1 << 3
	}

	if me.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (me *QueueDeclare) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.Passive = (bits&(1<<0) > 0)
	me.Durable = (bits&(1<<1) > 0)
	me.Exclusive = (bits&(1<<2) > 0)
	me.AutoDelete = (bits&(1<<3) > 0)
	me.NoWait = (bits&(1<<4) > 0)

	return
}

/*  
   This method confirms a Declare method and confirms the name of the queue, essential
   for automatically-named queues.
*/
type QueueDeclareOk struct {
	Queue         string
	MessageCount  uint32
	ConsumerCount uint32 // number of consumers

}

func (me *QueueDeclareOk) id() (uint16, uint16) {
	return 50, 11
}

func (me *QueueDeclareOk) wait() bool {
	return true
}

func (me *QueueDeclareOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}

	return
}

func (me *QueueDeclareOk) read(r io.Reader) (err error) {

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method binds a queue to an exchange. Until a queue is bound it will not
   receive any messages. In a classic messaging model, store-and-forward queues
   are bound to a direct exchange and subscription queues are bound to a topic
   exchange.
*/
type QueueBind struct {
	reserved1  uint16
	Queue      string
	Exchange   string // name of the exchange to bind to
	RoutingKey string // message routing key
	NoWait     bool
	Arguments  Table // arguments for binding

}

func (me *QueueBind) id() (uint16, uint16) {
	return 50, 20
}

func (me *QueueBind) wait() bool {
	return true && !me.NoWait
}

func (me *QueueBind) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, me.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, me.RoutingKey); err != nil {
		return
	}

	var bits byte

	if me.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (me *QueueBind) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}
	if me.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if me.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.NoWait = (bits&(1<<0) > 0)

	return
}

/*  This method confirms that the bind was successful.  */
type QueueBindOk struct {
}

func (me *QueueBindOk) id() (uint16, uint16) {
	return 50, 21
}

func (me *QueueBindOk) wait() bool {
	return true
}

func (me *QueueBindOk) write(w io.Writer) (err error) {

	return
}

func (me *QueueBindOk) read(r io.Reader) (err error) {

	return
}

/*  This method unbinds a queue from an exchange.  */
type QueueUnbind struct {
	reserved1  uint16
	Queue      string
	Exchange   string
	RoutingKey string // routing key of binding
	Arguments  Table  // arguments of binding

}

func (me *QueueUnbind) id() (uint16, uint16) {
	return 50, 50
}

func (me *QueueUnbind) wait() bool {
	return true
}

func (me *QueueUnbind) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, me.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, me.RoutingKey); err != nil {
		return
	}

	return
}

func (me *QueueUnbind) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}
	if me.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if me.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  This method confirms that the unbind was successful.  */
type QueueUnbindOk struct {
}

func (me *QueueUnbindOk) id() (uint16, uint16) {
	return 50, 51
}

func (me *QueueUnbindOk) wait() bool {
	return true
}

func (me *QueueUnbindOk) write(w io.Writer) (err error) {

	return
}

func (me *QueueUnbindOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method removes all messages from a queue which are not awaiting
   acknowledgment.
*/
type QueuePurge struct {
	reserved1 uint16
	Queue     string
	NoWait    bool
}

func (me *QueuePurge) id() (uint16, uint16) {
	return 50, 30
}

func (me *QueuePurge) wait() bool {
	return true && !me.NoWait
}

func (me *QueuePurge) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}

	return
}

func (me *QueuePurge) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  This method confirms the purge of a queue.  */
type QueuePurgeOk struct {
	MessageCount uint32
}

func (me *QueuePurgeOk) id() (uint16, uint16) {
	return 50, 31
}

func (me *QueuePurgeOk) wait() bool {
	return true
}

func (me *QueuePurgeOk) write(w io.Writer) (err error) {

	return
}

func (me *QueuePurgeOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method deletes a queue. When a queue is deleted any pending messages are sent
   to a dead-letter queue if this is defined in the server configuration, and all
   consumers on the queue are cancelled.
*/
type QueueDelete struct {
	reserved1 uint16
	Queue     string
	IfUnused  bool // delete only if unused
	IfEmpty   bool // delete only if empty
	NoWait    bool
}

func (me *QueueDelete) id() (uint16, uint16) {
	return 50, 40
}

func (me *QueueDelete) wait() bool {
	return true && !me.NoWait
}

func (me *QueueDelete) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}

	return
}

func (me *QueueDelete) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  This method confirms the deletion of a queue.  */
type QueueDeleteOk struct {
	MessageCount uint32
}

func (me *QueueDeleteOk) id() (uint16, uint16) {
	return 50, 41
}

func (me *QueueDeleteOk) wait() bool {
	return true
}

func (me *QueueDeleteOk) write(w io.Writer) (err error) {

	return
}

func (me *QueueDeleteOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method requests a specific quality of service. The QoS can be specified for the
   current channel or for all channels on the connection. The particular properties and
   semantics of a qos method always depend on the content class semantics. Though the
   qos method could in principle apply to both peers, it is currently meaningful only
   for the server.
*/
type BasicQos struct {
	PrefetchSize  uint32 // prefetch window in octets
	PrefetchCount uint16 // prefetch window in messages
	Global        bool   // apply to entire connection

}

func (me *BasicQos) id() (uint16, uint16) {
	return 60, 10
}

func (me *BasicQos) wait() bool {
	return true
}

func (me *BasicQos) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.PrefetchSize); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, me.PrefetchCount); err != nil {
		return
	}

	return
}

func (me *BasicQos) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.PrefetchSize); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &me.PrefetchCount); err != nil {
		return
	}

	return
}

/*  
   This method tells the client that the requested QoS levels could be handled by the
   server. The requested QoS applies to all active consumers until a new QoS is
   defined.
*/
type BasicQosOk struct {
}

func (me *BasicQosOk) id() (uint16, uint16) {
	return 60, 11
}

func (me *BasicQosOk) wait() bool {
	return true
}

func (me *BasicQosOk) write(w io.Writer) (err error) {

	return
}

func (me *BasicQosOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method asks the server to start a "consumer", which is a transient request for
   messages from a specific queue. Consumers last as long as the channel they were
   declared on, or until the client cancels them.
*/
type BasicConsume struct {
	reserved1   uint16
	Queue       string
	ConsumerTag string
	NoLocal     bool
	NoAck       bool
	Exclusive   bool // request exclusive access
	NoWait      bool
	Arguments   Table // arguments for declaration

}

func (me *BasicConsume) id() (uint16, uint16) {
	return 60, 20
}

func (me *BasicConsume) wait() bool {
	return true && !me.NoWait
}

func (me *BasicConsume) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, me.ConsumerTag); err != nil {
		return
	}

	var bits byte

	if me.NoLocal {
		bits |= 1 << 0
	}

	if me.NoAck {
		bits |= 1 << 1
	}

	if me.Exclusive {
		bits |= 1 << 2
	}

	if me.NoWait {
		bits |= 1 << 3
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (me *BasicConsume) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}
	if me.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.NoLocal = (bits&(1<<0) > 0)
	me.NoAck = (bits&(1<<1) > 0)
	me.Exclusive = (bits&(1<<2) > 0)
	me.NoWait = (bits&(1<<3) > 0)

	return
}

/*  
   The server provides the client with a consumer tag, which is used by the client
   for methods called on the consumer at a later stage.
*/
type BasicConsumeOk struct {
	ConsumerTag string
}

func (me *BasicConsumeOk) id() (uint16, uint16) {
	return 60, 21
}

func (me *BasicConsumeOk) wait() bool {
	return true
}

func (me *BasicConsumeOk) write(w io.Writer) (err error) {

	return
}

func (me *BasicConsumeOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method cancels a consumer. This does not affect already delivered
   messages, but it does mean the server will not send any more messages for
   that consumer. The client may receive an arbitrary number of messages in
   between sending the cancel method and receiving the cancel-ok reply.

   It may also be sent from the server to the client in the event
   of the consumer being unexpectedly cancelled (i.e. cancelled
   for any reason other than the server receiving the
   corresponding basic.cancel from the client). This allows
   clients to be notified of the loss of consumers due to events
   such as queue deletion. Note that as it is not a MUST for
   clients to accept this method from the client, it is advisable
   for the broker to be able to identify those clients that are
   capable of accepting the method, through some means of
   capability negotiation.
*/
type BasicCancel struct {
	ConsumerTag string
	NoWait      bool
}

func (me *BasicCancel) id() (uint16, uint16) {
	return 60, 30
}

func (me *BasicCancel) wait() bool {
	return true && !me.NoWait
}

func (me *BasicCancel) write(w io.Writer) (err error) {

	if err = writeShortstr(w, me.ConsumerTag); err != nil {
		return
	}

	return
}

func (me *BasicCancel) read(r io.Reader) (err error) {

	if me.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method confirms that the cancellation was completed.
*/
type BasicCancelOk struct {
	ConsumerTag string
}

func (me *BasicCancelOk) id() (uint16, uint16) {
	return 60, 31
}

func (me *BasicCancelOk) wait() bool {
	return true
}

func (me *BasicCancelOk) write(w io.Writer) (err error) {

	return
}

func (me *BasicCancelOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method publishes a message to a specific exchange. The message will be routed
   to queues as defined by the exchange configuration and distributed to any active
   consumers when the transaction, if any, is committed.
*/
type BasicPublish struct {
	reserved1  uint16
	Exchange   string
	RoutingKey string // Message routing key
	Mandatory  bool   // indicate mandatory routing
	Immediate  bool   // request immediate delivery
	Properties Properties
	Body       []byte
}

func (me *BasicPublish) id() (uint16, uint16) {
	return 60, 40
}

func (me *BasicPublish) wait() bool {
	return false
}

func (me *BasicPublish) GetContent() (Properties, []byte) {
	return me.Properties, me.Body
}

func (me *BasicPublish) SetContent(properties Properties, body []byte) {
	me.Properties, me.Body = properties, body
}

func (me *BasicPublish) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, me.RoutingKey); err != nil {
		return
	}

	return
}

func (me *BasicPublish) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if me.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method returns an undeliverable message that was published with the "immediate"
   flag set, or an unroutable message published with the "mandatory" flag set. The
   reply code and text provide information about the reason that the message was
   undeliverable.
*/
type BasicReturn struct {
	ReplyCode  uint16
	ReplyText  string
	Exchange   string
	RoutingKey string // Message routing key
	Properties Properties
	Body       []byte
}

func (me *BasicReturn) id() (uint16, uint16) {
	return 60, 50
}

func (me *BasicReturn) wait() bool {
	return false
}

func (me *BasicReturn) GetContent() (Properties, []byte) {
	return me.Properties, me.Body
}

func (me *BasicReturn) SetContent(properties Properties, body []byte) {
	me.Properties, me.Body = properties, body
}

func (me *BasicReturn) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.ReplyCode); err != nil {
		return
	}

	return
}

func (me *BasicReturn) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.ReplyCode); err != nil {
		return
	}

	return
}

/*  
   This method delivers a message to the client, via a consumer. In the asynchronous
   message delivery model, the client starts a consumer using the Consume method, then
   the server responds with Deliver methods as and when messages arrive for that
   consumer.
*/
type BasicDeliver struct {
	ConsumerTag string
	DeliveryTag uint64
	Redelivered bool
	Exchange    string
	RoutingKey  string // Message routing key
	Properties  Properties
	Body        []byte
}

func (me *BasicDeliver) id() (uint16, uint16) {
	return 60, 60
}

func (me *BasicDeliver) wait() bool {
	return false
}

func (me *BasicDeliver) GetContent() (Properties, []byte) {
	return me.Properties, me.Body
}

func (me *BasicDeliver) SetContent(properties Properties, body []byte) {
	me.Properties, me.Body = properties, body
}

func (me *BasicDeliver) write(w io.Writer) (err error) {

	if err = writeShortstr(w, me.ConsumerTag); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, me.DeliveryTag); err != nil {
		return
	}

	var bits byte

	if me.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (me *BasicDeliver) read(r io.Reader) (err error) {

	if me.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &me.DeliveryTag); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.Redelivered = (bits&(1<<0) > 0)

	return
}

/*  
   This method provides a direct access to the messages in a queue using a synchronous
   dialogue that is designed for specific types of application where synchronous
   functionality is more important than performance.
*/
type BasicGet struct {
	reserved1 uint16
	Queue     string
	NoAck     bool
}

func (me *BasicGet) id() (uint16, uint16) {
	return 60, 70
}

func (me *BasicGet) wait() bool {
	return true
}

func (me *BasicGet) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, me.Queue); err != nil {
		return
	}

	return
}

func (me *BasicGet) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.reserved1); err != nil {
		return
	}

	if me.Queue, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method delivers a message to the client following a get method. A message
   delivered by 'get-ok' must be acknowledged unless the no-ack option was set in the
   get method.
*/
type BasicGetOk struct {
	DeliveryTag  uint64
	Redelivered  bool
	Exchange     string
	RoutingKey   string // Message routing key
	MessageCount uint32
	Properties   Properties
	Body         []byte
}

func (me *BasicGetOk) id() (uint16, uint16) {
	return 60, 71
}

func (me *BasicGetOk) wait() bool {
	return true
}

func (me *BasicGetOk) GetContent() (Properties, []byte) {
	return me.Properties, me.Body
}

func (me *BasicGetOk) SetContent(properties Properties, body []byte) {
	me.Properties, me.Body = properties, body
}

func (me *BasicGetOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.DeliveryTag); err != nil {
		return
	}

	var bits byte

	if me.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeShortstr(w, me.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, me.RoutingKey); err != nil {
		return
	}

	return
}

func (me *BasicGetOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.DeliveryTag); err != nil {
		return
	}

	var bits byte
	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	me.Redelivered = (bits&(1<<0) > 0)

	if me.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if me.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

/*  
   This method tells the client that the queue has no messages available for the
   client.
*/
type BasicGetEmpty struct {
	reserved1 string
}

func (me *BasicGetEmpty) id() (uint16, uint16) {
	return 60, 72
}

func (me *BasicGetEmpty) wait() bool {
	return true
}

func (me *BasicGetEmpty) write(w io.Writer) (err error) {

	return
}

func (me *BasicGetEmpty) read(r io.Reader) (err error) {

	return
}

/*  
   When sent by the client, this method acknowledges one or more
   messages delivered via the Deliver or Get-Ok methods.

   When sent by server, this method acknowledges one or more
   messages published with the Publish method on a channel in
   confirm mode.

   The acknowledgement can be for a single message or a set of
   messages up to and including a specific message.
*/
type BasicAck struct {
	DeliveryTag uint64
	Multiple    bool // acknowledge multiple messages

}

func (me *BasicAck) id() (uint16, uint16) {
	return 60, 80
}

func (me *BasicAck) wait() bool {
	return false
}

func (me *BasicAck) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.DeliveryTag); err != nil {
		return
	}

	return
}

func (me *BasicAck) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.DeliveryTag); err != nil {
		return
	}

	return
}

/*  
   This method allows a client to reject a message. It can be used to interrupt and
   cancel large incoming messages, or return untreatable messages to their original
   queue.
*/
type BasicReject struct {
	DeliveryTag uint64
	Requeue     bool // requeue the message

}

func (me *BasicReject) id() (uint16, uint16) {
	return 60, 90
}

func (me *BasicReject) wait() bool {
	return false
}

func (me *BasicReject) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.DeliveryTag); err != nil {
		return
	}

	return
}

func (me *BasicReject) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.DeliveryTag); err != nil {
		return
	}

	return
}

/*  
   This method asks the server to redeliver all unacknowledged messages on a
   specified channel. Zero or more messages may be redelivered.  This method
   is deprecated in favour of the synchronous Recover/Recover-Ok.
*/
type BasicRecoverAsync struct {
	Requeue bool // requeue the message

}

func (me *BasicRecoverAsync) id() (uint16, uint16) {
	return 60, 100
}

func (me *BasicRecoverAsync) wait() bool {
	return false
}

func (me *BasicRecoverAsync) write(w io.Writer) (err error) {

	return
}

func (me *BasicRecoverAsync) read(r io.Reader) (err error) {

	return
}

/*  
   This method asks the server to redeliver all unacknowledged messages on a
   specified channel. Zero or more messages may be redelivered.  This method
   replaces the asynchronous Recover.
*/
type BasicRecover struct {
	Requeue bool // requeue the message

}

func (me *BasicRecover) id() (uint16, uint16) {
	return 60, 110
}

func (me *BasicRecover) wait() bool {
	return false
}

func (me *BasicRecover) write(w io.Writer) (err error) {

	return
}

func (me *BasicRecover) read(r io.Reader) (err error) {

	return
}

/*  
   This method acknowledges a Basic.Recover method.
*/
type BasicRecoverOk struct {
}

func (me *BasicRecoverOk) id() (uint16, uint16) {
	return 60, 111
}

func (me *BasicRecoverOk) wait() bool {
	return true
}

func (me *BasicRecoverOk) write(w io.Writer) (err error) {

	return
}

func (me *BasicRecoverOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method allows a client to reject one or more incoming messages. It can be
   used to interrupt and cancel large incoming messages, or return untreatable
   messages to their original queue.

   This method is also used by the server to inform publishers on channels in
   confirm mode of unhandled messages.  If a publisher receives this method, it
   probably needs to republish the offending messages.
*/
type BasicNack struct {
	DeliveryTag uint64
	Multiple    bool // reject multiple messages
	Requeue     bool // requeue the message

}

func (me *BasicNack) id() (uint16, uint16) {
	return 60, 120
}

func (me *BasicNack) wait() bool {
	return false
}

func (me *BasicNack) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, me.DeliveryTag); err != nil {
		return
	}

	return
}

func (me *BasicNack) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &me.DeliveryTag); err != nil {
		return
	}

	return
}

/*  
   This method sets the channel to use standard transactions. The client must use this
   method at least once on a channel before using the Commit or Rollback methods.
*/
type TxSelect struct {
}

func (me *TxSelect) id() (uint16, uint16) {
	return 90, 10
}

func (me *TxSelect) wait() bool {
	return true
}

func (me *TxSelect) write(w io.Writer) (err error) {

	return
}

func (me *TxSelect) read(r io.Reader) (err error) {

	return
}

/*  
   This method confirms to the client that the channel was successfully set to use
   standard transactions.
*/
type TxSelectOk struct {
}

func (me *TxSelectOk) id() (uint16, uint16) {
	return 90, 11
}

func (me *TxSelectOk) wait() bool {
	return true
}

func (me *TxSelectOk) write(w io.Writer) (err error) {

	return
}

func (me *TxSelectOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method commits all message publications and acknowledgments performed in
   the current transaction.  A new transaction starts immediately after a commit.
*/
type TxCommit struct {
}

func (me *TxCommit) id() (uint16, uint16) {
	return 90, 20
}

func (me *TxCommit) wait() bool {
	return true
}

func (me *TxCommit) write(w io.Writer) (err error) {

	return
}

func (me *TxCommit) read(r io.Reader) (err error) {

	return
}

/*  
   This method confirms to the client that the commit succeeded. Note that if a commit
   fails, the server raises a channel exception.
*/
type TxCommitOk struct {
}

func (me *TxCommitOk) id() (uint16, uint16) {
	return 90, 21
}

func (me *TxCommitOk) wait() bool {
	return true
}

func (me *TxCommitOk) write(w io.Writer) (err error) {

	return
}

func (me *TxCommitOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method abandons all message publications and acknowledgments performed in
   the current transaction. A new transaction starts immediately after a rollback.
   Note that unacked messages will not be automatically redelivered by rollback;
   if that is required an explicit recover call should be issued.
*/
type TxRollback struct {
}

func (me *TxRollback) id() (uint16, uint16) {
	return 90, 30
}

func (me *TxRollback) wait() bool {
	return true
}

func (me *TxRollback) write(w io.Writer) (err error) {

	return
}

func (me *TxRollback) read(r io.Reader) (err error) {

	return
}

/*  
   This method confirms to the client that the rollback succeeded. Note that if an
   rollback fails, the server raises a channel exception.
*/
type TxRollbackOk struct {
}

func (me *TxRollbackOk) id() (uint16, uint16) {
	return 90, 31
}

func (me *TxRollbackOk) wait() bool {
	return true
}

func (me *TxRollbackOk) write(w io.Writer) (err error) {

	return
}

func (me *TxRollbackOk) read(r io.Reader) (err error) {

	return
}

/*  
   This method sets the channel to use publisher acknowledgements.
   The client can only use this method on a non-transactional
   channel.
*/
type ConfirmSelect struct {
	Nowait bool
}

func (me *ConfirmSelect) id() (uint16, uint16) {
	return 85, 10
}

func (me *ConfirmSelect) wait() bool {
	return true
}

func (me *ConfirmSelect) write(w io.Writer) (err error) {

	return
}

func (me *ConfirmSelect) read(r io.Reader) (err error) {

	return
}

/*  
   This method confirms to the client that the channel was successfully
   set to use publisher acknowledgements.
*/
type ConfirmSelectOk struct {
}

func (me *ConfirmSelectOk) id() (uint16, uint16) {
	return 85, 11
}

func (me *ConfirmSelectOk) wait() bool {
	return true
}

func (me *ConfirmSelectOk) write(w io.Writer) (err error) {

	return
}

func (me *ConfirmSelectOk) read(r io.Reader) (err error) {

	return
}

func (me *Framer) parseMethodFrame(channel uint16, size uint32) (frame Frame, err error) {
	mf := &MethodFrame{
		ChannelId: channel,
	}

	if err = binary.Read(me.r, binary.BigEndian, &mf.ClassId); err != nil {
		return
	}

	if err = binary.Read(me.r, binary.BigEndian, &mf.MethodId); err != nil {
		return
	}

	switch mf.ClassId {

	case 10: // connection
		switch mf.MethodId {

		case 10: // connection start
			//fmt.Println("NextMethod: class:10 method:10")
			method := &ConnectionStart{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // connection start-ok
			//fmt.Println("NextMethod: class:10 method:11")
			method := &ConnectionStartOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // connection secure
			//fmt.Println("NextMethod: class:10 method:20")
			method := &ConnectionSecure{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // connection secure-ok
			//fmt.Println("NextMethod: class:10 method:21")
			method := &ConnectionSecureOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // connection tune
			//fmt.Println("NextMethod: class:10 method:30")
			method := &ConnectionTune{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // connection tune-ok
			//fmt.Println("NextMethod: class:10 method:31")
			method := &ConnectionTuneOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // connection open
			//fmt.Println("NextMethod: class:10 method:40")
			method := &ConnectionOpen{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // connection open-ok
			//fmt.Println("NextMethod: class:10 method:41")
			method := &ConnectionOpenOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // connection close
			//fmt.Println("NextMethod: class:10 method:50")
			method := &ConnectionClose{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // connection close-ok
			//fmt.Println("NextMethod: class:10 method:51")
			method := &ConnectionCloseOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 20: // channel
		switch mf.MethodId {

		case 10: // channel open
			//fmt.Println("NextMethod: class:20 method:10")
			method := &ChannelOpen{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // channel open-ok
			//fmt.Println("NextMethod: class:20 method:11")
			method := &ChannelOpenOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // channel flow
			//fmt.Println("NextMethod: class:20 method:20")
			method := &ChannelFlow{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // channel flow-ok
			//fmt.Println("NextMethod: class:20 method:21")
			method := &ChannelFlowOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // channel close
			//fmt.Println("NextMethod: class:20 method:40")
			method := &ChannelClose{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // channel close-ok
			//fmt.Println("NextMethod: class:20 method:41")
			method := &ChannelCloseOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 40: // exchange
		switch mf.MethodId {

		case 10: // exchange declare
			//fmt.Println("NextMethod: class:40 method:10")
			method := &ExchangeDeclare{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // exchange declare-ok
			//fmt.Println("NextMethod: class:40 method:11")
			method := &ExchangeDeclareOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // exchange delete
			//fmt.Println("NextMethod: class:40 method:20")
			method := &ExchangeDelete{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // exchange delete-ok
			//fmt.Println("NextMethod: class:40 method:21")
			method := &ExchangeDeleteOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // exchange bind
			//fmt.Println("NextMethod: class:40 method:30")
			method := &ExchangeBind{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // exchange bind-ok
			//fmt.Println("NextMethod: class:40 method:31")
			method := &ExchangeBindOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // exchange unbind
			//fmt.Println("NextMethod: class:40 method:40")
			method := &ExchangeUnbind{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // exchange unbind-ok
			//fmt.Println("NextMethod: class:40 method:51")
			method := &ExchangeUnbindOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 50: // queue
		switch mf.MethodId {

		case 10: // queue declare
			//fmt.Println("NextMethod: class:50 method:10")
			method := &QueueDeclare{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // queue declare-ok
			//fmt.Println("NextMethod: class:50 method:11")
			method := &QueueDeclareOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // queue bind
			//fmt.Println("NextMethod: class:50 method:20")
			method := &QueueBind{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // queue bind-ok
			//fmt.Println("NextMethod: class:50 method:21")
			method := &QueueBindOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // queue unbind
			//fmt.Println("NextMethod: class:50 method:50")
			method := &QueueUnbind{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // queue unbind-ok
			//fmt.Println("NextMethod: class:50 method:51")
			method := &QueueUnbindOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // queue purge
			//fmt.Println("NextMethod: class:50 method:30")
			method := &QueuePurge{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // queue purge-ok
			//fmt.Println("NextMethod: class:50 method:31")
			method := &QueuePurgeOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // queue delete
			//fmt.Println("NextMethod: class:50 method:40")
			method := &QueueDelete{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // queue delete-ok
			//fmt.Println("NextMethod: class:50 method:41")
			method := &QueueDeleteOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 60: // basic
		switch mf.MethodId {

		case 10: // basic qos
			//fmt.Println("NextMethod: class:60 method:10")
			method := &BasicQos{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // basic qos-ok
			//fmt.Println("NextMethod: class:60 method:11")
			method := &BasicQosOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // basic consume
			//fmt.Println("NextMethod: class:60 method:20")
			method := &BasicConsume{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // basic consume-ok
			//fmt.Println("NextMethod: class:60 method:21")
			method := &BasicConsumeOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // basic cancel
			//fmt.Println("NextMethod: class:60 method:30")
			method := &BasicCancel{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // basic cancel-ok
			//fmt.Println("NextMethod: class:60 method:31")
			method := &BasicCancelOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // basic publish
			//fmt.Println("NextMethod: class:60 method:40")
			method := &BasicPublish{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // basic return
			//fmt.Println("NextMethod: class:60 method:50")
			method := &BasicReturn{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 60: // basic deliver
			//fmt.Println("NextMethod: class:60 method:60")
			method := &BasicDeliver{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 70: // basic get
			//fmt.Println("NextMethod: class:60 method:70")
			method := &BasicGet{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 71: // basic get-ok
			//fmt.Println("NextMethod: class:60 method:71")
			method := &BasicGetOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 72: // basic get-empty
			//fmt.Println("NextMethod: class:60 method:72")
			method := &BasicGetEmpty{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 80: // basic ack
			//fmt.Println("NextMethod: class:60 method:80")
			method := &BasicAck{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 90: // basic reject
			//fmt.Println("NextMethod: class:60 method:90")
			method := &BasicReject{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 100: // basic recover-async
			//fmt.Println("NextMethod: class:60 method:100")
			method := &BasicRecoverAsync{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 110: // basic recover
			//fmt.Println("NextMethod: class:60 method:110")
			method := &BasicRecover{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 111: // basic recover-ok
			//fmt.Println("NextMethod: class:60 method:111")
			method := &BasicRecoverOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 120: // basic nack
			//fmt.Println("NextMethod: class:60 method:120")
			method := &BasicNack{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 90: // tx
		switch mf.MethodId {

		case 10: // tx select
			//fmt.Println("NextMethod: class:90 method:10")
			method := &TxSelect{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // tx select-ok
			//fmt.Println("NextMethod: class:90 method:11")
			method := &TxSelectOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // tx commit
			//fmt.Println("NextMethod: class:90 method:20")
			method := &TxCommit{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // tx commit-ok
			//fmt.Println("NextMethod: class:90 method:21")
			method := &TxCommitOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // tx rollback
			//fmt.Println("NextMethod: class:90 method:30")
			method := &TxRollback{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // tx rollback-ok
			//fmt.Println("NextMethod: class:90 method:31")
			method := &TxRollbackOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 85: // confirm
		switch mf.MethodId {

		case 10: // confirm select
			//fmt.Println("NextMethod: class:85 method:10")
			method := &ConfirmSelect{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // confirm select-ok
			//fmt.Println("NextMethod: class:85 method:11")
			method := &ConfirmSelectOk{}
			if err = method.read(me.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	default:
		return nil, fmt.Errorf("Bad method frame, unknown class %d", mf.ClassId)
	}

	return mf, nil
}
