package amqp_test

import (
	"amqp"
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
)

// Returns a conneciton to the AMQP if the AMQP_URL environment
// variable is set and a connnection can be established.
func integrationConnection(t *testing.T) *amqp.Connection {
	u, err := url.Parse(os.Getenv("AMQP_URL"))
	if err != nil {
		t.Log("Skipping integration tests, AMQP_URL not found in the environment")
		return nil
	}

	host, port, username, password, vhost := "localhost", 5672, "guest", "guest", "/"
	fmt.Scanf("%s:%d", &host, &port)

	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	if u.Path != "" {
		vhost = u.Path
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		t.Error("Failed to connect to integration server:", err)
		return nil
	}

	c, err := amqp.NewConnection(&logIO{t, "integration", conn}, &amqp.PlainAuth{username, password}, vhost)
	if err != nil {
		t.Error("Failed to create client against integration server:", err)
		return nil
	}

	return c
}

func assertMessageBody(t *testing.T, msg *amqp.Delivery, body []byte) {
	if bytes.Compare(msg.Body, body) != 0 {
		t.Errorf("Message body does not match have: %v expect %v", msg.Body, body)
	}
}

func TestIntegrationConnect(t *testing.T) {
	if c := integrationConnection(t); c != nil {
		t.Log("have client")
	}
}

func TestIntegrationPublishConsume(t *testing.T) {
	queue := "test.integration.publish.consume"

	pub, _ := integrationConnection(t).OpenChannel()
	sub, _ := integrationConnection(t).OpenChannel()

	if pub != nil && sub != nil {
		sub.QueueDeclare(queue, false, true, false, nil)

		pub.BasicPublish("", queue, false, false, []byte("1"), amqp.ContentProperties{})
		pub.BasicPublish("", queue, false, false, []byte("2"), amqp.ContentProperties{})
		pub.BasicPublish("", queue, false, false, []byte("3"), amqp.ContentProperties{})

		messages, _ := sub.BasicConsume(queue, "", false, false, false, nil)

		assertMessageBody(t, <-messages, []byte("1"))
		assertMessageBody(t, <-messages, []byte("2"))
		assertMessageBody(t, <-messages, []byte("3"))
	}
}
