/* GENERATED FILE - DO NOT EDIT */
/* Rebuild from the protocol/gen.go tool */

package wire

import (
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

func (me ConnectionStart) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(10)
	buf.PutOctet(me.VersionMajor)
	buf.PutOctet(me.VersionMinor)
	buf.PutTable(me.ServerProperties)
	buf.PutLongstr(me.Mechanisms)
	buf.PutLongstr(me.Locales)

	fmt.Println("encode: connection.start 10 10", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me ConnectionStartOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(11)
	buf.PutTable(me.ClientProperties)
	buf.PutShortstr(me.Mechanism)
	buf.PutLongstr(me.Response)
	buf.PutShortstr(me.Locale)

	fmt.Println("encode: connection.start-ok 10 11", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   The SASL protocol works by exchanging challenges and responses until both peers have
   received sufficient information to authenticate each other. This method challenges
   the client to provide more information.
*/
type ConnectionSecure struct {
	Challenge string // security challenge data
}

func (me ConnectionSecure) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(20)
	buf.PutLongstr(me.Challenge)

	fmt.Println("encode: connection.secure 10 20", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method attempts to authenticate, passing a block of SASL data for the security
   mechanism at the server side.
*/
type ConnectionSecureOk struct {
	Response string // security response data
}

func (me ConnectionSecureOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(21)
	buf.PutLongstr(me.Response)

	fmt.Println("encode: connection.secure-ok 10 21", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me ConnectionTune) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(30)
	buf.PutShort(me.ChannelMax)
	buf.PutLong(me.FrameMax)
	buf.PutShort(me.Heartbeat)

	fmt.Println("encode: connection.tune 10 30", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me ConnectionTuneOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(31)
	buf.PutShort(me.ChannelMax)
	buf.PutLong(me.FrameMax)
	buf.PutShort(me.Heartbeat)

	fmt.Println("encode: connection.tune-ok 10 31", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method opens a connection to a virtual host, which is a collection of
   resources, and acts to separate multiple application domains within a server.
   The server may apply arbitrary limits per virtual host, such as the number
   of each type of entity that may be used, per connection and/or in total.
*/
type ConnectionOpen struct {
	VirtualHost string // virtual host name    
}

func (me ConnectionOpen) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(40)
	buf.PutShortstr(me.VirtualHost)
	var Reserved1 string
	buf.PutShortstr(Reserved1)
	var Reserved2 bool
	buf.PutOctet(0)
	buf.PutBit(Reserved2, 0)

	fmt.Println("encode: connection.open 10 40", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method signals to the client that the connection is ready for use.
*/
type ConnectionOpenOk struct {
}

func (me ConnectionOpenOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(41)
	var Reserved1 string
	buf.PutShortstr(Reserved1)

	fmt.Println("encode: connection.open-ok 10 41", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me ConnectionClose) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(50)
	buf.PutShort(me.ReplyCode)
	buf.PutShortstr(me.ReplyText)
	buf.PutShort(me.ClassId)
	buf.PutShort(me.MethodId)

	fmt.Println("encode: connection.close 10 50", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms a Connection.Close method and tells the recipient that it is
   safe to release resources for the connection and close the socket.
*/
type ConnectionCloseOk struct {
}

func (me ConnectionCloseOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(10)
	buf.PutShort(51)

	fmt.Println("encode: connection.close-ok 10 51", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method opens a channel to the server.
*/
type ChannelOpen struct {
}

func (me ChannelOpen) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(20)
	buf.PutShort(10)
	var Reserved1 string
	buf.PutShortstr(Reserved1)

	fmt.Println("encode: channel.open 20 10", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method signals to the client that the channel is ready for use.
*/
type ChannelOpenOk struct {
}

func (me ChannelOpenOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(20)
	buf.PutShort(11)
	var Reserved1 string
	buf.PutLongstr(Reserved1)

	fmt.Println("encode: channel.open-ok 20 11", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me ChannelFlow) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(20)
	buf.PutShort(20)
	buf.PutOctet(0)
	buf.PutBit(me.Active, 0)

	fmt.Println("encode: channel.flow 20 20", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   Confirms to the peer that a flow command was received and processed.
*/
type ChannelFlowOk struct {
	Active bool // current flow setting
}

func (me ChannelFlowOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(20)
	buf.PutShort(21)
	buf.PutBit(me.Active, 1)

	fmt.Println("encode: channel.flow-ok 20 21", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me ChannelClose) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(20)
	buf.PutShort(40)
	buf.PutShort(me.ReplyCode)
	buf.PutShortstr(me.ReplyText)
	buf.PutShort(me.ClassId)
	buf.PutShort(me.MethodId)

	fmt.Println("encode: channel.close 20 40", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms a Channel.Close method and tells the recipient that it is safe
   to release resources for the channel.
*/
type ChannelCloseOk struct {
}

func (me ChannelCloseOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(20)
	buf.PutShort(41)

	fmt.Println("encode: channel.close-ok 20 41", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method creates an exchange if it does not already exist, and if the exchange
   exists, verifies that it is of the correct and expected class.
*/
type ExchangeDeclare struct {
	Exchange   string
	Type       string // exchange type 
	Passive    bool   // do not create exchange 
	Durable    bool   // request a durable exchange 
	AutoDelete bool   // auto-delete when unused 
	Internal   bool   // create internal exchange 
	NoWait     bool
	Arguments  Table // arguments for declaration
}

func (me ExchangeDeclare) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(10)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Exchange)
	buf.PutShortstr(me.Type)
	buf.PutOctet(0)
	buf.PutBit(me.Passive, 0)
	buf.PutBit(me.Durable, 1)
	buf.PutBit(me.AutoDelete, 2)
	buf.PutBit(me.Internal, 3)
	buf.PutBit(me.NoWait, 4)
	buf.PutTable(me.Arguments)

	fmt.Println("encode: exchange.declare 40 10", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms a Declare method and confirms the name of the exchange,
   essential for automatically-named exchanges.
*/
type ExchangeDeclareOk struct {
}

func (me ExchangeDeclareOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(11)

	fmt.Println("encode: exchange.declare-ok 40 11", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method deletes an exchange. When an exchange is deleted all queue bindings on
   the exchange are cancelled.
*/
type ExchangeDelete struct {
	Exchange string
	IfUnused bool // delete only if unused 
	NoWait   bool
}

func (me ExchangeDelete) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(20)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Exchange)
	buf.PutOctet(0)
	buf.PutBit(me.IfUnused, 0)
	buf.PutBit(me.NoWait, 1)

	fmt.Println("encode: exchange.delete 40 20", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method confirms the deletion of an exchange.  */
type ExchangeDeleteOk struct {
}

func (me ExchangeDeleteOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(21)

	fmt.Println("encode: exchange.delete-ok 40 21", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method binds an exchange to an exchange.  */
type ExchangeBind struct {
	Destination string // name of the destination exchange to bind to 
	Source      string // name of the source exchange to bind to 
	RoutingKey  string // message routing key 
	NoWait      bool
	Arguments   Table // arguments for binding
}

func (me ExchangeBind) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(30)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Destination)
	buf.PutShortstr(me.Source)
	buf.PutShortstr(me.RoutingKey)
	buf.PutOctet(0)
	buf.PutBit(me.NoWait, 0)
	buf.PutTable(me.Arguments)

	fmt.Println("encode: exchange.bind 40 30", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method confirms that the bind was successful.  */
type ExchangeBindOk struct {
}

func (me ExchangeBindOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(31)

	fmt.Println("encode: exchange.bind-ok 40 31", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method unbinds an exchange from an exchange.  */
type ExchangeUnbind struct {
	Destination string
	Source      string
	RoutingKey  string // routing key of binding 
	NoWait      bool
	Arguments   Table // arguments of binding
}

func (me ExchangeUnbind) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(40)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Destination)
	buf.PutShortstr(me.Source)
	buf.PutShortstr(me.RoutingKey)
	buf.PutOctet(0)
	buf.PutBit(me.NoWait, 0)
	buf.PutTable(me.Arguments)

	fmt.Println("encode: exchange.unbind 40 40", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method confirms that the unbind was successful.  */
type ExchangeUnbindOk struct {
}

func (me ExchangeUnbindOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(40)
	buf.PutShort(51)

	fmt.Println("encode: exchange.unbind-ok 40 51", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method creates or checks a queue. When creating a new queue the client can
   specify various properties that control the durability of the queue and its
   contents, and the level of sharing for the queue.
*/
type QueueDeclare struct {
	Queue      string
	Passive    bool // do not create queue 
	Durable    bool // request a durable queue 
	Exclusive  bool // request an exclusive queue 
	AutoDelete bool // auto-delete queue when unused 
	NoWait     bool
	Arguments  Table // arguments for declaration
}

func (me QueueDeclare) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(10)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Queue)
	buf.PutOctet(0)
	buf.PutBit(me.Passive, 0)
	buf.PutBit(me.Durable, 1)
	buf.PutBit(me.Exclusive, 2)
	buf.PutBit(me.AutoDelete, 3)
	buf.PutBit(me.NoWait, 4)
	buf.PutTable(me.Arguments)

	fmt.Println("encode: queue.declare 50 10", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me QueueDeclareOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(11)
	buf.PutShortstr(me.Queue)
	buf.PutLong(me.MessageCount)
	buf.PutLong(me.ConsumerCount)

	fmt.Println("encode: queue.declare-ok 50 11", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method binds a queue to an exchange. Until a queue is bound it will not
   receive any messages. In a classic messaging model, store-and-forward queues
   are bound to a direct exchange and subscription queues are bound to a topic
   exchange.
*/
type QueueBind struct {
	Queue      string
	Exchange   string // name of the exchange to bind to 
	RoutingKey string // message routing key 
	NoWait     bool
	Arguments  Table // arguments for binding
}

func (me QueueBind) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(20)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Queue)
	buf.PutShortstr(me.Exchange)
	buf.PutShortstr(me.RoutingKey)
	buf.PutOctet(0)
	buf.PutBit(me.NoWait, 0)
	buf.PutTable(me.Arguments)

	fmt.Println("encode: queue.bind 50 20", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method confirms that the bind was successful.  */
type QueueBindOk struct {
}

func (me QueueBindOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(21)

	fmt.Println("encode: queue.bind-ok 50 21", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method unbinds a queue from an exchange.  */
type QueueUnbind struct {
	Queue      string
	Exchange   string
	RoutingKey string // routing key of binding 
	Arguments  Table  // arguments of binding
}

func (me QueueUnbind) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(50)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Queue)
	buf.PutShortstr(me.Exchange)
	buf.PutShortstr(me.RoutingKey)
	buf.PutTable(me.Arguments)

	fmt.Println("encode: queue.unbind 50 50", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method confirms that the unbind was successful.  */
type QueueUnbindOk struct {
}

func (me QueueUnbindOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(51)

	fmt.Println("encode: queue.unbind-ok 50 51", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method removes all messages from a queue which are not awaiting
   acknowledgment.
*/
type QueuePurge struct {
	Queue  string
	NoWait bool
}

func (me QueuePurge) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(30)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Queue)
	buf.PutOctet(0)
	buf.PutBit(me.NoWait, 0)

	fmt.Println("encode: queue.purge 50 30", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method confirms the purge of a queue.  */
type QueuePurgeOk struct {
	MessageCount uint32
}

func (me QueuePurgeOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(31)
	buf.PutLong(me.MessageCount)

	fmt.Println("encode: queue.purge-ok 50 31", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method deletes a queue. When a queue is deleted any pending messages are sent
   to a dead-letter queue if this is defined in the server configuration, and all
   consumers on the queue are cancelled.
*/
type QueueDelete struct {
	Queue    string
	IfUnused bool // delete only if unused 
	IfEmpty  bool // delete only if empty 
	NoWait   bool
}

func (me QueueDelete) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(40)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Queue)
	buf.PutOctet(0)
	buf.PutBit(me.IfUnused, 0)
	buf.PutBit(me.IfEmpty, 1)
	buf.PutBit(me.NoWait, 2)

	fmt.Println("encode: queue.delete 50 40", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  This method confirms the deletion of a queue.  */
type QueueDeleteOk struct {
	MessageCount uint32
}

func (me QueueDeleteOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(50)
	buf.PutShort(41)
	buf.PutLong(me.MessageCount)

	fmt.Println("encode: queue.delete-ok 50 41", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me BasicQos) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(10)
	buf.PutLong(me.PrefetchSize)
	buf.PutShort(me.PrefetchCount)
	buf.PutOctet(0)
	buf.PutBit(me.Global, 0)

	fmt.Println("encode: basic.qos 60 10", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method tells the client that the requested QoS levels could be handled by the
   server. The requested QoS applies to all active consumers until a new QoS is
   defined.
*/
type BasicQosOk struct {
}

func (me BasicQosOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(11)

	fmt.Println("encode: basic.qos-ok 60 11", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method asks the server to start a "consumer", which is a transient request for
   messages from a specific queue. Consumers last as long as the channel they were
   declared on, or until the client cancels them.
*/
type BasicConsume struct {
	Queue       string
	ConsumerTag string
	NoLocal     bool
	NoAck       bool
	Exclusive   bool // request exclusive access 
	NoWait      bool
	Arguments   Table // arguments for declaration
}

func (me BasicConsume) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(20)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Queue)
	buf.PutShortstr(me.ConsumerTag)
	buf.PutOctet(0)
	buf.PutBit(me.NoLocal, 0)
	buf.PutBit(me.NoAck, 1)
	buf.PutBit(me.Exclusive, 2)
	buf.PutBit(me.NoWait, 3)
	buf.PutTable(me.Arguments)

	fmt.Println("encode: basic.consume 60 20", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   The server provides the client with a consumer tag, which is used by the client
   for methods called on the consumer at a later stage.
*/
type BasicConsumeOk struct {
	ConsumerTag string
}

func (me BasicConsumeOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(21)
	buf.PutShortstr(me.ConsumerTag)

	fmt.Println("encode: basic.consume-ok 60 21", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me BasicCancel) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(30)
	buf.PutShortstr(me.ConsumerTag)
	buf.PutOctet(0)
	buf.PutBit(me.NoWait, 0)

	fmt.Println("encode: basic.cancel 60 30", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms that the cancellation was completed.
*/
type BasicCancelOk struct {
	ConsumerTag string
}

func (me BasicCancelOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(31)
	buf.PutShortstr(me.ConsumerTag)

	fmt.Println("encode: basic.cancel-ok 60 31", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method publishes a message to a specific exchange. The message will be routed
   to queues as defined by the exchange configuration and distributed to any active
   consumers when the transaction, if any, is committed.
*/
type BasicPublish struct {
	Exchange   string
	RoutingKey string // Message routing key 
	Mandatory  bool   // indicate mandatory routing 
	Immediate  bool   // request immediate delivery
}

func (me BasicPublish) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(40)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Exchange)
	buf.PutShortstr(me.RoutingKey)
	buf.PutOctet(0)
	buf.PutBit(me.Mandatory, 0)
	buf.PutBit(me.Immediate, 1)

	fmt.Println("encode: basic.publish 60 40", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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
}

func (me BasicReturn) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(50)
	buf.PutShort(me.ReplyCode)
	buf.PutShortstr(me.ReplyText)
	buf.PutShortstr(me.Exchange)
	buf.PutShortstr(me.RoutingKey)

	fmt.Println("encode: basic.return 60 50", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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
}

func (me BasicDeliver) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(60)
	buf.PutShortstr(me.ConsumerTag)
	buf.PutLonglong(me.DeliveryTag)
	buf.PutOctet(0)
	buf.PutBit(me.Redelivered, 0)
	buf.PutShortstr(me.Exchange)
	buf.PutShortstr(me.RoutingKey)

	fmt.Println("encode: basic.deliver 60 60", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method provides a direct access to the messages in a queue using a synchronous
   dialogue that is designed for specific types of application where synchronous
   functionality is more important than performance.
*/
type BasicGet struct {
	Queue string
	NoAck bool
}

func (me BasicGet) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(70)
	var Reserved1 uint16
	buf.PutShort(Reserved1)
	buf.PutShortstr(me.Queue)
	buf.PutOctet(0)
	buf.PutBit(me.NoAck, 0)

	fmt.Println("encode: basic.get 60 70", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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
}

func (me BasicGetOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(71)
	buf.PutLonglong(me.DeliveryTag)
	buf.PutOctet(0)
	buf.PutBit(me.Redelivered, 0)
	buf.PutShortstr(me.Exchange)
	buf.PutShortstr(me.RoutingKey)
	buf.PutLong(me.MessageCount)

	fmt.Println("encode: basic.get-ok 60 71", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method tells the client that the queue has no messages available for the
   client.
*/
type BasicGetEmpty struct {
}

func (me BasicGetEmpty) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(72)
	var Reserved1 string
	buf.PutShortstr(Reserved1)

	fmt.Println("encode: basic.get-empty 60 72", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me BasicAck) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(80)
	buf.PutLonglong(me.DeliveryTag)
	buf.PutOctet(0)
	buf.PutBit(me.Multiple, 0)

	fmt.Println("encode: basic.ack 60 80", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me BasicReject) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(90)
	buf.PutLonglong(me.DeliveryTag)
	buf.PutOctet(0)
	buf.PutBit(me.Requeue, 0)

	fmt.Println("encode: basic.reject 60 90", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method asks the server to redeliver all unacknowledged messages on a
   specified channel. Zero or more messages may be redelivered.  This method
   is deprecated in favour of the synchronous Recover/Recover-Ok.
*/
type BasicRecoverAsync struct {
	Requeue bool // requeue the message
}

func (me BasicRecoverAsync) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(100)
	buf.PutBit(me.Requeue, 1)

	fmt.Println("encode: basic.recover-async 60 100", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method asks the server to redeliver all unacknowledged messages on a
   specified channel. Zero or more messages may be redelivered.  This method
   replaces the asynchronous Recover.
*/
type BasicRecover struct {
	Requeue bool // requeue the message
}

func (me BasicRecover) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(110)
	buf.PutBit(me.Requeue, 2)

	fmt.Println("encode: basic.recover 60 110", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method acknowledges a Basic.Recover method.
*/
type BasicRecoverOk struct {
}

func (me BasicRecoverOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(111)

	fmt.Println("encode: basic.recover-ok 60 111", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
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

func (me BasicNack) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(60)
	buf.PutShort(120)
	buf.PutLonglong(me.DeliveryTag)
	buf.PutOctet(0)
	buf.PutBit(me.Multiple, 0)
	buf.PutBit(me.Requeue, 1)

	fmt.Println("encode: basic.nack 60 120", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method sets the channel to use standard transactions. The client must use this
   method at least once on a channel before using the Commit or Rollback methods.
*/
type TxSelect struct {
}

func (me TxSelect) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(90)
	buf.PutShort(10)

	fmt.Println("encode: tx.select 90 10", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms to the client that the channel was successfully set to use
   standard transactions.
*/
type TxSelectOk struct {
}

func (me TxSelectOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(90)
	buf.PutShort(11)

	fmt.Println("encode: tx.select-ok 90 11", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method commits all message publications and acknowledgments performed in
   the current transaction.  A new transaction starts immediately after a commit.
*/
type TxCommit struct {
}

func (me TxCommit) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(90)
	buf.PutShort(20)

	fmt.Println("encode: tx.commit 90 20", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms to the client that the commit succeeded. Note that if a commit
   fails, the server raises a channel exception.
*/
type TxCommitOk struct {
}

func (me TxCommitOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(90)
	buf.PutShort(21)

	fmt.Println("encode: tx.commit-ok 90 21", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method abandons all message publications and acknowledgments performed in
   the current transaction. A new transaction starts immediately after a rollback.
   Note that unacked messages will not be automatically redelivered by rollback;
   if that is required an explicit recover call should be issued.
*/
type TxRollback struct {
}

func (me TxRollback) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(90)
	buf.PutShort(30)

	fmt.Println("encode: tx.rollback 90 30", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms to the client that the rollback succeeded. Note that if an
   rollback fails, the server raises a channel exception.
*/
type TxRollbackOk struct {
}

func (me TxRollbackOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(90)
	buf.PutShort(31)

	fmt.Println("encode: tx.rollback-ok 90 31", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method sets the channel to use publisher acknowledgements.
   The client can only use this method on a non-transactional
   channel.
*/
type ConfirmSelect struct {
	Nowait bool
}

func (me ConfirmSelect) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(85)
	buf.PutShort(10)
	buf.PutBit(me.Nowait, 2)

	fmt.Println("encode: confirm.select 85 10", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

/*  
   This method confirms to the client that the channel was successfully
   set to use publisher acknowledgements.
*/
type ConfirmSelectOk struct {
}

func (me ConfirmSelectOk) WriteTo(w io.Writer) (int64, error) {
	var buf buffer
	buf.PutShort(85)
	buf.PutShort(11)

	fmt.Println("encode: confirm.select-ok 85 11", buf.Bytes())
	wn, err := w.Write(buf.Bytes())
	return int64(wn), err
}

func (me *buffer) NextMethod() (value Method) {
	class := me.NextShort()
	method := me.NextShort()

	switch class {

	case 10: // connection
		switch method {

		case 10: // connection start
			fmt.Println("NextMethod: class:10 method:10")
			message := ConnectionStart{}
			me.NextOctet()
			message.VersionMajor = me.NextOctet()
			message.VersionMinor = me.NextOctet()
			message.ServerProperties = me.NextTable()
			message.Mechanisms = me.NextLongstr()
			message.Locales = me.NextLongstr()

			return message

		case 11: // connection start-ok
			fmt.Println("NextMethod: class:10 method:11")
			message := ConnectionStartOk{}
			message.ClientProperties = me.NextTable()
			message.Mechanism = me.NextShortstr()
			message.Response = me.NextLongstr()
			message.Locale = me.NextShortstr()

			return message

		case 20: // connection secure
			fmt.Println("NextMethod: class:10 method:20")
			message := ConnectionSecure{}
			message.Challenge = me.NextLongstr()

			return message

		case 21: // connection secure-ok
			fmt.Println("NextMethod: class:10 method:21")
			message := ConnectionSecureOk{}
			message.Response = me.NextLongstr()

			return message

		case 30: // connection tune
			fmt.Println("NextMethod: class:10 method:30")
			message := ConnectionTune{}
			message.ChannelMax = me.NextShort()
			message.FrameMax = me.NextLong()
			message.Heartbeat = me.NextShort()

			return message

		case 31: // connection tune-ok
			fmt.Println("NextMethod: class:10 method:31")
			message := ConnectionTuneOk{}
			message.ChannelMax = me.NextShort()
			message.FrameMax = me.NextLong()
			message.Heartbeat = me.NextShort()

			return message

		case 40: // connection open
			fmt.Println("NextMethod: class:10 method:40")
			message := ConnectionOpen{}
			message.VirtualHost = me.NextShortstr()
			_ = me.NextShortstr()
			_ = me.NextBit(0)
			me.NextOctet()
			return message

		case 41: // connection open-ok
			fmt.Println("NextMethod: class:10 method:41")
			message := ConnectionOpenOk{}
			_ = me.NextShortstr()

			return message

		case 50: // connection close
			fmt.Println("NextMethod: class:10 method:50")
			message := ConnectionClose{}
			message.ReplyCode = me.NextShort()
			message.ReplyText = me.NextShortstr()
			message.ClassId = me.NextShort()
			message.MethodId = me.NextShort()

			return message

		case 51: // connection close-ok
			fmt.Println("NextMethod: class:10 method:51")
			message := ConnectionCloseOk{}

			return message

		}

	case 20: // channel
		switch method {

		case 10: // channel open
			fmt.Println("NextMethod: class:20 method:10")
			message := ChannelOpen{}
			_ = me.NextShortstr()

			return message

		case 11: // channel open-ok
			fmt.Println("NextMethod: class:20 method:11")
			message := ChannelOpenOk{}
			_ = me.NextLongstr()

			return message

		case 20: // channel flow
			fmt.Println("NextMethod: class:20 method:20")
			message := ChannelFlow{}
			message.Active = me.NextBit(0)
			me.NextOctet()
			return message

		case 21: // channel flow-ok
			fmt.Println("NextMethod: class:20 method:21")
			message := ChannelFlowOk{}
			message.Active = me.NextBit(0)
			me.NextOctet()
			return message

		case 40: // channel close
			fmt.Println("NextMethod: class:20 method:40")
			message := ChannelClose{}
			message.ReplyCode = me.NextShort()
			message.ReplyText = me.NextShortstr()
			message.ClassId = me.NextShort()
			message.MethodId = me.NextShort()

			return message

		case 41: // channel close-ok
			fmt.Println("NextMethod: class:20 method:41")
			message := ChannelCloseOk{}

			return message

		}

	case 40: // exchange
		switch method {

		case 10: // exchange declare
			fmt.Println("NextMethod: class:40 method:10")
			message := ExchangeDeclare{}
			_ = me.NextShort()
			message.Exchange = me.NextShortstr()
			message.Type = me.NextShortstr()
			message.Passive = me.NextBit(0)
			message.Durable = me.NextBit(1)
			message.AutoDelete = me.NextBit(2)
			message.Internal = me.NextBit(3)
			message.NoWait = me.NextBit(4)
			me.NextOctet()
			message.Arguments = me.NextTable()

			return message

		case 11: // exchange declare-ok
			fmt.Println("NextMethod: class:40 method:11")
			message := ExchangeDeclareOk{}

			return message

		case 20: // exchange delete
			fmt.Println("NextMethod: class:40 method:20")
			message := ExchangeDelete{}
			_ = me.NextShort()
			message.Exchange = me.NextShortstr()
			message.IfUnused = me.NextBit(0)
			message.NoWait = me.NextBit(1)
			me.NextOctet()
			return message

		case 21: // exchange delete-ok
			fmt.Println("NextMethod: class:40 method:21")
			message := ExchangeDeleteOk{}

			return message

		case 30: // exchange bind
			fmt.Println("NextMethod: class:40 method:30")
			message := ExchangeBind{}
			_ = me.NextShort()
			message.Destination = me.NextShortstr()
			message.Source = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()
			message.NoWait = me.NextBit(0)
			me.NextOctet()
			message.Arguments = me.NextTable()

			return message

		case 31: // exchange bind-ok
			fmt.Println("NextMethod: class:40 method:31")
			message := ExchangeBindOk{}

			return message

		case 40: // exchange unbind
			fmt.Println("NextMethod: class:40 method:40")
			message := ExchangeUnbind{}
			_ = me.NextShort()
			message.Destination = me.NextShortstr()
			message.Source = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()
			message.NoWait = me.NextBit(0)
			me.NextOctet()
			message.Arguments = me.NextTable()

			return message

		case 51: // exchange unbind-ok
			fmt.Println("NextMethod: class:40 method:51")
			message := ExchangeUnbindOk{}

			return message

		}

	case 50: // queue
		switch method {

		case 10: // queue declare
			fmt.Println("NextMethod: class:50 method:10")
			message := QueueDeclare{}
			_ = me.NextShort()
			message.Queue = me.NextShortstr()
			message.Passive = me.NextBit(0)
			message.Durable = me.NextBit(1)
			message.Exclusive = me.NextBit(2)
			message.AutoDelete = me.NextBit(3)
			message.NoWait = me.NextBit(4)
			me.NextOctet()
			message.Arguments = me.NextTable()

			return message

		case 11: // queue declare-ok
			fmt.Println("NextMethod: class:50 method:11")
			message := QueueDeclareOk{}
			message.Queue = me.NextShortstr()
			message.MessageCount = me.NextLong()
			message.ConsumerCount = me.NextLong()

			return message

		case 20: // queue bind
			fmt.Println("NextMethod: class:50 method:20")
			message := QueueBind{}
			_ = me.NextShort()
			message.Queue = me.NextShortstr()
			message.Exchange = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()
			message.NoWait = me.NextBit(0)
			me.NextOctet()
			message.Arguments = me.NextTable()

			return message

		case 21: // queue bind-ok
			fmt.Println("NextMethod: class:50 method:21")
			message := QueueBindOk{}

			return message

		case 50: // queue unbind
			fmt.Println("NextMethod: class:50 method:50")
			message := QueueUnbind{}
			_ = me.NextShort()
			message.Queue = me.NextShortstr()
			message.Exchange = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()
			message.Arguments = me.NextTable()

			return message

		case 51: // queue unbind-ok
			fmt.Println("NextMethod: class:50 method:51")
			message := QueueUnbindOk{}

			return message

		case 30: // queue purge
			fmt.Println("NextMethod: class:50 method:30")
			message := QueuePurge{}
			_ = me.NextShort()
			message.Queue = me.NextShortstr()
			message.NoWait = me.NextBit(0)
			me.NextOctet()
			return message

		case 31: // queue purge-ok
			fmt.Println("NextMethod: class:50 method:31")
			message := QueuePurgeOk{}
			message.MessageCount = me.NextLong()

			return message

		case 40: // queue delete
			fmt.Println("NextMethod: class:50 method:40")
			message := QueueDelete{}
			_ = me.NextShort()
			message.Queue = me.NextShortstr()
			message.IfUnused = me.NextBit(0)
			message.IfEmpty = me.NextBit(1)
			message.NoWait = me.NextBit(2)
			me.NextOctet()
			return message

		case 41: // queue delete-ok
			fmt.Println("NextMethod: class:50 method:41")
			message := QueueDeleteOk{}
			message.MessageCount = me.NextLong()

			return message

		}

	case 60: // basic
		switch method {

		case 10: // basic qos
			fmt.Println("NextMethod: class:60 method:10")
			message := BasicQos{}
			message.PrefetchSize = me.NextLong()
			message.PrefetchCount = me.NextShort()
			message.Global = me.NextBit(0)
			me.NextOctet()
			return message

		case 11: // basic qos-ok
			fmt.Println("NextMethod: class:60 method:11")
			message := BasicQosOk{}

			return message

		case 20: // basic consume
			fmt.Println("NextMethod: class:60 method:20")
			message := BasicConsume{}
			_ = me.NextShort()
			message.Queue = me.NextShortstr()
			message.ConsumerTag = me.NextShortstr()
			message.NoLocal = me.NextBit(0)
			message.NoAck = me.NextBit(1)
			message.Exclusive = me.NextBit(2)
			message.NoWait = me.NextBit(3)
			me.NextOctet()
			message.Arguments = me.NextTable()

			return message

		case 21: // basic consume-ok
			fmt.Println("NextMethod: class:60 method:21")
			message := BasicConsumeOk{}
			message.ConsumerTag = me.NextShortstr()

			return message

		case 30: // basic cancel
			fmt.Println("NextMethod: class:60 method:30")
			message := BasicCancel{}
			message.ConsumerTag = me.NextShortstr()
			message.NoWait = me.NextBit(0)
			me.NextOctet()
			return message

		case 31: // basic cancel-ok
			fmt.Println("NextMethod: class:60 method:31")
			message := BasicCancelOk{}
			message.ConsumerTag = me.NextShortstr()

			return message

		case 40: // basic publish
			fmt.Println("NextMethod: class:60 method:40")
			message := BasicPublish{}
			_ = me.NextShort()
			message.Exchange = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()
			message.Mandatory = me.NextBit(0)
			message.Immediate = me.NextBit(1)
			me.NextOctet()
			return message

		case 50: // basic return
			fmt.Println("NextMethod: class:60 method:50")
			message := BasicReturn{}
			message.ReplyCode = me.NextShort()
			message.ReplyText = me.NextShortstr()
			message.Exchange = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()

			return message

		case 60: // basic deliver
			fmt.Println("NextMethod: class:60 method:60")
			message := BasicDeliver{}
			message.ConsumerTag = me.NextShortstr()
			message.DeliveryTag = me.NextLonglong()
			message.Redelivered = me.NextBit(0)
			me.NextOctet()
			message.Exchange = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()

			return message

		case 70: // basic get
			fmt.Println("NextMethod: class:60 method:70")
			message := BasicGet{}
			_ = me.NextShort()
			message.Queue = me.NextShortstr()
			message.NoAck = me.NextBit(0)
			me.NextOctet()
			return message

		case 71: // basic get-ok
			fmt.Println("NextMethod: class:60 method:71")
			message := BasicGetOk{}
			message.DeliveryTag = me.NextLonglong()
			message.Redelivered = me.NextBit(0)
			me.NextOctet()
			message.Exchange = me.NextShortstr()
			message.RoutingKey = me.NextShortstr()
			message.MessageCount = me.NextLong()

			return message

		case 72: // basic get-empty
			fmt.Println("NextMethod: class:60 method:72")
			message := BasicGetEmpty{}
			_ = me.NextShortstr()

			return message

		case 80: // basic ack
			fmt.Println("NextMethod: class:60 method:80")
			message := BasicAck{}
			message.DeliveryTag = me.NextLonglong()
			message.Multiple = me.NextBit(0)
			me.NextOctet()
			return message

		case 90: // basic reject
			fmt.Println("NextMethod: class:60 method:90")
			message := BasicReject{}
			message.DeliveryTag = me.NextLonglong()
			message.Requeue = me.NextBit(0)
			me.NextOctet()
			return message

		case 100: // basic recover-async
			fmt.Println("NextMethod: class:60 method:100")
			message := BasicRecoverAsync{}
			message.Requeue = me.NextBit(0)
			me.NextOctet()
			return message

		case 110: // basic recover
			fmt.Println("NextMethod: class:60 method:110")
			message := BasicRecover{}
			message.Requeue = me.NextBit(0)
			me.NextOctet()
			return message

		case 111: // basic recover-ok
			fmt.Println("NextMethod: class:60 method:111")
			message := BasicRecoverOk{}

			return message

		case 120: // basic nack
			fmt.Println("NextMethod: class:60 method:120")
			message := BasicNack{}
			message.DeliveryTag = me.NextLonglong()
			message.Multiple = me.NextBit(0)
			message.Requeue = me.NextBit(1)
			me.NextOctet()
			return message

		}

	case 90: // tx
		switch method {

		case 10: // tx select
			fmt.Println("NextMethod: class:90 method:10")
			message := TxSelect{}

			return message

		case 11: // tx select-ok
			fmt.Println("NextMethod: class:90 method:11")
			message := TxSelectOk{}

			return message

		case 20: // tx commit
			fmt.Println("NextMethod: class:90 method:20")
			message := TxCommit{}

			return message

		case 21: // tx commit-ok
			fmt.Println("NextMethod: class:90 method:21")
			message := TxCommitOk{}

			return message

		case 30: // tx rollback
			fmt.Println("NextMethod: class:90 method:30")
			message := TxRollback{}

			return message

		case 31: // tx rollback-ok
			fmt.Println("NextMethod: class:90 method:31")
			message := TxRollbackOk{}

			return message

		}

	case 85: // confirm
		switch method {

		case 10: // confirm select
			fmt.Println("NextMethod: class:85 method:10")
			message := ConfirmSelect{}
			message.Nowait = me.NextBit(0)
			me.NextOctet()
			return message

		case 11: // confirm select-ok
			fmt.Println("NextMethod: class:85 method:11")
			message := ConfirmSelectOk{}

			return message

		}

	}

	return nil
}
