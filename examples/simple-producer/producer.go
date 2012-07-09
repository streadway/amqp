// This example declares a durable Exchange, and publishes a single message to
// that Exchange with a given routing key.
//
package main

import (
	"github.com/streadway/amqp"
	"fmt"
	"log"
	"flag"
)

var (
	uri          *string = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchangeName *string = flag.String("exchange", "test-exchange", "AMQP exchange name")
	routingKey   *string = flag.String("routing-key", "test-key", "AMQP routing key")
	body         *string = flag.String("body", "foobar", "Body of message")
)

func init() {
	flag.Parse()
}

func main() {
	if err := publish(*uri, *exchangeName, *routingKey, *body); err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("published %dB OK", len(*body))
}

func publish(amqpURI, exchange, routingKey, body string) error {

	// This function dials, connects, declares, publishes, and tears down,
	// all in one go. In a real service, you probably want to maintain a
	// long-lived connection as state, and publish against that.

	log.Printf("dialing %s", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}
	defer closeChannel(channel)

	log.Printf("got Channel, declaring Exchange (%s)", exchange)
	noArgs := amqp.Table{}
	e := channel.E(exchange)
	if err := e.Declare(
		amqp.UntilDeleted, // lifetime = durable
		"direct",          // type
		false,             // internal
		false,             // noWait
		noArgs,            // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, publishing %dB body (%s)", len(body), body)
	err = e.Publish(
		routingKey,
		true, // mandatory
		true, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    1, // 1=non-persistent, 2=persistent
			Priority:        0, // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
	if err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}

	return nil
}

func closeChannel(c *amqp.Channel) {
	if err := c.Close(); err != nil {
		log.Printf("AMQP Channel Close error: %s", err)
	}
}
