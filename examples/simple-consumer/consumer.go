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
	uri          *string = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchangeName *string = flag.String("exchange", "test-exchange", "AMQP exchange name")
	queueName    *string = flag.String("queue", "test-queue", "AMQP queue name")
	routingKey   *string = flag.String("routing-key", "test-key", "AMQP routing key")
	lifetimeStr  *string = flag.String("lifetime", "5s", "lifetime of process before shutdown (0s=infinite)")
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
	c, err := NewConsumer(*uri, *exchangeName, *queueName, *routingKey)
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
	quit    chan chan bool
}

func NewConsumer(amqpURI, exchange, queue, routing string) (*Consumer, error) {
	c := &Consumer{
		Channel: nil,
		quit:    make(chan chan bool),
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
		closeChannel(c.Channel)
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue (%s)", queue)
	queueState, err := channel.QueueDeclare(
		queue,            // name of the queue
		amqp.UntilUnused, // lifetime = auto-delete
		false,            // exclusive
		false,            // noWait
		noArgs,           // arguments
	)
	if err != nil {
		closeChannel(c.Channel)
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}
	if !queueState.Declared {
		closeChannel(c.Channel)
		return nil, fmt.Errorf("Queue Declare: somehow Undeclared")
	}

	log.Printf("declared Queue, binding to Exchange (routing '%s')", routing)
	if err := channel.QueueBind(
		queue,    // name of the queue
		routing,  // routingKey
		exchange, // sourceExchange
		false,    // noWait
		noArgs,   // arguments
	); err != nil {
		closeChannel(c.Channel)
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume")
	deliveries, err := channel.Consume(
		queue,  // name
		false,  // noAck
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		"",     // consumerTag,
		noArgs, // arguments
		nil,    // deliveries (ie. create a deliveries channel for me)
	)
	if err != nil {
		closeChannel(c.Channel)
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go c.handle(deliveries)
	return c, nil
}

func (c *Consumer) Shutdown() {
	q := make(chan bool)
	c.quit <- q
	<-q
}

func (c *Consumer) handle(deliveries chan amqp.Delivery) {
	for {
		select {
		case q := <-c.quit:
			closeChannel(c.Channel)
			q <- true
			return
		case d, ok := <-deliveries:
			if ok {
				log.Printf(
					"got %dB delivery: [%v] %s",
					len(d.Body),
					d.DeliveryTag,
					d.Body,
				)
			} else {
				log.Printf("delivery channel closed")
				c.Shutdown()
				return
			}
		}
	}
}

func closeChannel(c *amqp.Channel) {
	if err := c.Close(); err != nil {
		log.Printf("AMQP Channel Close error: %s", err)
	}
}
