package amqp

import (
	"log"
	"testing"
)

func TestChannelOpenOnAClosedConnectionFails(t *testing.T) {
	conn, err := Dial("amqp://guest:guest@localhost:5672/")
	conn.Close()

	if !conn.IsClosed() {
		log.Fatalf("connection %s is expected to be closed", conn)
	}
	
	_, err = conn.Channel()
	if err == nil {
		log.Fatalf("channel.open on a closed connection %s is expected to fail", conn)
	}
}

func TestQueueDeclareOnAClosedConnectionFails(t *testing.T) {
	conn, err := Dial("amqp://guest:guest@localhost:5672/")
	ch, _ := conn.Channel()
	
	conn.Close()

	if !conn.IsClosed() {
		log.Fatalf("connection %s is expected to be closed", conn)
	}
	
	_, err = ch.QueueDeclare("an example", false, false, false, false, nil)
	if err == nil {
		log.Fatalf("queue.declare on a closed connection %s is expected to fail", conn)
	}
}
