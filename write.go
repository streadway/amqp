package amqp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"
)

func (me *Framer) WriteFrame(frame Frame) error {
	return frame.write(me.w)
}

func (me *MethodFrame) write(w io.Writer) (err error) {
	var payload bytes.Buffer

	if me.Method == nil {
		return errors.New("malformed frame: missing method")
	}

	if err = binary.Write(w, binary.BigEndian, me.ClassId); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, me.MethodId); err != nil {
		return
	}

	if err = me.Method.write(&payload); err != nil {
		return
	}

	return writeFrame(w, FrameMethod, me.ChannelId, payload.Bytes())
}

// Heartbeat
//
// Payload is empty
func (me *HeartbeatFrame) write(w io.Writer) (err error) {
	return writeFrame(w, FrameHeartbeat, me.ChannelId, []byte{})
}

// CONTENT HEADER
// 0          2        4           12               14
// +----------+--------+-----------+----------------+------------- - -
// | class-id | weight | body size | property flags | property list...
// +----------+--------+-----------+----------------+------------- - -
//    short     short    long long       short        remainder... 
//
func (me *HeaderFrame) write(w io.Writer) (err error) {
	var payload bytes.Buffer

	if err = binary.Write(&payload, binary.BigEndian, me.ClassId); err != nil {
		return
	}

	if err = binary.Write(&payload, binary.BigEndian, me.weight); err != nil {
		return
	}

	if err = binary.Write(&payload, binary.BigEndian, me.Size); err != nil {
		return
	}

	// First pass will build the mask to be serialized, second pass will serialize
	// each of the fields that appear in the mask.

	var mask uint16

	if len(me.Properties.ContentType) > 0 {
		mask = mask | flagContentType
	}
	if len(me.Properties.ContentEncoding) > 0 {
		mask = mask | flagContentEncoding
	}
	if me.Properties.Headers != nil && len(me.Properties.Headers) > 0 {
		mask = mask | flagHeaders
	}
	if me.Properties.DeliveryMode > 0 {
		mask = mask | flagDeliveryMode
	}
	if me.Properties.Priority > 0 {
		mask = mask | flagPriority
	}
	if len(me.Properties.CorrelationId) > 0 {
		mask = mask | flagCorrelationId
	}
	if len(me.Properties.ReplyTo) > 0 {
		mask = mask | flagReplyTo
	}
	if len(me.Properties.Expiration) > 0 {
		mask = mask | flagExpiration
	}
	if len(me.Properties.MessageId) > 0 {
		mask = mask | flagMessageId
	}
	if me.Properties.Timestamp.Unix() > 0 {
		mask = mask | flagTimestamp
	}
	if len(me.Properties.Type) > 0 {
		mask = mask | flagType
	}
	if len(me.Properties.UserId) > 0 {
		mask = mask | flagUserId
	}
	if len(me.Properties.AppId) > 0 {
		mask = mask | flagAppId
	}

	if err = binary.Write(&payload, binary.BigEndian, mask); err != nil {
		return
	}

	if hasProperty(mask, flagContentType) {
		if err = writeShortstr(&payload, me.Properties.ContentType); err != nil {
			return
		}
	}
	if hasProperty(mask, flagContentEncoding) {
		if err = writeShortstr(&payload, me.Properties.ContentEncoding); err != nil {
			return
		}
	}
	if hasProperty(mask, flagHeaders) {
		if err = writeTable(&payload, me.Properties.Headers); err != nil {
			return
		}
	}
	if hasProperty(mask, flagDeliveryMode) {
		if err = binary.Write(&payload, binary.BigEndian, me.Properties.DeliveryMode); err != nil {
			return
		}
	}
	if hasProperty(mask, flagPriority) {
		if err = binary.Write(&payload, binary.BigEndian, me.Properties.Priority); err != nil {
			return
		}
	}
	if hasProperty(mask, flagCorrelationId) {
		if err = writeShortstr(&payload, me.Properties.CorrelationId); err != nil {
			return
		}
	}
	if hasProperty(mask, flagReplyTo) {
		if err = writeShortstr(&payload, me.Properties.ReplyTo); err != nil {
			return
		}
	}
	if hasProperty(mask, flagExpiration) {
		if err = writeShortstr(&payload, me.Properties.Expiration); err != nil {
			return
		}
	}
	if hasProperty(mask, flagMessageId) {
		if err = writeShortstr(&payload, me.Properties.MessageId); err != nil {
			return
		}
	}
	if hasProperty(mask, flagTimestamp) {
		if err = binary.Write(&payload, binary.BigEndian, uint64(me.Properties.Timestamp.Unix())); err != nil {
			return
		}
	}
	if hasProperty(mask, flagType) {
		if err = writeShortstr(&payload, me.Properties.Type); err != nil {
			return
		}
	}
	if hasProperty(mask, flagUserId) {
		if err = writeShortstr(&payload, me.Properties.UserId); err != nil {
			return
		}
	}
	if hasProperty(mask, flagAppId) {
		if err = writeShortstr(&payload, me.Properties.AppId); err != nil {
			return
		}
	}

	return writeFrame(w, FrameHeader, me.ChannelId, payload.Bytes())
}

// Body
//
// Payload is one byterange from the full body who's size is declared in the
// Header frame
func (me *BodyFrame) write(w io.Writer) (err error) {
	return writeFrame(w, FrameBody, me.ChannelId, me.Body)
}

func writeFrame(w io.Writer, typ uint8, channel uint16, payload []byte) (err error) {
	end := []byte{FrameEnd}
	size := uint(len(payload))

	_, err = w.Write([]byte{
		byte(typ),
		byte(channel & 0xff00 >> 8),
		byte(channel & 0x00ff >> 0),
		byte(size & 0xff000000 >> 24),
		byte(size & 0x00ff0000 >> 16),
		byte(size & 0x0000ff00 >> 8),
		byte(size & 0x000000ff >> 0),
	})

	if err != nil {
		return
	}

	if _, err = w.Write(payload); err != nil {
		return
	}

	if _, err = w.Write(end); err != nil {
		return
	}

	return
}

func writeShortstr(w io.Writer, s string) (err error) {
	b := []byte(s)

	var length uint8 = uint8(len(b))

	if err = binary.Write(w, binary.BigEndian, length); err != nil {
		return
	}

	if _, err = w.Write(b[:length]); err != nil {
		return
	}

	return
}

func writeLongstr(w io.Writer, s string) (err error) {
	b := []byte(s)

	var length uint32 = uint32(len(b))

	if err = binary.Write(w, binary.BigEndian, length); err != nil {
		return
	}

	if _, err = w.Write(b[:length]); err != nil {
		return
	}

	return
}

func writeField(w io.Writer, value interface{}) (err error) {
	var values []interface{}

	switch v := value.(type) {
	case bool:
		if v {
			values = []interface{}{
				byte('t'),
				uint8(1),
			}
		} else {
			values = []interface{}{
				byte('t'),
				uint8(0),
			}
		}

	case int8: // short-short-int
		values = []interface{}{byte('b'), v}

	case uint8: // short-short-uint
		values = []interface{}{byte('B'), v}

	case int16: // short-int
		values = []interface{}{byte('U'), v}

	case uint16: // short-uint
		values = []interface{}{byte('u'), v}

	case int32: // short-uint
		values = []interface{}{byte('I'), v}

	case uint32: // long-uint
		values = []interface{}{byte('i'), v}

	case int64: // long-long-int
		values = []interface{}{byte('L'), v}

	case uint64: // long-long-uint
		values = []interface{}{byte('l'), v}

	case float32: // float
		values = []interface{}{byte('f'), v}

	case float64: // double
		values = []interface{}{byte('d'), v}

	case Decimal: // decimal-value
		values = []interface{}{byte('d'), v.Scale, v.Value}

	case string: // short-string
		if len(v) < 256 {
			values = []interface{}{byte('s'), []byte(v)}
		} else {
			values = []interface{}{byte('S'), []byte(v)}
		}

	case []interface{}: // field-array
		values = []interface{}{'A', uint32(len(v))}
		if err = binary.Write(w, binary.BigEndian, values); err != nil {
			return
		}
		for _, val := range v {
			writeField(w, val)
		}
		return

	case time.Time: // timestamp
		values = []interface{}{'T', uint64(v.Unix())}

	case Table: // field-table
		if err = binary.Write(w, binary.BigEndian, 'F'); err != nil {
			return
		}
		return writeTable(w, v)

	default: // Unit and unknown types, must exist to maintain offsets for field-array
		values = []interface{}{'V'}
	}

	return binary.Write(w, binary.BigEndian, values)
}

func writeTable(w io.Writer, table Table) (err error) {
	var buf bytes.Buffer

	for key, val := range table {
		if err = writeShortstr(&buf, key); err != nil {
			return
		}
		if err = writeField(&buf, val); err != nil {
			return
		}
	}

	return writeLongstr(w, string(buf.Bytes()))
}
