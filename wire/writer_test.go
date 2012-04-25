package wire

import (
	"bytes"
	"testing"
)

var (
	writeConnectionTune = []byte{1, 0, 11, 0, 0, 0, 12, 0, 10, 0, 30, 0, 5, 0, 0, 1, 0, 0, 6, 206}
	writeHeartbeat = []byte{8, 0, 11, 0, 0, 0, 0, 206}
	writeBody = []byte{3, 0, 11, 0, 0, 0, 4, 'o', 'h', 'a', 'i', 206}

	// CONTENT HEADER
	// 0          2        4           12               14
	// +----------+--------+-----------+----------------+------------- - -
	// | class-id | weight | body size | property flags | property list...
	// +----------+--------+-----------+----------------+------------- - -
	//    short     short    long long       short        remainder... 
	//
	writeHeader = []byte{2, 0, 11, 0, 0, 0, 15, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x10, 0x00, 1, 206}
)

func TestWriteMethod(t *testing.T) {
	var buf bytes.Buffer

	f := MethodFrame{
		Channel: 11,
		Method: ConnectionTune{
			ChannelMax: 5,
			FrameMax:   0x100,
			Heartbeat:  6,
		},
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), writeConnectionTune) != 0 {
		t.Error("ConnectionTune not written", buf.Bytes(), writeConnectionTune)
	}
}

func TestWriteHeartbeat(t *testing.T) {
	var buf bytes.Buffer

	f := HeartbeatFrame{
		Channel: 11,
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), writeHeartbeat) != 0 {
		t.Error("ConnectionTune not written", buf.Bytes(), writeHeartbeat)
	}
}

func TestWriteHeader(t *testing.T) {
	var buf bytes.Buffer

	f := HeaderFrame{
		Channel: 11,
		Header: ContentHeader {
			Class: 10,
			Size: 0,
			Properties: ContentProperties {
				DeliveryMode: 1,
			},
		},
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), writeHeader) != 0 {
		t.Error("Header not written", buf.Bytes(), writeHeader)
	}
}

func TestWriteBody(t *testing.T) {
	var buf bytes.Buffer

	f := BodyFrame{
		Channel: 11,
		Payload: []byte("ohai"),
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), writeBody) != 0 {
		t.Error("Body not written", buf.Bytes(), writeBody)
	}
}
