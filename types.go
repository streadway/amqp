package amqp

import (
	"errors"
	"io"
	"time"
)

var (
	ErrBadProtocol         = errors.New("Unexpected protocol message")
	ErrUnexpectedMethod    = errors.New("Bad protocol: Received out of order method")
	ErrUnknownMethod       = errors.New("Unknown wire method id")
	ErrUnknownClass        = errors.New("Unknown wire class id")
	ErrUnknownFrameType    = errors.New("Bad frame: unknown type")
	ErrUnknownFieldType    = errors.New("Bad frame: unknown table field")
	ErrBadFrameSize        = errors.New("Bad frame: invalid size")
	ErrBadFrameTermination = errors.New("Bad frame: invalid terminator")

	ProtocolHeader = []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
)

type Message interface {
	id() (uint16, uint16)
	wait() bool
	read(io.Reader) error
	write(io.Writer) error
}

type MessageWithContent interface {
	Message
	GetContent() (Properties, []byte)
	SetContent(Properties, []byte)
}

type Properties struct {
	ContentType     string    // MIME content type
	ContentEncoding string    // MIME content encoding
	Headers         Table     // Application or header exchange table
	DeliveryMode    uint8     // queue implemention use - non-persistent (1) or persistent (2)
	Priority        uint8     // queue implementation use - 0 to 9
	CorrelationId   string    // application use - correlation identifier
	ReplyTo         string    // application use - address to to reply to (ex: RPC)
	Expiration      string    // implementation use - message expiration spec
	MessageId       string    // application use - message identifier
	Timestamp       time.Time // application use - message timestamp
	Type            string    // application use - message type name
	UserId          string    // application use - creating user id
	AppId           string    // application use - creating application id
	reserved1       string    // was cluster-id - process for buffer consumption
}

const (
	TransientDelivery  uint8 = 1
	PersistentDelivery uint8 = 2
)

// The property flags are an array of bits that indicate the presence or
// absence of each property value in sequence.  The bits are ordered from most
// high to low - bit 15 indicates the first property.
const (
	flagContentType     = 0x8000
	flagContentEncoding = 0x4000
	flagHeaders         = 0x2000
	flagDeliveryMode    = 0x1000
	flagPriority        = 0x0800
	flagCorrelationId   = 0x0400
	flagReplyTo         = 0x0200
	flagExpiration      = 0x0100
	flagMessageId       = 0x0080
	flagTimestamp       = 0x0040
	flagType            = 0x0020
	flagUserId          = 0x0010
	flagAppId           = 0x0008
	flagReserved1       = 0x0004
)

//type Message interface {
//	frames(int) []*Frame
//	content() *Content, bool
//	synchronous() bool
//}

type Decimal struct {
	Scale uint8
	Value uint32
}

type Table map[string]interface{}

var (
	// The method in the frame could not be parsed
	ErrBadMethod = errors.New("Bad Frame Method")

	// The content properties in the frame could not be parsed
	ErrBadHeader = errors.New("Bad Frame Header")

	// The content in the frame could not be parsed
	ErrBadContent = errors.New("Bad Frame Content")

	// The frame was not terminated by the special 206 (0xCE) byte
	ErrBadFrameEnd = errors.New("Bad Frame End")

	// The frame type was not recognized
	ErrBadFrameType = errors.New("Bad Frame Type")
)

/*
The base interface implemented as:

2.3.5  Frame Details

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

In realistic implementations where performance is a concern, we would use “read-ahead buffering” or
“gathering reads” to avoid doing three separate system calls to read a frame.

*/
type Frame interface {
	write(io.Writer) error
	channel() uint16
}

type Framer struct {
	r       io.Reader
	w       io.Writer
	scratch [8]byte
}

/*
Intended to run sequentially over a single reader, not threadsafe
*/
func NewFramer(r io.Reader, w io.Writer) *Framer {
	return &Framer{r: r, w: w}
}

/*
Method frames carry the high-level protocol commands (which we call "methods").
One method frame carries one command.  The method frame payload has this format:

  0          2           4
  +----------+-----------+-------------- - -
  | class-id | method-id | arguments...
  +----------+-----------+-------------- - -
     short      short    ...

To process a method frame, we:
 1. Read the method frame payload.
 2. Unpack it into a structure.  A given method always has the same structure,
 so we can unpack the method rapidly.  3. Check that the method is allowed in
 the current context.
 4. Check that the method arguments are valid.
 5. Execute the method.

Method frame bodies are constructed as a list of AMQP data fields (bits,
integers, strings and string tables).  The marshalling code is trivially
generated directly from the protocol specifications, and can be very rapid.
*/
type MethodFrame struct {
	ChannelId uint16
	ClassId   uint16
	MethodId  uint16
	Method    Message
}

func (me *MethodFrame) channel() uint16 { return me.ChannelId }

/*
Heartbeating is a technique designed to undo one of TCP/IP's features, namely
its ability to recover from a broken physical connection by closing only after
a quite long time-out.  In some scenarios we need to know very rapidly if a
peer is disconnected or not responding for other reasons (e.g. it is looping).
Since heartbeating can be done at a low level, we implement this as a special
type of frame that peers exchange at the transport level, rather than as a
class method.
*/
type HeartbeatFrame struct {
	ChannelId uint16
}

func (me *HeartbeatFrame) channel() uint16 { return me.ChannelId }

/*
Certain methods (such as Basic.Publish, Basic.Deliver, etc.) are formally
defined as carrying content.  When a peer sends such a method frame, it always
follows it with a content header and zero or more content body frames.

A content header frame has this format:

    0          2        4           12               14
    +----------+--------+-----------+----------------+------------- - -
    | class-id | weight | body size | property flags | property list...
    +----------+--------+-----------+----------------+------------- - -
      short     short    long long       short        remainder...

We place content body in distinct frames (rather than including it in the
method) so that AMQP may support "zero copy" techniques in which content is
never marshalled or encoded.  We place the content properties in their own
frame so that recipients can selectively discard contents they do not want to
process
*/
type HeaderFrame struct {
	ChannelId  uint16
	ClassId    uint16
	weight     uint16
	Size       uint64
	Properties Properties
}

func (me *HeaderFrame) channel() uint16 { return me.ChannelId }

/*
Content is the application data we carry from client-to-client via the AMQP
server.  Content is, roughly speaking, a set of properties plus a binary data
part.  The set of allowed properties are defined by the Basic class, and these
form the "content header frame".  The data can be any size, and MAY be broken
into several (or many) chunks, each forming a "content body frame".

Looking at the frames for a specific channel, as they pass on the wire, we
might see something like this:

		[method]
		[method] [header] [body] [body]
		[method]
		...
*/
type BodyFrame struct {
	ChannelId uint16
	Body      []byte
}

func (me *BodyFrame) channel() uint16 { return me.ChannelId }
