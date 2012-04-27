package wire

import (
	"bytes"
	"io"
	"testing"
)

var (
	// FRAME
	// 0      1         3             7                  size+7 size+8
	// +------+---------+-------------+  +------------+  +-----------+
	// | type | channel |     size    |  |  payload   |  | frame-end |
	// +------+---------+-------------+  +------------+  +-----------+
	//  octet   short         long         size octets       octet 
	//
	frameTune      = []byte{1, 0, 11, 0, 0, 0, 12, 0, 10, 0, 30, 0, 5, 0, 0, 1, 0, 0, 7, 206}
	frameHeartbeat = []byte{8, 0, 11, 0, 0, 0, 0, 206}
	frameBody      = []byte{3, 0, 11, 0, 0, 0, 4, 'o', 'h', 'a', 'i', 206}

	// CONTENT HEADER
	// 0          2        4           12               14
	// +----------+--------+-----------+----------------+------------- - -
	// | class-id | weight | body size | property flags | property list...
	// +----------+--------+-----------+----------------+------------- - -
	//    short     short    long long       short        remainder... 
	//
	frameHeader = []byte{2, 0, 11, 0, 0, 0, 15, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x10, 0x00, 1, 206}

	readShortShort = []byte{1, 0, 11, 0, 0, 0, 0, 206, 1, 0, 2, 0, 0, 0, 1, 206}
	readBadMethod  = []byte{7, 0, 11, 0, 0, 0, 0, 206}
)

func TestReadMethodFrame(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(frameTune))
	f, err := r.NextFrame()
	if err != nil {
		t.Error("Bad read:", err)
	}

	m, ok := f.(MethodFrame)
	if !ok {
		t.Error("Method Type error", m)
	}

	if m.Channel != 11 {
		t.Error("Channel error", m)
	}

	p, ok := m.Method.(ConnectionTune)
	if !ok {
		t.Error("Method cast not ConnectionTune")
	}

	if p.ChannelMax != 5 {
		t.Error("Not decoded ChannelMax")
	}

	if p.FrameMax != 0x0100 {
		t.Error("Not decoded FrameMax", p.FrameMax)
	}

	if p.Heartbeat != 7 {
		t.Error("Not decoded Heartbeat", p.Heartbeat)
	}
}

func TestReadShortFrameShouldReturnErrShortBuffer(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(readShortShort))
	_, err := r.NextFrame()
	if err != io.ErrShortBuffer {
		t.Error("Bad error:", err)
	}
}

func TestReadBadMethod(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(readBadMethod))
	_, err := r.NextFrame()
	if err != ErrBadFrameType {
		t.Error("Bad error:", err)
	}
}

func TestReadContentHeader(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(frameHeader))
	f, err := r.NextFrame()
	if err != nil {
		t.Error("Bad read:", err)
	}

	m, ok := f.(HeaderFrame)
	if !ok {
		t.Error("Header Type error", m)
	}

	if m.Channel != 11 {
		t.Error("Channel error", m)
	}

	if m.Header.Properties.DeliveryMode != 1 {
		t.Error("Delivery Mode missing", m)
	}
}

func TestReadContent(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(frameBody))
	f, err := r.NextFrame()
	if err != nil {
		t.Error("Bad read:", err)
	}

	m, ok := f.(BodyFrame)
	if !ok {
		t.Error("Header Type error", m)
	}

	if m.Channel != 11 {
		t.Error("Channel error", m)
	}

	if string(m.Payload) != "ohai" {
		t.Error("Payload error", m)
	}
}

func TestReadHeartbeatFrame(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(frameHeartbeat))
	f, err := r.NextFrame()
	if err != nil {
		t.Error("Bad read:", err)
	}

	m, ok := f.(HeartbeatFrame)
	if !ok {
		t.Error("Method Type error", m)
	}

	if m.Channel != 11 {
		t.Error("Channel error", m)
	}
}

var ()

func TestWriteMethod(t *testing.T) {
	var buf bytes.Buffer

	f := MethodFrame{
		Channel: 11,
		Method: ConnectionTune{
			ChannelMax: 5,
			FrameMax:   0x100,
			Heartbeat:  7,
		},
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), frameTune) != 0 {
		t.Error("ConnectionTune not written", buf.Bytes(), frameTune)
	}
}

func TestWriteHeartbeat(t *testing.T) {
	var buf bytes.Buffer

	f := HeartbeatFrame{
		Channel: 11,
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), frameHeartbeat) != 0 {
		t.Error("ConnectionTune not written", buf.Bytes(), frameHeartbeat)
	}
}

func TestWriteHeader(t *testing.T) {
	var buf bytes.Buffer

	f := HeaderFrame{
		Channel: 11,
		Header: ContentHeader{
			Class: 10,
			Size:  0,
			Properties: ContentProperties{
				DeliveryMode: 1,
			},
		},
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), frameHeader) != 0 {
		t.Error("Header not written", buf.Bytes(), frameHeader)
	}
}

func TestWriteBody(t *testing.T) {
	var buf bytes.Buffer

	f := BodyFrame{
		Channel: 11,
		Payload: []byte("ohai"),
	}

	f.WriteTo(&buf)

	if bytes.Compare(buf.Bytes(), frameBody) != 0 {
		t.Error("Body not written", buf.Bytes(), frameBody)
	}
}
