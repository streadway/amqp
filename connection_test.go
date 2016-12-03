package amqp

import (
	"log"
	"net"
	"sync"
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

func TestConcurrentClose(t *testing.T) {
	conn, err := Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Could't connect to amqp server, err = %s", err)
	}

	n := 32
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			err := conn.Close()
			if err != nil && err != ErrClosed {
				switch err.(type) {
				case *Error:
					log.Fatalf("Expected no error, or ErrClosed, or a net.OpError from conn.Close(), got %#v (%#v) of type %T", err, err.Error(), err)
				case *net.OpError:
					// this is acceptable: we got a net.OpError
					// before the connection was marked as closed
					break;
				}

			}
			wg.Done()
		}()
	}
	wg.Wait()
}
