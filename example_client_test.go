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
			fmt.Printf("Push failed: %s\n", err)
		} else {
			fmt.Println("Push succeeded!")
		}
	}
}

// Queue represents a connection to a specific queue.
type Queue struct {
	name            string
	logger          *log.Logger
	connection      *amqp.Connection
	channel         *amqp.Channel
	done            chan bool
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	isReady         bool
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When setting up the channel after a channel exception
	resetupDelay = 2 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to the queue")
	errAlreadyClosed = errors.New("already closed: not connected to the queue")
	errShutdown      = errors.New("queue is shutting down")
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
// notifyConnClose, and then continously attempt to reconnect.
func (queue *Queue) handleReconnect(addr string) {
	for {
		queue.isReady = false
		log.Println("Attempting to connect")

		conn, err := queue.connect(addr)

		if err != nil {
			log.Println("Failed to connect. Retrying...")

			select {
			case <-queue.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := queue.handelResetup(conn); done {
			break
		}
	}
}

// connect will create a new amqp connection
func (queue *Queue) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)

	if err != nil {
		return nil, err
	}

	queue.changeConnection(conn)
	log.Println("Connected!")
	return conn, nil
}

func (queue *Queue) handelResetup(conn *amqp.Connection) bool {
	for {
		queue.isReady = false

		err := queue.setup(conn)

		if err != nil {
			log.Println("Failed to setup queue. Retrying...")

			select {
			case <-queue.done:
				return true
			case <-time.After(resetupDelay):
			}
			continue
		}

		select {
		case <-queue.done:
			return true
		case <-queue.notifyConnClose:
			log.Println("Connection closed. Reconnecting...")
			return false
		case <-queue.notifyChanClose:
			log.Println("Channel closed. Re-running setup...")
		}
	}
}

// setup will setup channel & declare queue
func (queue *Queue) setup(conn *amqp.Connection) error {
	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	err = ch.Confirm(false)

	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(
		queue.name,
		false, // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)

	if err != nil {
		return err
	}

	queue.changeChannel(ch)
	queue.isReady = true
	log.Println("Setup!")

	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (queue *Queue) changeConnection(connection *amqp.Connection) {
	queue.connection = connection
	queue.notifyConnClose = make(chan *amqp.Error)
	queue.connection.NotifyClose(queue.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (queue *Queue) changeChannel(channel *amqp.Channel) {
	queue.channel = channel
	queue.notifyChanClose = make(chan *amqp.Error)
	queue.notifyConfirm = make(chan amqp.Confirmation)
	queue.channel.NotifyClose(queue.notifyChanClose)
	queue.channel.NotifyPublish(queue.notifyConfirm)
}

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are recieved until within the resendTimeout,
// it continuously resends messages until a confirm is recieved.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (queue *Queue) Push(data []byte) error {
	if !queue.isReady {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := queue.UnsafePush(data)
		if err != nil {
			queue.logger.Println("Push failed. Retrying...")
			select {
			case <-queue.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
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
	if !queue.isReady {
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
	if !queue.isReady {
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
	if !queue.isReady {
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
	queue.isReady = false
	return nil
}
