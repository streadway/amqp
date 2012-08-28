package amqp_test

import (
	"github.com/streadway/amqp"
	"log"
	"runtime"
	"time"
)

func ExampleChannel_Confirm_bridge() {
	// This example acts as a bridge, shoveling all messages sent from the source
	// exchange "log" to destination exchange "log".

	// Confirming publishes can help from overproduction and ensure every message
	// is delivered.

	// Setup the source of the store and forward
	source, err := amqp.Dial("amqp://source/")
	if err != nil {
		log.Fatalf("connection.open source: %s", err)
	}
	defer source.Close()

	chs, err := source.Channel()
	if err != nil {
		log.Fatalf("channel.open source: %s", err)
	}

	if err := chs.ExchangeDeclare("log", "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("exchange.declare destination: %s", err)
	}

	if _, err := chs.QueueDeclare("remote-tee", true, true, false, false, nil); err != nil {
		log.Fatalf("queue.declare source: %s", err)
	}

	if err := chs.QueueBind("remote-tee", "#", "logs", false, nil); err != nil {
		log.Fatalf("queue.bind source: %s", err)
	}

	shovel, err := chs.Consume("remote-tee", "shovel", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("basic.consume source: %s", err)
	}

	// Setup the destination of the store and forward
	destination, err := amqp.Dial("amqp://destination/")
	if err != nil {
		log.Fatalf("connection.open destination: %s", err)
	}
	defer destination.Close()

	chd, err := destination.Channel()
	if err != nil {
		log.Fatalf("channel.open destination: %s", err)
	}

	if err := chd.ExchangeDeclare("log", "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("exchange.declare destination: %s", err)
	}

	pubAcks, pubNacks := chd.NotifyConfirm(make(chan uint64), make(chan uint64))

	if err := chd.Confirm(false); err != nil {
		log.Fatalf("confirm.select destination: %s", err)
	}

	// Now pump the messages, one by one, a smarter implementation
	// would batch the deliveries and use multiple ack/nacks
	for {
		msg, ok := <-shovel
		if !ok {
			log.Fatalf("source channel closed, see the reconnect example for handling this")
		}

		err = chd.Publish("logs", msg.RoutingKey, false, false, amqp.Publishing{
			// Copy all the properties
			ContentType:     msg.ContentType,
			ContentEncoding: msg.ContentEncoding,
			DeliveryMode:    msg.DeliveryMode,
			Priority:        msg.Priority,
			CorrelationId:   msg.CorrelationId,
			ReplyTo:         msg.ReplyTo,
			Expiration:      msg.Expiration,
			MessageId:       msg.MessageId,
			Timestamp:       msg.Timestamp,
			Type:            msg.Type,
			UserId:          msg.UserId,
			AppId:           msg.AppId,

			// Custom headers
			Headers: msg.Headers,

			// And the body
			Body: msg.Body,
		})

		if err != nil {
			msg.Nack(false, false)
			log.Fatalf("basic.publish destination: %s", msg)
		}

		// only ack the source delivery when the destination acks the publishing
		// here you could check for delivery order by keeping a local state of
		// expected delivery tags
		select {
		case <-pubAcks:
			msg.Ack(false)
		case <-pubNacks:
			msg.Nack(false, false)
		}
	}
}

func ExampleChannel_Consume() {
	// Connects opens an AMQP connection from the credentials in the URL.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}
	defer conn.Close()

	c, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel.open: %s", err)
	}

	// We declare our topology on both the publisher and consumer to ensure they
	// are the same.  This is part of AMQP being a programmable messaging model.
	//
	// See the Channel.Publish example for the complimentary declare.
	err = c.ExchangeDeclare("logs", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatal("exchange.declare: %s", err)
	}

	// Establish our queue topologies that we are responsible for
	type bind struct {
		queue string
		key   string
	}

	bindings := []bind{
		bind{"page", "alert"},
		bind{"email", "info"},
		bind{"firehose", "#"},
	}

	for _, b := range bindings {
		_, err = c.QueueDeclare(b.queue, true, false, false, false, nil)
		if err != nil {
			log.Fatal("queue.declare: %s", err)
		}

		err = c.QueueBind(b.queue, b.key, "logs", false, nil)
		if err != nil {
			log.Fatal("queue.bind: %s", err)
		}
	}

	// Set our quality of service.  Since we're sharing 3 consumers on the same
	// channel, we want at least 3 messages in flight.
	err = c.Qos(3, 0, false)
	if err != nil {
		log.Fatal("basic.qos: %s", err)
	}

	// Establish our consumers that have different responsibilities.  Our first
	// two queues do not ack the messages on the server, so require to be acked
	// on the client.

	pages, err := c.Consume("page", "pager", false, false, false, false, nil)
	if err != nil {
		log.Fatal("basic.consume: %s", err)
	}

	go func() {
		for log := range pages {
			// ... this consumer is responsible for sending pages per log
			log.Ack(false)
		}
	}()

	// Notice how the concern for which messages arrive here are in the AMQP
	// topology and not in the queue.  We let the server pick a consumer tag this
	// time.

	emails, err := c.Consume("email", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("basic.consume: %s", err)
	}

	go func() {
		for log := range emails {
			// ... this consumer is responsible for sending emails per log
			log.Ack(false)
		}
	}()

	// This consumer requests that every message is acknowledged as soon as it's
	// delivered.

	firehose, err := c.Consume("firehose", "", true, false, false, false, nil)
	if err != nil {
		log.Fatal("basic.consume: %s", err)
	}

	// To show how to process the items in parallel, we'll use a work pool.
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(work <-chan amqp.Delivery) {
			for _ = range work {
				// ... this consumer pulls from the firehose and doesn't need to acknowledge
			}
		}(firehose)
	}

	// Wait until you're ready to finish, could be a signal handler here.
	time.Sleep(10 * time.Second)

	// Cancelling a consumer by name will finish the range and gracefully end the
	// goroutine
	err = c.Cancel("pager", false)
	if err != nil {
		log.Fatal("basic.cancel: %s", err)
	}

	// deferred closing the Connection will also finish the consumer's ranges of
	// their delivery chans.  If you need every delivery to be processed, make
	// sure to wait for all consumers goroutines to finish before exiting your
	// process.
}

func ExampleChannel_Publish() {
	// Connects opens an AMQP connection from the credentials in the URL.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}

	// This waits for a server acknowledgment which means the sockets will have
	// flushed all outbound publishings prior to returning.  It's important to
	// block on Close to not lose any publishings.
	defer conn.Close()

	c, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel.open: %s", err)
	}

	// We declare our topology on both the publisher and consumer to ensure they
	// are the same.  This is part of AMQP being a programmable messaging model.
	//
	// See the Channel.Consume example for the complimentary declare.
	err = c.ExchangeDeclare("logs", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatal("exchange.declare: %s", err)
	}

	// Prepare this message to be persistent.  Your publishing requirements may
	// be different.
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         []byte("Go Go AMQP!"),
	}

	// This is not a mandatory delivery, so it will be dropped if there are no
	// queues bound to the logs exchange.
	err = c.Publish("logs", "info", false, false, msg)
	if err != nil {
		// Since publish is asynchronous this can happen if the network connection
		// is reset or if the server has run out of resources.
		log.Fatal("basic.publish: %s", err)
	}
}
