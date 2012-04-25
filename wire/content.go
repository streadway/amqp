package wire

import (
	"io"
)

// The payload of a content header frame 
type ContentHeader struct {
	Class      uint16
	Size       uint64
	Properties ContentProperties
}

type ContentProperties struct {
	ContentType     string    // MIME content type
	ContentEncoding string    // MIME content encoding
	Headers         Table     // Application or header exchange table
	DeliveryMode    uint8     // queue implemention use - non-persistent (1) or persistent (2)
	Priority        uint8     // queue implementation use - 0 to 9
	CorrelationId   string    // application use - correlation identifier
	ReplyTo         string    // application use - address to to reply to (ex: RPC)
	Expiration      string    // implementation use - message expiration spec
	MessageId       string    // application use - message identifier
	Timestamp       Timestamp // application use - message timestamp
	Type            string    // application use - message type name
	UserId          string    // application use - creating user id
	AppId           string    // application use - creating application id
	// Reserved1 shortstr     // was cluster-id - process for buffer consumption
}

// DeliveryMode enum - this isn't a well typed enum because the ContentProperties
// struct is within the wire package so all fields are assume to have
// encoders/decoders.  Promote a new enum type in the client usage instead.
const (
	TransientDelivery  uint8 = 1
	PersistentDelivery uint8 = 2
)

// The property flags are an array of bits that indicate the presence or
// absence of each property value in sequence.  The bits are ordered from most
// high to low - bit 15 indicates the first property.
const (
	FlagContentType     = 0x8000
	FlagContentEncoding = 0x4000
	FlagHeaders         = 0x2000
	FlagDeliveryMode    = 0x1000
	FlagPriority        = 0x0800
	FlagCorrelationId   = 0x0400
	FlagReplyTo         = 0x0200
	FlagExpiration      = 0x0100
	FlagMessageId       = 0x0080
	FlagTimestamp       = 0x0040
	FlagType            = 0x0020
	FlagUserId          = 0x0010
	FlagAppId           = 0x0008
	FlagReserved1       = 0x0004
)

func hasProperty(mask uint16, prop int) bool {
	return int(mask)&prop > 0
}

// Content

// A content header frame has this format:
//
// 0          2        4           12               14
// +----------+--------+-----------+----------------+------------- - -
// | class-id | weight | body size | property flags | property list...
// +----------+--------+-----------+----------------+------------- - -
//    short     short    long long       short        remainder... 
// 
func (me *buffer) NextContentHeader() ContentHeader {
	msg := ContentHeader{}
	msg.Class = me.NextShort()
	_ = me.NextShort() // weight
	msg.Size = me.NextLonglong()

	flags := me.NextShort()

	props := ContentProperties{}

	if hasProperty(flags, FlagContentType) {
		props.ContentType = me.NextShortstr()
	}
	if hasProperty(flags, FlagContentEncoding) {
		props.ContentEncoding = me.NextShortstr()
	}
	if hasProperty(flags, FlagHeaders) {
		props.Headers = me.NextTable()
	}
	if hasProperty(flags, FlagDeliveryMode) {
		props.DeliveryMode = me.NextShortshort()
	}
	if hasProperty(flags, FlagPriority) {
		props.Priority = me.NextShortshort()
	}
	if hasProperty(flags, FlagCorrelationId) {
		props.CorrelationId = me.NextShortstr()
	}
	if hasProperty(flags, FlagReplyTo) {
		props.ReplyTo = me.NextShortstr()
	}
	if hasProperty(flags, FlagExpiration) {
		props.Expiration = me.NextShortstr()
	}
	if hasProperty(flags, FlagMessageId) {
		props.MessageId = me.NextShortstr()
	}
	if hasProperty(flags, FlagTimestamp) {
		props.Timestamp = me.NextTimestamp()
	}
	if hasProperty(flags, FlagType) {
		props.Type = me.NextShortstr()
	}
	if hasProperty(flags, FlagUserId) {
		props.UserId = me.NextShortstr()
	}
	if hasProperty(flags, FlagAppId) {
		props.AppId = me.NextShortstr()
	}
	if hasProperty(flags, FlagReserved1) {
		_ = me.NextShortstr()
	}

	msg.Properties = props

	return msg
}

func (me *ContentHeader) WriteTo(w io.Writer) (int64, error) {
	var header buffer
	var buf buffer
	var mask int

	header.PutShort(me.Class)
	header.PutShort(0) // Weight
	header.PutLonglong(me.Size)

	if len(me.Properties.ContentType) > 0 {
		mask = mask | FlagContentType
		buf.PutShortstr(me.Properties.ContentType)
	}
	if len(me.Properties.ContentEncoding) > 0 {
		mask = mask | FlagContentEncoding
		buf.PutShortstr(me.Properties.ContentEncoding)
	}
	if me.Properties.Headers != nil && len(me.Properties.Headers) > 0 {
		mask = mask | FlagHeaders
		buf.PutTable(me.Properties.Headers)
	}
	if me.Properties.DeliveryMode > 0 {
		mask = mask | FlagDeliveryMode
		buf.PutShortshort(me.Properties.DeliveryMode)
	}
	if me.Properties.Priority > 0 {
		mask = mask | FlagPriority
		buf.PutShortshort(me.Properties.Priority)
	}
	if len(me.Properties.CorrelationId) > 0 {
		mask = mask | FlagCorrelationId
		buf.PutShortstr(me.Properties.CorrelationId)
	}
	if len(me.Properties.ReplyTo) > 0 {
		mask = mask | FlagReplyTo
		buf.PutShortstr(me.Properties.ReplyTo)
	}
	if len(me.Properties.Expiration) > 0 {
		mask = mask | FlagExpiration
		buf.PutShortstr(me.Properties.Expiration)
	}
	if len(me.Properties.MessageId) > 0 {
		mask = mask | FlagMessageId
		buf.PutShortstr(me.Properties.MessageId)
	}
	if me.Properties.Timestamp > 0 {
		mask = mask | FlagTimestamp
		buf.PutTimestamp(me.Properties.Timestamp)
	}
	if len(me.Properties.Type) > 0 {
		mask = mask | FlagType
		buf.PutShortstr(me.Properties.Type)
	}
	if len(me.Properties.UserId) > 0 {
		mask = mask | FlagUserId
		buf.PutShortstr(me.Properties.UserId)
	}
	if len(me.Properties.AppId) > 0 {
		mask = mask | FlagAppId
		buf.PutShortstr(me.Properties.AppId)
	}

	header.PutShort(uint16(mask))
	header.Write(buf.Bytes())

	return io.Copy(w, &header)
}
