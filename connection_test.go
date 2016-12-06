// Copyright (c) 2016, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

// +build integration

package amqp

import (
	"net"
	"sync"
	"testing"
)

func TestChannelOpenOnAClosedConnectionFails(t *testing.T) {
	conn := integrationConnection(t, "channel on close")

	conn.Close()

	if _, err := conn.Channel(); err != ErrClosed {
		t.Fatalf("channel.open on a closed connection %#v is expected to fail", conn)
	}
}

func TestQueueDeclareOnAClosedConnectionFails(t *testing.T) {
	conn := integrationConnection(t, "queue declare on close")
	ch, _ := conn.Channel()

	conn.Close()

	if _, err := ch.QueueDeclare("an example", false, false, false, false, nil); err != ErrClosed {
		t.Fatalf("queue.declare on a closed connection %#v is expected to return ErrClosed, returned: %#v", conn, err)
	}
}

func TestConcurrentClose(t *testing.T) {
	const concurrency = 32

	conn := integrationConnection(t, "concurrent close")
	defer conn.Close()

	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			err := conn.Close()

			if err == nil {
				t.Log("first concurrent close was successful")
				return
			}

			if err == ErrClosed {
				t.Log("later concurrent close were successful and returned ErrClosed")
				return
			}

			// BUG(st) is this really acceptable? we got a net.OpError before the
			// connection was marked as closed means a race condition between the
			// network connection and handshake state. It should be a package error
			// returned.
			if _, neterr := err.(*net.OpError); neterr {
				t.Logf("unknown net.OpError during close, ignoring: %+v", err)
				return
			}

			// A different/protocol error occurred indicating a race or missed condition
			if _, other := err.(*Error); other {
				t.Fatalf("Expected no error, or ErrClosed, or a net.OpError from conn.Close(), got %#v (%s) of type %T", err, err, err)
			}
		}()
	}
	wg.Wait()
}
