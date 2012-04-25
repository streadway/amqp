package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Method interface {
	io.WriterTo
}

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
