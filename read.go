// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

/*
Reads a frame from an input stream and returns an interface that can be cast into
one of the following:

   methodFrame
   PropertiesFrame
   bodyFrame
   heartbeatFrame

2.3.5  frame Details

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
func (me *reader) ReadFrame() (frame frame, err error) {
	var scratch [7]byte

	if _, err = io.ReadFull(me.r, scratch[:7]); err != nil {
		return
	}

	typ := uint8(scratch[0])
	channel := binary.BigEndian.Uint16(scratch[1:3])
	size := binary.BigEndian.Uint32(scratch[3:7])

	switch typ {
	case frameMethod:
		if frame, err = me.parseMethodFrame(channel, size); err != nil {
			return
		}

	case frameHeader:
		if frame, err = me.parseHeaderFrame(channel, size); err != nil {
			return
		}

	case frameBody:
		if frame, err = me.parseBodyFrame(channel, size); err != nil {
			return
		}

	case frameHeartbeat:
		if frame, err = me.parseHeartbeatFrame(channel, size); err != nil {
			return
		}

	default:
		return nil, ErrFrame
	}

	if _, err = io.ReadFull(me.r, scratch[:1]); err != nil {
		return
	}

	if scratch[0] != frameEnd {
		return nil, ErrFrame
	}

	return
}

func readShortstr(r io.Reader) (v string, err error) {
	var length uint8
	if err = binary.Read(r, binary.BigEndian, &length); err != nil {
		return
	}

	bytes := make([]byte, length)
	if _, err = io.ReadFull(r, bytes); err != nil {
		return
	}
	return string(bytes), nil
}

func readLongstr(r io.Reader) (v string, err error) {
	var length uint32
	if err = binary.Read(r, binary.BigEndian, &length); err != nil {
		return
	}

	bytes := make([]byte, length)
	if _, err = io.ReadFull(r, bytes); err != nil {
		return
	}
	return string(bytes), nil
}

func readDecimal(r io.Reader) (v Decimal, err error) {
	if err = binary.Read(r, binary.BigEndian, &v.Scale); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &v.Value); err != nil {
		return
	}
	return
}

func readFloat32(r io.Reader) (v float32, err error) {
	if err = binary.Read(r, binary.BigEndian, &v); err != nil {
		return
	}
	return
}

func readFloat64(r io.Reader) (v float64, err error) {
	if err = binary.Read(r, binary.BigEndian, &v); err != nil {
		return
	}
	return
}

func readTimestamp(r io.Reader) (v time.Time, err error) {
	var sec int64
	if err = binary.Read(r, binary.BigEndian, &sec); err != nil {
		return
	}
	return time.Unix(sec, 0), nil
}

/*
	field-table         = long-uint *field-value-pair
	field-value-pair    = field-name field-value
	field-name          = short-string
	field-value         = 't' boolean
											/ 'b' short-short-int
											/ 'B' short-short-uint
											/ 'U' short-int
											/ 'u' short-uint
											/ 'I' long-int
											/ 'i' long-uint
											/ 'L' long-long-int
											/ 'l' long-long-uint
											/ 'f' float
											/ 'd' double
											/ 'D' decimal-value
											/ 's' short-string
											/ 'S' long-string
											/ 'A' field-array
											/ 'T' timestamp
											/ 'F' field-table
											/ 'V'                       ; no field
*/
func readField(r io.Reader) (v interface{}, err error) {
	var typ byte
	if err = binary.Read(r, binary.BigEndian, &typ); err != nil {
		return
	}

	switch typ {
	case 't': // boolean
		var value uint8
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return (value != 0), nil

	case 'b': // short-short-int
		var value int8
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'B': // short-short-uint
		var value uint8
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'U': // short-int
		var value int16
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'u': // short-uint
		var value uint16
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'I': // long-int
		var value int32
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'i': // long-uint
		var value uint32
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'L': // long-long-int
		var value int64
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'l': // long-long-uint
		var value uint64
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'f': // float
		var value float32
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'd': // double
		var value float64
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return
		}
		return value, nil

	case 'D': // decimal-value
		return readDecimal(r)

	case 's': // short-string
		return readShortstr(r)

	case 'S': // long-string
		return readLongstr(r)

	case 'A': // field-array
		var size uint32
		if err = binary.Read(r, binary.BigEndian, &size); err != nil {
			return
		}

		lim := &io.LimitedReader{R: r, N: int64(size)}
		arr := make([]interface{}, 0)
		var field interface{}

		for {
			if field, err = readField(lim); err != nil {
				if err == io.EOF {
					return arr, nil
				} else {
					return nil, err
				}
			}
			arr = append(arr, field)
		}

	case 'T': // timestamp
		return readTimestamp(r)

	case 'F': // field-table
		return readTable(r)

	case 'V': // no field
		return nil, nil
	}

	return nil, ErrSyntax
}

/*
	Field tables are long strings that contain packed name-value pairs.  The
	name-value pairs are encoded as short string defining the name, and octet
	defining the values type and then the value itself.   The valid field types for
	tables are an extension of the native integer, bit, string, and timestamp
	types, and are shown in the grammar.  Multi-octet integer fields are always
	held in network byte order.
*/
func readTable(r io.Reader) (table Table, err error) {
	var nested bytes.Buffer
	var str string

	if str, err = readLongstr(r); err != nil {
		return
	}

	nested.Write([]byte(str))

	table = make(Table)

	for nested.Len() > 0 {
		var key string
		var value interface{}

		if key, err = readShortstr(&nested); err != nil {
			return
		}

		if value, err = readField(&nested); err != nil {
			return
		}

		table[key] = value
	}

	return
}

// Checks if this bit mask matches the flags bitset
func hasProperty(mask uint16, prop int) bool {
	return int(mask)&prop > 0
}

func (me *reader) parseHeaderFrame(channel uint16, size uint32) (frame frame, err error) {
	hf := &headerFrame{
		ChannelId: channel,
	}

	if err = binary.Read(me.r, binary.BigEndian, &hf.ClassId); err != nil {
		return
	}

	if err = binary.Read(me.r, binary.BigEndian, &hf.weight); err != nil {
		return
	}

	if err = binary.Read(me.r, binary.BigEndian, &hf.Size); err != nil {
		return
	}

	var flags uint16

	if err = binary.Read(me.r, binary.BigEndian, &flags); err != nil {
		return
	}

	if hasProperty(flags, flagContentType) {
		if hf.Properties.ContentType, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagContentEncoding) {
		if hf.Properties.ContentEncoding, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagHeaders) {
		if hf.Properties.Headers, err = readTable(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagDeliveryMode) {
		if err = binary.Read(me.r, binary.BigEndian, &hf.Properties.DeliveryMode); err != nil {
			return
		}
	}
	if hasProperty(flags, flagPriority) {
		if err = binary.Read(me.r, binary.BigEndian, &hf.Properties.Priority); err != nil {
			return
		}
	}
	if hasProperty(flags, flagCorrelationId) {
		if hf.Properties.CorrelationId, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagReplyTo) {
		if hf.Properties.ReplyTo, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagExpiration) {
		if hf.Properties.Expiration, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagMessageId) {
		if hf.Properties.MessageId, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagTimestamp) {
		if hf.Properties.Timestamp, err = readTimestamp(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagType) {
		if hf.Properties.Type, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagUserId) {
		if hf.Properties.UserId, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagAppId) {
		if hf.Properties.AppId, err = readShortstr(me.r); err != nil {
			return
		}
	}
	if hasProperty(flags, flagReserved1) {
		if hf.Properties.reserved1, err = readShortstr(me.r); err != nil {
			return
		}
	}

	return hf, nil
}

func (me *reader) parseBodyFrame(channel uint16, size uint32) (frame frame, err error) {
	bf := &bodyFrame{
		ChannelId: channel,
		Body:      make([]byte, size),
	}

	if _, err = io.ReadFull(me.r, bf.Body); err != nil {
		return
	}

	return bf, nil
}

func (me *reader) parseHeartbeatFrame(channel uint16, size uint32) (frame frame, err error) {
	hf := &heartbeatFrame{
		ChannelId: channel,
	}

	if size > 0 {
		panic("Heartbeats should not have a payload")
	}

	return hf, nil
}
