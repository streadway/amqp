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
	frameTune       = []byte{1, 0, 11, 0, 0, 0, 12, 0, 10, 0, 30, 0, 5, 0, 0, 1, 0, 0, 7, 206}
	frameShortShort = []byte{1, 0, 11, 0, 0, 0, 0, 206, 1, 0, 2, 0, 0, 0, 1, 206}
	frameBadMethod  = []byte{7, 0, 11, 0, 0, 0, 0, 206}
	frameHeartbeat  = []byte{8, 0, 11, 0, 0, 0, 0, 206}
	frameBody       = []byte{3, 0, 11, 0, 0, 0, 4, 'o', 'h', 'a', 'i', 206}

	// CONTENT HEADER
	// 0          2        4           12               14
	// +----------+--------+-----------+----------------+------------- - -
	// | class-id | weight | body size | property flags | property list...
	// +----------+--------+-----------+----------------+------------- - -
	//    short     short    long long       short        remainder... 
	//
	frameHeader = []byte{2, 0, 11, 0, 0, 0, 15, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x10, 0x00, 1, 206}
)

func TestReadMethodFrame(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(frameTune))
	f, err := r.Read()
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
	r := NewFrameReader(bytes.NewBuffer(frameShortShort))
	_, err := r.Read()
	if err != io.ErrShortBuffer {
		t.Error("Bad error:", err)
	}
}

func TestReadBadMethod(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(frameBadMethod))
	_, err := r.Read()
	if err != ErrBadFrameType {
		t.Error("Bad error:", err)
	}
}

func TestReadContentHeader(t *testing.T) {
	r := NewFrameReader(bytes.NewBuffer(frameHeader))
	f, err := r.Read()
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
	f, err := r.Read()
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
	f, err := r.Read()
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
