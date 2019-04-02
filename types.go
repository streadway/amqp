// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"fmt"
	"io"
	"time"

	"github.com/streadway/amqp/internal/proto"
)

// Protocol error exception codes
const (
	ContentTooLarge    = proto.ContentTooLarge
	NoRoute            = proto.NoRoute
	NoConsumers        = proto.NoConsumers
	ConnectionForced   = proto.ConnectionForced
	InvalidPath        = proto.InvalidPath
	AccessRefused      = proto.AccessRefused
	NotFound           = proto.NotFound
	ResourceLocked     = proto.ResourceLocked
	PreconditionFailed = proto.PreconditionFailed
	FrameError         = proto.FrameError
	SyntaxError        = proto.SyntaxError
	CommandInvalid     = proto.CommandInvalid
	ChannelError       = proto.ChannelError
	UnexpectedFrame    = proto.UnexpectedFrame
	ResourceError      = proto.ResourceError
	NotAllowed         = proto.NotAllowed
	NotImplemented     = proto.NotImplemented
	InternalError      = proto.InternalError
)

// Constants for standard AMQP 0-9-1 exchange types.
const (
	ExchangeDirect  = "direct"
	ExchangeFanout  = "fanout"
	ExchangeTopic   = "topic"
	ExchangeHeaders = "headers"
)

var (
	// ErrClosed is returned when the channel or connection is not open
	ErrClosed = &Error{Code: ChannelError, Reason: "channel/connection is not open"}

	// ErrChannelMax is returned when Connection.Channel has been called enough
	// times that all channel IDs have been exhausted in the client or the
	// server.
	ErrChannelMax = &Error{Code: ChannelError, Reason: "channel id space exhausted"}

	// ErrSASL is returned from Dial when the authentication mechanism could not
	// be negoated.
	ErrSASL = &Error{Code: AccessRefused, Reason: "SASL could not negotiate a shared mechanism"}

	// ErrCredentials is returned when the authenticated client is not authorized
	// to any vhost.
	ErrCredentials = &Error{Code: AccessRefused, Reason: "username or password not allowed"}

	// ErrVhost is returned when the authenticated user is not permitted to
	// access the requested Vhost.
	ErrVhost = &Error{Code: AccessRefused, Reason: "no access to this vhost"}

	// ErrCommandInvalid is returned when the server sends an unexpected response
	// to this requested message type. This indicates a bug in this client.
	ErrCommandInvalid = &Error{Code: CommandInvalid, Reason: "unexpected command received"}

	// ErrUnexpectedFrame is returned when something other than a method or
	// heartbeat frame is delivered to the Connection, indicating a bug in the
	// client.
	ErrUnexpectedFrame = &Error{Code: UnexpectedFrame, Reason: "unexpected frame received"}

	// ErrSyntax is hard protocol error, indicating an unsupported protocol,
	// implementation or encoding.
	ErrSyntax = proto.ErrSyntax

	// ErrFrame is returned when the protocol frame cannot be read from the
	// server, indicating an unsupported protocol or unsupported frame type.
	ErrFrame = proto.ErrFrame

	// ErrFieldType is returned when writing a message containing a Go type unsupported by AMQP.
	ErrFieldType = proto.ErrFieldType
)

// Error captures the code and reason a channel or connection has been closed
// by the server.
type Error struct {
	Code    int    // constant code from the specification
	Reason  string // description of the error
	Server  bool   // true when initiated from the server, false when from this library
	Recover bool   // true when this error can be recovered by retrying later or with different parameters
}

func newError(code uint16, text string) *Error {
	return &Error{
		Code:    int(code),
		Reason:  text,
		Recover: proto.IsSoftExceptionCode(int(code)),
		Server:  true,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("Exception (%d) Reason: %q", e.Code, e.Reason)
}

// DeliveryMode.  Transient means higher throughput but messages will not be
// restored on broker restart.  The delivery mode of publishings is unrelated
// to the durability of the queues they reside on.  Transient messages will
// not be restored to durable queues, persistent messages will be restored to
// durable queues and lost on non-durable queues during server restart.
//
// This remains typed as uint8 to match Publishing.DeliveryMode.  Other
// delivery modes specific to custom queue implementations are not enumerated
// here.
const (
	Transient  uint8 = 1
	Persistent uint8 = 2
)

// Queue captures the current server state of the queue on the server returned
// from Channel.QueueDeclare or Channel.QueueInspect.
type Queue struct {
	Name      string // server confirmed or generated name
	Messages  int    // count of messages not awaiting acknowledgment
	Consumers int    // number of consumers receiving deliveries
}

// Publishing captures the client message sent to the server.  The fields
// outside of the Headers table included in this struct mirror the underlying
// fields in the content frame.  They use native types for convenience and
// efficiency.
type Publishing struct {
	// Application or exchange specific fields,
	// the headers exchange will inspect this field.
	Headers Table

	// Properties
	ContentType     string    // MIME content type
	ContentEncoding string    // MIME content encoding
	DeliveryMode    uint8     // Transient (0 or 1) or Persistent (2)
	Priority        uint8     // 0 to 9
	CorrelationId   string    // correlation identifier
	ReplyTo         string    // address to to reply to (ex: RPC)
	Expiration      string    // message expiration spec
	MessageId       string    // message identifier
	Timestamp       time.Time // message timestamp
	Type            string    // message type name
	UserId          string    // creating user id - ex: "guest"
	AppId           string    // creating application id

	// The application specific payload of the message
	Body []byte
}

// Blocking notifies the server's TCP flow control of the Connection.  When a
// server hits a memory or disk alarm it will block all connections until the
// resources are reclaimed.  Use NotifyBlock on the Connection to receive these
// events.
type Blocking struct {
	Active bool   // TCP pushback active/inactive on server
	Reason string // Server reason for activation
}

// Confirmation notifies the acknowledgment or negative acknowledgement of a
// publishing identified by its delivery tag.  Use NotifyPublish on the Channel
// to consume these events.
type Confirmation struct {
	DeliveryTag uint64 // A 1 based counter of publishings from when the channel was put in Confirm mode
	Ack         bool   // True when the server successfully received the publishing
}

// Table stores user supplied fields of the following types:
//
//   bool
//   byte
//   float32
//   float64
//   int
//   int16
//   int32
//   int64
//   nil
//   string
//   time.Time
//   amqp.Decimal
//   amqp.Table
//   []byte
//   []interface{} - containing above types
//
// Functions taking a table will immediately fail when the table contains a
// value of an unsupported type.
//
// The caller must be specific in which precision of integer it wishes to
// encode.
//
// Use a type assertion when reading values from a table for type conversion.
//
// RabbitMQ expects int32 for integer values.
//
type Table = proto.Table

// Decimal matches the AMQP decimal type.  Scale is the number of decimal
// digits Scale == 2, Value == 12345, Decimal == 123.45
type Decimal = proto.Decimal

// Heap interface for maintaining delivery tags
type tagSet []uint64

func (set tagSet) Len() int              { return len(set) }
func (set tagSet) Less(i, j int) bool    { return (set)[i] < (set)[j] }
func (set tagSet) Swap(i, j int)         { (set)[i], (set)[j] = (set)[j], (set)[i] }
func (set *tagSet) Push(tag interface{}) { *set = append(*set, tag.(uint64)) }
func (set *tagSet) Pop() interface{} {
	val := (*set)[len(*set)-1]
	*set = (*set)[:len(*set)-1]
	return val
}

type message interface {
	ID() (uint16, uint16)
	Wait() bool
	Read(io.Reader) error
	Write(io.Writer) error
}

type messageWithContent interface {
	message
	GetContent() (proto.Properties, []byte)
	SetContent(proto.Properties, []byte)
}

/*
The base interface implemented as:

2.3.5  frame Details

All frames consist of a header (7 octets), a payload of arbitrary size, and a 'frame-end' octet that detects
malformed frames:

  0      1         3             7                  size+7 size+8
  +------+---------+-------------+  +------------+  +-----------+
  | type | channel |     size    |  |  payload   |  | frame-end |
  +------+---------+-------------+  +------------+  +-----------+
   octet   short         long         size octets       octet

To read a frame, we:

 1. Read the header and check the frame type and channel.
 2. Depending on the frame type, we read the payload and process it.
 3. Read the frame end octet.

In realistic implementations where performance is a concern, we would use
“read-ahead buffering” or “gathering reads” to avoid doing three separate
system calls to read a frame.

*/
type frame interface {
	Write(io.Writer) error
	Channel() uint16
}
