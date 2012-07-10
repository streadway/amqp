// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"testing"
)

// Returns a conneciton to the AMQP if the AMQP_URL environment
// variable is set and a connnection can be established.
func integrationConnection(t *testing.T, name string) *Connection {
	urlStr := os.Getenv("AMQP_URL")
	if urlStr == "" {
		t.Logf("Skipping; AMQP_URL not found in the environment")
		return nil
	}

	uri, err := ParseURI(urlStr)
	if err != nil {
		t.Errorf("Failed to parse URI: %s", err)
		return nil
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", uri.Host, uri.Port))
	if err != nil {
		t.Errorf("Failed to connect to integration server: %s", err)
		return nil
	}

	c, err := NewConnection(&logIO{t, name, conn}, uri.PlainAuth(), uri.Vhost)
	if err != nil {
		t.Errorf("Failed to create client against integration server: %s", err)
		return nil
	}

	return c
}

func assertMessageBody(t *testing.T, msg Delivery, body []byte) bool {
	if bytes.Compare(msg.Body, body) != 0 {
		t.Errorf("Message body does not match have: %v expect %v", msg.Body, body)
		return false
	}
	return true
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
