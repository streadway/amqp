// This example declares a durable Exchange, an ephemeral (auto-delete) Queue,
// binds the Queue to the Exchange with a routing key, and consumes every
// message published to that Exchange with that routing key.
//
package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchangeName = flag.String("exchange", "test-exchange", "AMQP exchange name")
	queueName    = flag.String("queue", "test-queue", "AMQP queue name")
	routingKey   = flag.String("routing-key", "test-key", "AMQP routing key")
	consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	lifetimeStr  = flag.String("lifetime", "5s", "lifetime of process before shutdown (0s=infinite)")
	lifetime     time.Duration
)

func init() {
	flag.Parse()
	var err error
	if lifetime, err = time.ParseDuration(*lifetimeStr); err != nil {
		log.Fatalf("%s: invalid -lifetime", *lifetimeStr)
	}
}

func main() {
	c, err := NewConsumer(*uri, *exchangeName, *queueName, *routingKey, *consumerTag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	if lifetime == 0 {
		log.Printf("running forever")
		select {}
	} else {
		log.Printf("running for %s", lifetime)
		time.Sleep(lifetime)
	}

	log.Printf("shutting down")
	c.Shutdown()
}

type Consumer struct {
	Channel *amqp.Channel
	tag     string
	done    chan bool
}

func NewConsumer(amqpURI, exchange, queue, routing, ctag string) (*Consumer, error) {
	c := &Consumer{
		Channel: nil,
		tag:     ctag,
		done:    make(chan bool),
	}

	log.Printf("dialing %s", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}
	c.Channel = channel

	log.Printf("got Channel, declaring Exchange (%s)", exchange)
	noArgs := amqp.Table{}
	if err := c.Channel.ExchangeDeclare(
		exchange,          // name of the exchange
		amqp.UntilDeleted, // lifetime = durable
		amqp.Direct,       // type
		false,             // internal
		false,             // noWait
		noArgs,            // arguments
	); err != nil {
		c.closeChannel()
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue (%s)", queue)
	queueState, err := c.Channel.QueueDeclare(
		queue,            // name of the queue
		amqp.UntilUnused, // lifetime = auto-delete
		false,            // exclusive
		false,            // noWait
		noArgs,           // arguments
	)
	if err != nil {
		c.closeChannel()
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}
	if !queueState.Declared {
		c.closeChannel()
		return nil, fmt.Errorf("Queue Declare: somehow Undeclared")
	}

	log.Printf("declared Queue, binding to Exchange (routing '%s')", routing)
	if err := c.Channel.QueueBind(
		queue,    // name of the queue
		routing,  // routingKey
		exchange, // sourceExchange
		false,    // noWait
		noArgs,   // arguments
	); err != nil {
		c.closeChannel()
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag '%s')", c.tag)
	deliveries, err := c.Channel.Consume(
		queue,  // name
		false,  // noAck
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		c.tag,  // consumerTag,
		noArgs, // arguments
		nil,    // deliveries (ie. create a deliveries channel for me)
	)
	if err != nil {
		c.closeChannel()
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go c.handle(deliveries)
	return c, nil
}

func (c *Consumer) Shutdown() {
	c.Channel.Cancel(c.tag, true) // will Close() the channel
	<-c.done                      // wait for handle() to exit
}

func (c *Consumer) handle(deliveries chan amqp.Delivery) {
	for d := range deliveries {
		log.Printf(
			"got %dB delivery: [%v] %s",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
	}
	log.Printf("handle: deliveries channel closed")
	c.done <- true
}

func (c *Consumer) closeChannel() {
	if c.Channel == nil {
		log.Printf("AMQP channel already closed")
		return
	}
	defer func() { c.Channel = nil }()

	if err := c.Channel.Close(); err != nil {
		log.Printf("AMQP channel close error: %s", err)
		return
	}

	log.Printf("AMQP channel closed OK")
}
