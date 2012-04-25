package wire

import (
	"bytes"
	"errors"
	"encoding/binary"
	"fmt"
	"io"
)

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

MethodFrame
HeaderFrame
BodyFrame
HeaderFrame

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
	io.WriterTo
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
	Channel uint16
	Method  Method
}

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
	Channel uint16
}

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
	Channel uint16
	Header  ContentHeader
}

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
	Channel uint16
	Payload []byte
}

func (me MethodFrame) WriteTo(w io.Writer) (int64, error) {
	var payload bytes.Buffer

	if me.Method == nil {
		return 0, errors.New("malformed frame: missing method")
	}

	_, err := me.Method.WriteTo(&payload)
	if err != nil {
		return 0, err
	}

	return writeFrameTo(w, FrameMethod, me.Channel, payload.Bytes())
}

// CONTENT HEADER
// 0          2        4           12               14
// +----------+--------+-----------+----------------+------------- - -
// | class-id | weight | body size | property flags | property list...
// +----------+--------+-----------+----------------+------------- - -
//    short     short    long long       short        remainder... 
//
func (me HeaderFrame) WriteTo(w io.Writer) (int64, error) {
	var payload bytes.Buffer
	_, err := me.Header.WriteTo(&payload)
	if err != nil {
		return 0, err
	}
	return writeFrameTo(w, FrameHeader, me.Channel, payload.Bytes())
}

// Body
//
// Payload is one byterange from the full body who's size is declared in the
// Header frame
func (me BodyFrame) WriteTo(w io.Writer) (n int64, err error) {
	return writeFrameTo(w, FrameBody, me.Channel, me.Payload)
}

// Heartbeat
//
// Payload is empty
func (me HeartbeatFrame) WriteTo(w io.Writer) (n int64, err error) {
	return writeFrameTo(w, FrameHeartbeat, me.Channel, []byte{})
}

// Writes a complete buffer to the io.Writer to avoid fragmentation on the wire
func writeFrameTo(w io.Writer, typ uint8, channel uint16, payload []byte) (int64, error) {
	var frame bytes.Buffer

	size := uint(len(payload))

	frame.Write([]byte{
		byte(typ),
		byte(channel & 0xff00 >> 8),
		byte(channel & 0x00ff),
		byte(size & 0xff000000 >> 24),
		byte(size & 0x00ff0000 >> 16),
		byte(size & 0x0000ff00 >> 8),
		byte(size & 0x000000ff),
	})

	frame.Write(payload)
	frame.WriteByte(FrameEnd)

	return io.Copy(w, &frame)
}

type frameReader struct {
	reader  io.Reader
	scratch [8]byte
	payload *buffer
}

/*
Intended to run sequentially over a single reader, not threadsafe
*/
func NewFrameReader(r io.Reader) *frameReader {
	return &frameReader{reader: r, payload: new(buffer)}
}

/*
Reads a frame from an input stream and returns an interface that can be cast into
one of the following:

   MethodFrame
   PropertiesFrame
   BodyFrame
   HeartbeatFrame

2.3.5  Frame Details

All frames consist of a header (7 octets), a payload of arbitrary size, and a
'frame-end' octet that detects malformed frames:

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
“read-ahead buffering” or

“gathering reads” to avoid doing three separate system calls to read a frame.
*/
func (me *frameReader) Read() (frame Frame, err error) {
	// Capture and recover from any short buffers during read
	defer func() {
		if r := recover(); r != nil {
			println("recovered", r)
			if e, ok := r.(error); ok {
				switch e {
				case io.ErrShortBuffer:
					err = e
					return
				}
			}
			panic(r)
		}
	}()

	if _, err = io.ReadFull(me.reader, me.scratch[:7]); err != nil {
		return
	}

	typ := uint8(me.scratch[0])
	channel := binary.BigEndian.Uint16(me.scratch[1:3])
	size := binary.BigEndian.Uint32(me.scratch[3:7])

	me.payload.Reset()

	if _, err = io.CopyN(me.payload, me.reader, int64(size+1)); err != nil {
		return
	}

	switch typ {
	case FrameMethod:
		frame = MethodFrame{
			Channel: channel,
			Method:  me.payload.NextMethod(),
		}

	case FrameHeader:
		frame = HeaderFrame{
			Channel: channel,
			Header:  me.payload.NextContentHeader(),
		}

	case FrameBody:
		frame = BodyFrame{
			Channel: channel,
			Payload: me.payload.Next(int(size)),
		}

	case FrameHeartbeat:
		frame = HeartbeatFrame{
			Channel: channel,
		}

	default:
		return nil, ErrBadFrameType
	}

	fmt.Println("payload:", me.payload.Bytes())

	if end := me.payload.Next(1)[0]; end != FrameEnd {
		return nil, ErrBadFrameEnd
	}

	return
}
