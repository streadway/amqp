// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
	"testing"
	"time"
)

type pipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func (p pipe) Read(b []byte) (int, error) {
	return p.r.Read(b)
}
func (p pipe) Write(b []byte) (int, error) {
	return p.w.Write(b)
}
func (p pipe) Close() error {
	p.r.Close()
	p.w.Close()
	return nil
}

type logIO struct {
	t      *testing.T
	prefix string
	proxy  io.ReadWriteCloser
}

func (me *logIO) Read(p []byte) (n int, err error) {
	me.t.Logf("%s reading %d\n", me.prefix, len(p))
	n, err = me.proxy.Read(p)
	if err != nil {
		me.t.Logf("%s read %x: %v\n", me.prefix, p[0:n], err)
	} else {
		me.t.Logf("%s read:\n%s\n", me.prefix, hex.Dump(p[0:n]))
		//fmt.Printf("%s read:\n%s\n", me.prefix, hex.Dump(p[0:n]))
	}
	return
}

func (me *logIO) Write(p []byte) (n int, err error) {
	me.t.Logf("%s writing %d\n", me.prefix, len(p))
	n, err = me.proxy.Write(p)
	if err != nil {
		me.t.Logf("%s write %d, %x: %v\n", me.prefix, len(p), p[0:n], err)
	} else {
		me.t.Logf("%s write %d:\n%s", me.prefix, len(p), hex.Dump(p[0:n]))
		//fmt.Printf("%s write %d:\n%s", me.prefix, len(p), hex.Dump(p[0:n]))
	}
	return
}

func (me *logIO) Close() (err error) {
	err = me.proxy.Close()
	if err != nil {
		me.t.Logf("%s close : %v\n", me.prefix, err)
	} else {
		me.t.Logf("%s close\n", me.prefix, err)
	}
	return
}

func (me *logIO) Test() {
	me.t.Logf("test: %v\n", me)
}
func integrationUri(t *testing.T) (*URI, bool) {
	urlStr := os.Getenv("AMQP_URL")
	if urlStr == "" {
		t.Logf("Skipping; AMQP_URL not found in the environment")
		return nil, false
	}

	uri, err := ParseURI(urlStr)
	if err != nil {
		t.Errorf("Failed to parse integration URI: %s", err)
		return nil, false
	}

	return &uri, true
}

// Returns a conneciton to the AMQP if the AMQP_URL environment
// variable is set and a connnection can be established.
func integrationConnection(t *testing.T, name string) *Connection {
	if uri, ok := integrationUri(t); ok {
		conn, err := Dial(uri.String())
		if err != nil {
			t.Errorf("Failed to connect to integration server: %s", err)
			return nil
		}

		if name != "" {
			conn.conn = &logIO{t, name, conn.conn}
		}

		return conn
	}

	return nil
}

// Returns a connection, channel and delcares a queue when the AMQP_URL is in the environment
func integrationQueue(t *testing.T, name string) (*Connection, *Channel) {
	if conn := integrationConnection(t, name); conn != nil {
		if channel, err := conn.Channel(); err == nil {
			if _, err = channel.QueueDeclare(name, false, true, false, false, nil); err == nil {
				return conn, channel
			}
		}
	}
	return nil, nil
}

// Delegates to integrationConnection and only returns a connection if the
// product is RabbitMQ
func integrationRabbitMQ(t *testing.T, name string) *Connection {
	if conn := integrationConnection(t, "connect"); conn != nil {
		if server, ok := conn.Properties["product"]; ok && server == "RabbitMQ" {
			return conn
		}
	}

	return nil
}

func assertConsumeBody(t *testing.T, messages <-chan Delivery, body []byte) *Delivery {
	select {
	case msg := <-messages:
		if bytes.Compare(msg.Body, body) != 0 {
			t.Fatalf("Message body does not match have: %v expect %v", msg.Body, body)
			return &msg
		}
		return &msg
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("Timeout waiting for %s", body)
		return nil
	}
	panic("unreachable")
}

// Pulls out the CRC and verifies the remaining content against the CRC
func assertMessageCrc32(t *testing.T, msg []byte, assert string) {
	size := binary.BigEndian.Uint32(msg[:4])

	crc := crc32.NewIEEE()
	crc.Write(msg[8:])

	if binary.BigEndian.Uint32(msg[4:8]) != crc.Sum32() {
		t.Fatalf("Message does not match CRC: %s", assert)
	}

	if int(size) != len(msg)-8 {
		t.Fatalf("Message does not match size, should=%d, is=%d: %s", size, len(msg)-8, assert)
	}
}

// Creates a random body size with a leading 32-bit CRC in network byte order
// that verifies the remaining slice
func generateCrc32Random(size int) []byte {
	msg := make([]byte, size+8)
	if _, err := io.ReadFull(rand.Reader, msg); err != nil {
		panic(err)
	}

	crc := crc32.NewIEEE()
	crc.Write(msg[8:])

	binary.BigEndian.PutUint32(msg[0:4], uint32(size))
	binary.BigEndian.PutUint32(msg[4:8], crc.Sum32())

	return msg
}
