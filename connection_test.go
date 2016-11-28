package amqp

import (
	"log"
	"testing"
)

func TestChannelOpenOnAClosedConnectionFails(t *testing.T) {
	conn, err := Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Could't connect to amqp server, err = %s", err)
	}
	conn.Close()

	if !conn.IsClosed() {
		log.Fatalf("connection %s is expected to be closed", conn)
	}

	_, err = conn.Channel()
	if err != ErrClosed {
		log.Fatalf("channel.open on a closed connection %s is expected to fail", conn)
	}
}

func TestQueueDeclareOnAClosedConnectionFails(t *testing.T) {
	conn, err := Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Could't connect to amqp server, err = %s", err)
	}
	ch, _ := conn.Channel()

	conn.Close()

	if !conn.IsClosed() {
		log.Fatalf("connection %s is expected to be closed", conn)
	}

	_, err = ch.QueueDeclare("an example", false, false, false, false, nil)
	if err != ErrClosed {
		log.Fatalf("queue.declare on a closed connection %s is expected to fail", conn)
	}
}
