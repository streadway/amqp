package amqp_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/streadway/amqp"
)

func ExampleConfig_timeout() {
	// Provide your own anonymous Dial function that delgates to net.DialTimout
	// for custom timeouts

	conn, err := amqp.DialConfig("amqp:///", amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 2*time.Second)
		},
	})

	log.Printf("conn: %v, err: %v", conn, err)
}

func ExampleDialTLS() {
	// This is a step-by-step guide to configure TLS in RabbitMQ 3.6.9
	// running on OS X 10.11
	//
	// RabbitMQ was installed via Homebrew:
	//
	// 	brew install rabbitmq
	//
	// 1. Create a self-signed CA certificate, client and server key pairs
	// using tls-gen:
	//
	// 	cd ~
	// 	git clone https://github.com/michaelklishin/tls-gen
	// 	cd tls-gen/basic
	// 	make
	//
	// If you get stuck, refer to the tls-gen README: https://github.com/michaelklishin/tls-gen
	//
	// 2. Configure TLS in RabbitMQ, edit /usr/local/etc/rabbitmq/rabbitmq.config:
	//
	//   	[
	//   		{rabbit, [
	//   		  	{tcp_listeners, []},     			% disable plain AMQP port
	//   		  	{ssl_listeners, ["127.0.0.1", 5671]}, 		% enable SSL AMQP port
	//   		  	{ssl_options, [{cacertfile,"~/tls-gen/basic/result/ca_certificate.pem"},
	//   		  	               {certfile,"~/tls-gen/basic/result/server_certificate.pem"},
	//   		  	               {keyfile,"~/tls-gen/basic/result/server_key.pem"},
	//   		  	               {verify,verify_peer},
	//   		  	               {fail_if_no_peer_cert,true}]} 	% fail auth if no certificate is present
	//   		  ]}
	//   	].
	//
	// 3. Start RabbitMQ (stop any running instances first)
	//
	//      rabbitmq-server
	//
	// 4. Confirm that RabbitMQ is listening on 127.0.0.1:5671
	//
	// 	rabbitmqctl status | grep listeners -A 5
	//
	// A comprehensive RabbitMQ TLS guide, suitable for production
	// deployments, can be found at http://www.rabbitmq.com/ssl.html
	cfg := new(tls.Config)

	// The self-signing CA certificate must be included in the RootCAs to
	// be trusted, otherwise the server certificate will fail certificate
	// verification.
	cfg.RootCAs = x509.NewCertPool()

	if ca, err := ioutil.ReadFile("~/tls-gen/basic/result/ca_certificate.pem"); err == nil {
		cfg.RootCAs.AppendCertsFromPEM(ca)
	}

	// Add client certificate & key
	if cert, err := tls.LoadX509KeyPair("~/tls-gen/basic/result/client_certificate.pem", "~/tls-gen/basic/result/client_key.pem"); err == nil {
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	// tls-gen ensures the Common Name (CN) is generated correctly
	conn, err := amqp.DialTLS("amqps://127.0.0.1/", cfg)

	// Connection will succeed, there are no errors
	log.Printf("conn: %v, err: %v", conn, err)
}

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

	// Buffer of 1 for our single outstanding publishing
	confirms := chd.NotifyPublish(make(chan amqp.Confirmation, 1))

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
			log.Fatalf("basic.publish destination: %+v", msg)
		}

		// only ack the source delivery when the destination acks the publishing
		if confirmed := <-confirms; confirmed.Ack {
			msg.Ack(false)
		} else {
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
		log.Fatalf("exchange.declare: %s", err)
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
			log.Fatalf("queue.declare: %v", err)
		}

		err = c.QueueBind(b.queue, b.key, "logs", false, nil)
		if err != nil {
			log.Fatalf("queue.bind: %v", err)
		}
	}

	// Set our quality of service.  Since we're sharing 3 consumers on the same
	// channel, we want at least 3 messages in flight.
	err = c.Qos(3, 0, false)
	if err != nil {
		log.Fatalf("basic.qos: %v", err)
	}

	// Establish our consumers that have different responsibilities.  Our first
	// two queues do not ack the messages on the server, so require to be acked
	// on the client.

	pages, err := c.Consume("page", "pager", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("basic.consume: %v", err)
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
		log.Fatalf("basic.consume: %v", err)
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
		log.Fatalf("basic.consume: %v", err)
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
		log.Fatalf("basic.cancel: %v", err)
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
		log.Fatalf("exchange.declare: %v", err)
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
		log.Fatalf("basic.publish: %v", err)
	}
}

func publishAllTheThings(conn *amqp.Connection) {
	// ... snarf snarf, barf barf
}

func ExampleConnection_NotifyBlocked() {
	// Simply logs when the server throttles the TCP connection for publishers

	// Test this by tuning your server to have a low memory watermark:
	// rabbitmqctl set_vm_memory_high_watermark 0.00000001

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}
	defer conn.Close()

	blockings := conn.NotifyBlocked(make(chan amqp.Blocking))
	go func() {
		for b := range blockings {
			if b.Active {
				log.Printf("TCP blocked: %q", b.Reason)
			} else {
				log.Printf("TCP unblocked")
			}
		}
	}()

	// Your application domain channel setup publishings
	publishAllTheThings(conn)
}
