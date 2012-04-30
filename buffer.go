package amqp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io" // for io.ErrShortBuffer
	"math"
)

var (
	ErrUnknownFieldType = errors.New("Bad wire protocol: Unknown field type")
)

// Handles the encoding and decoding from and to a splice.  It contains
// the encoding rules for the wire format and is extended with the generated
// method parser from the spec.  Buffer overruns are expected to be checked
// and recovered in any container.
type buffer struct {
	bytes.Buffer
}

func newBuffer(data []byte) *buffer {
	return &buffer{
		*bytes.NewBuffer(data),
	}
}

// Overrides bytes.Buffer to do checked reads, will panic with
// io.ErrShortBuffer if there are not enough bytes to satisfy this call
func (me *buffer) Next(n int) []byte {
	next := me.Buffer.Next(n)
	if len(next) < n {
		panic(io.ErrShortBuffer)
	}
	return next
}

// Returns a slice of the next n bytes without consuming, used for Bit fields
// will panic wiht io.ErrShortBuffer if there are not enough bytes in the buffer
// to satisfy this call.
func (me *buffer) Peek(n int) []byte {
	raw := me.Bytes()
	if len(raw) < n {
		panic(io.ErrShortBuffer)
	}
	return raw[len(raw)-n:]
}

// Returns true if the bit is 1 offset starting at the least significant bit
func (me *buffer) NextBit(offset uint8) bool {
	if me.Peek(1)[0]&(1<<offset) > 0 {
		return true
	}
	return false
}

// Consumes and returns the next byte
func (me *buffer) NextOctet() byte {
	return me.Next(1)[0]
}

// Consumes and returns the next 2 bytes in BigEndian order
func (me *buffer) NextShort() uint16 {
	return binary.BigEndian.Uint16(me.Next(2))
}

// Consumes and returns the next 4 bytes
func (me *buffer) NextLong() uint32 {
	return binary.BigEndian.Uint32(me.Next(4))
}

// Consumes and returns the next 8 bytes in BigEndian order
func (me *buffer) NextLonglong() uint64 {
	return binary.BigEndian.Uint64(me.Next(8))
}

// Consumes and returns the next 4 bytes in IEEE packing
func (me *buffer) NextFloat() float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(me.Next(4)))
}

// Consumes and returns the next 8 bytes in IEEE packing
func (me *buffer) NextDouble() float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(me.Next(8)))
}

// Consumes 1 byte for the length and 0-255 bytes as a UTF-8 encoded string
func (me *buffer) NextShortstr() string {
	size := int(me.NextOctet())
	return string(me.Next(size))
}

// Consumes 4 bytes for the length and 0-2^31 bytes as a UTF-8 encoded string
func (me *buffer) NextLongstr() string {
	size := int(me.NextLong())
	return string(me.Next(size))
}

// Consumes 1 byte and return true if the byte is not 0
func (me *buffer) NextBool() bool {
	if me.NextOctet() != 0 {
		return true
	}
	return false
}

// Consumes 1 byte
func (me *buffer) NextShortshort() uint8 {
	return uint8(me.NextOctet())
}

// Consumes 1 byte
func (me *buffer) NextShortshortSigned() int8 {
	return int8(me.NextOctet())
}

// Consumes 2 bytes
func (me *buffer) NextShortSigned() int16 {
	return int16(me.NextShort())
}

// Consumes 4 bytes
func (me *buffer) NextLongSigned() int32 {
	return int32(me.NextLong())
}

// Consumes 8 bytes
func (me *buffer) NextLonglongSigned() int64 {
	return int64(me.NextLong())
}

// Consumes 8 bytes
func (me *buffer) NextTimestamp() Timestamp {
	return Timestamp(me.NextLonglong())
}

// Consumes 5 bytes
func (me *buffer) NextDecimal() Decimal {
	return Decimal{
		Scale: me.NextOctet(),
		Value: me.NextLong(),
	}
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
func (me *buffer) NextField() interface{} {
	switch me.NextOctet() {
	case 't': // boolean
		return me.NextBool()

	case 'b': // short-short-int
		return me.NextShortshortSigned()

	case 'B': // short-short-uint
		return me.NextShortshort()

	case 'U': // short-int
		return me.NextShortSigned()

	case 'u': // short-uint
		return me.NextShort()

	case 'I': // long-int
		return me.NextLongSigned()

	case 'i': // long-uint
		return me.NextLong()

	case 'L': // long-long-int
		return me.NextLonglongSigned()

	case 'l': // long-long-uint
		return me.NextLonglong()

	case 'f': // float
		return me.NextFloat()

	case 'd': // double
		return me.NextDouble()

	case 'D': // decimal-value
		return me.NextDecimal()

	case 's': // short-string
		return me.NextShortstr()

	case 'S': // long-string
		return me.NextLongstr()

	case 'A': // field-array
		size := int(me.NextLong())
		array := make([]interface{}, size)
		for i, _ := range array {
			array[i] = me.NextField()
		}
		return array

	case 'T': // timestamp
		return me.NextTimestamp()

	case 'F': // field-table
		return me.NextTable()

	case 'V': // no field
		return Unit{}
	}

	panic(ErrUnknownFieldType)
}

/*
	Field tables are long strings that contain packed name-value pairs.  The
	name-value pairs are encoded as short string defining the name, and octet
	defining the values type and then the value itself.   The valid field types for
	tables are an extension of the native integer, bit, string, and timestamp
	types, and are shown in the grammar.  Multi-octet integer fields are always
	held in network byte order.
*/
func (me *buffer) NextTable() Table {
	var nested buffer
	nested.Write([]byte(me.NextLongstr()))

	table := make(Table)

	for nested.Len() > 0 {
		key := nested.NextShortstr()
		value := nested.NextField()
		table[key] = value
	}

	return table
}

// Writers

func (me *buffer) PutBit(on bool, offset uint8) {
	raw := me.Peek(1)
	if on {
		raw[0] = raw[0] | 1<<offset
	} else {
		raw[0] = raw[0] & ^(1 << offset)
	}
}

func (me *buffer) PutOctet(value byte) {
	me.WriteByte(value)
}

func (me *buffer) PutShort(value uint16) {
	var tmp [2]byte
	binary.BigEndian.PutUint16(tmp[:], value)
	me.Write(tmp[:])
}

func (me *buffer) PutLong(value uint32) {
	var tmp [4]byte
	binary.BigEndian.PutUint32(tmp[:], value)
	me.Write(tmp[:])
}

func (me *buffer) PutLonglong(value uint64) {
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], value)
	me.Write(tmp[:])
}

func (me *buffer) PutShortstr(value string) {
	size := byte(len(value))
	me.PutOctet(size)
	me.Write([]byte(value[:size]))
}

func (me *buffer) PutLongstr(value string) {
	size := uint32(len(value))
	me.PutLong(size)
	me.Write([]byte(value[:size]))
}

// Composites
func (me *buffer) PutBool(value bool) {
	if value == true {
		me.PutOctet(1)
	} else {
		me.PutOctet(0)
	}
}

func (me *buffer) PutShortshort(value uint8) {
	me.PutOctet(byte(value))
}

func (me *buffer) PutShortshortSigned(value int8) {
	me.PutOctet(byte(value))
}

func (me *buffer) PutShortSigned(value int16) {
	me.PutShort(uint16(value))
}

func (me *buffer) PutLongSigned(value int32) {
	me.PutLong(uint32(value))
}

func (me *buffer) PutLonglongSigned(value int64) {
	me.PutLonglong(uint64(value))
}

func (me *buffer) PutFloat(value float32) {
	me.PutLong(math.Float32bits(value))
}

func (me *buffer) PutDouble(value float64) {
	me.PutLonglong(math.Float64bits(value))
}

func (me *buffer) PutTimestamp(value Timestamp) {
	me.PutLonglong(uint64(value))
}

func (me *buffer) PutDecimal(value Decimal) {
	me.PutOctet(value.Scale)
	me.PutLong(value.Value)
}

func (me *buffer) PutField(value interface{}) {
	switch v := value.(type) {
	case bool:
		me.PutOctet('t')
		me.PutBool(v)

	case int8: // short-short-int
		me.PutOctet('b')
		me.PutShortshortSigned(v)

	case uint8: // short-short-uint
		me.PutOctet('B')
		me.PutShortshort(v)

	case int16: // short-int
		me.PutOctet('U')
		me.PutShortSigned(v)

	case uint16: // short-uint
		me.PutOctet('u')
		me.PutShort(v)

	case int32: // short-uint
		me.PutOctet('I')
		me.PutLongSigned(v)

	case uint32: // long-uint
		me.PutOctet('i')
		me.PutLong(v)

	case int64: // long-long-int
		me.PutOctet('L')
		me.PutLonglongSigned(v)

	case uint64: // long-long-uint
		me.PutOctet('l')
		me.PutLonglong(v)

	case float32: // float
		me.PutOctet('f')
		me.PutFloat(v)

	case float64: // double
		me.PutOctet('d')
		me.PutDouble(v)

	case Decimal: // decimal-value
		me.PutOctet('D')
		me.PutDecimal(v)

	case string: // short-string
		if len(v) < 256 {
			me.PutOctet('s')
			me.PutShortstr(v)
		} else {
			me.PutOctet('S')
			me.PutLongstr(v)
		}

	case []interface{}: // field-array
		me.PutOctet('A')
		me.PutLong(uint32(len(v)))
		for _, val := range v {
			me.PutField(val)
		}

	case Timestamp: // timestamp
		me.PutOctet('T')
		me.PutTimestamp(v)

	case Table: // field-table
		me.PutOctet('F')
		me.PutTable(v)

	default: // Unit and unknown types, must exist to maintain offsets for field-array
		me.PutOctet('V')
	}

	return
}

func (me *buffer) PutTable(value Table) {
	var nested buffer

	for key, val := range value {
		nested.PutShortstr(key)
		nested.PutField(val)
	}

	me.PutLongstr(string(nested.Bytes()))
}
