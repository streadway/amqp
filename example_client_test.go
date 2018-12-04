package amqp_test

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

// This exports a Queue object that wraps this library. It
// automatically reconnects when the connection fails, and
// blocks all pushes until the connection succeeds. It also
// confirms every outgoing message, so none are lost.
// It doesn't automatically ack each message, but leaves that
// to the parent process, since it is usage-dependent.
//
// Try running this in one terminal, and `rabbitmq-server` in another.
// Stop & restart RabbitMQ to see how the queue reacts.
func Example() {
	name := "job_queue"
	addr := "amqp://guest:guest@localhost:5672/"
	queue := New(name, addr)
	message := []byte("message")
	// Attempt to push a message every 2 seconds
	for {
		time.Sleep(time.Second * 3)
		if err := queue.Push(message); err != nil {
			fmt.Printf("Push failed: %s", err)
		} else {
			fmt.Println("Push succeeded!")
		}
	}
}

// Queue represents a connection to a specific queue.
type Queue struct {
	name          string
	logger        *log.Logger
	connection    *amqp.Connection
	channel       *amqp.Channel
	done          chan bool
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to the queue")
	errNotConfirmed  = errors.New("message not confirmed")
	errAlreadyClosed = errors.New("already closed: not connected to the queue")
)

// New creates a new queue instance, and automatically
// attempts to connect to the server.
func New(name string, addr string) *Queue {
	queue := Queue{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		name:   name,
		done:   make(chan bool),
	}
	go queue.handleReconnect(addr)
	return &queue
}

// handleReconnect will wait for a connection error on
// notifyClose, and then continously attempt to reconnect.
func (queue *Queue) handleReconnect(addr string) {
	for {
		queue.isConnected = false
		log.Println("Attempting to connect")
		for !queue.connect(addr) {
			log.Println("Failed to connect. Retrying...")
			time.Sleep(reconnectDelay)
		}
		select {
		case <-queue.done:
			return
		case <-queue.notifyClose:
		}
	}
}

// connect will make a single attempt to connect to
// RabbitMQ. It returns the success of the attempt.
func (queue *Queue) connect(addr string) bool {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return false
	}
	ch, err := conn.Channel()
	if err != nil {
		return false
	}
	ch.Confirm(false)
	_, err = ch.QueueDeclare(
		queue.name,
		false, // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		return false
	}
	queue.changeConnection(conn, ch)
	queue.isConnected = true
	log.Println("Connected!")
	return true
}

// changeConnection takes a new connection to the queue,
// and updates the channel listeners to reflect this.
func (queue *Queue) changeConnection(connection *amqp.Connection, channel *amqp.Channel) {
	queue.connection = connection
	queue.channel = channel
	queue.notifyClose = make(chan *amqp.Error)
	queue.notifyConfirm = make(chan amqp.Confirmation)
	queue.channel.NotifyClose(queue.notifyClose)
	queue.channel.NotifyPublish(queue.notifyConfirm)
}

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are recieved until within the resendTimeout,
// it continuously resends messages until a confirm is recieved.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (queue *Queue) Push(data []byte) error {
	if !queue.isConnected {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := queue.UnsafePush(data)
		if err != nil {
			queue.logger.Println("Push failed. Retrying...")
			continue
		}
		select {
		case confirm := <-queue.notifyConfirm:
			if confirm.Ack {
				queue.logger.Println("Push confirmed!")
				return nil
			}
		case <-time.After(resendDelay):
		}
		queue.logger.Println("Push didn't confirm. Retrying...")
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// recieve the message.
func (queue *Queue) UnsafePush(data []byte) error {
	if !queue.isConnected {
		return errNotConnected
	}
	return queue.channel.Publish(
		"",         // Exchange
		queue.name, // Routing key
		false,      // Mandatory
		false,      // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}

// Stream will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (queue *Queue) Stream() (<-chan amqp.Delivery, error) {
	if !queue.isConnected {
		return nil, errNotConnected
	}
	return queue.channel.Consume(
		queue.name,
		"",    // Consumer
		false, // Auto-Ack
		false, // Exclusive
		false, // No-local
		false, // No-Wait
		nil,   // Args
	)
}

// Close will cleanly shutdown the channel and connection.
func (queue *Queue) Close() error {
	if !queue.isConnected {
		return errAlreadyClosed
	}
	err := queue.channel.Close()
	if err != nil {
		return err
	}
	err = queue.connection.Close()
	if err != nil {
		return err
	}
	close(queue.done)
	queue.isConnected = false
	return nil
}
