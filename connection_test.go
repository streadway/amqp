package amqp

import (
	"log"
	"sync"
	"testing"
)

func TestChannelOpenOnAClosedConnectionFails(t *testing.T) {
	conn, err := Dial("amqp://guest:guest@localhost:5672/")
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

func TestConcurrentClose(t *testing.T) {
	conn, err := Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Could't connect to amqp server, err = %s", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			err := conn.Close()
			if err != nil && err != ErrClosed {
				log.Fatalf("Expected nil or ErrClosed - got %#v, type is %T", err, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
