package amqp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"net/url"
	"os"
	"testing"
)

func parseUrl(amqp string) (hostport string, username string, password string, vhost string) {
	u, err := url.Parse(amqp)
	if err != nil {
		return
	}

	host, port := "localhost", 5672
	username, password, vhost = "guest", "guest", "/"

	fmt.Sscanf(u.Host, "%s:%d", &host, &port)

	hostport = fmt.Sprintf("%s:%d", host, port)

	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	if u.Path != "" {
		vhost = u.Path
	}

	return
}

// Returns a conneciton to the AMQP if the AMQP_URL environment
// variable is set and a connnection can be established.
func integrationConnection(t *testing.T, name string) *Connection {
	u := os.Getenv("AMQP_URL")
	if u == "" {
		t.Log("Skipping integration tests, AMQP_URL not found in the environment")
		return nil
	}

	hostport, username, password, vhost := parseUrl(u)

	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		t.Error("Failed to connect to integration server:", err)
		return nil
	}

	//io := conn
	io := &logIO{t, name, conn}

	c, err := NewConnection(io, &PlainAuth{username, password}, vhost)

	if err != nil {
		t.Error("Failed to create client against integration server:", err)
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
		t.Fatal("Message does not match CRC", assert)
	}

	if int(size) != len(msg)-8 {
		t.Fatal("Message does not match size, should:", size, "is:", len(msg)-8, assert)
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
