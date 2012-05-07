package amqp

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
)

// Returns a conneciton to the AMQP if the AMQP_URL environment
// variable is set and a connnection can be established.
func integrationConnection(t *testing.T, name string) *Connection {
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

	c, err := NewConnection(&logIO{t, name, conn}, &PlainAuth{username, password}, vhost)
	if err != nil {
		t.Error("Failed to create client against integration server:", err)
		return nil
	}

	return c
}

func assertMessageBody(t *testing.T, msg Delivery, body []byte) {
	if bytes.Compare(msg.Body, body) != 0 {
		t.Errorf("Message body does not match have: %v expect %v", msg.Body, body)
	}
}

func TestIntegrationConnect(t *testing.T) {
	if c := integrationConnection(t, "connect"); c != nil {
		t.Log("have client")
	}
}

func TestIntegrationConnectChannel(t *testing.T) {
	c := integrationConnection(t, "channel")
	if c != nil {
		t.Log("have client")
	}

	if _, err := c.Channel(); err != nil {
		t.Errorf("Channel could not be open", err)
	}

}

func TestIntegrationPublishConsume(t *testing.T) {
	queue := "test.integration.publish.consume"

	pub, _ := integrationConnection(t, "pub").Channel()
	sub, _ := integrationConnection(t, "sub").Channel()

	if pub != nil && sub != nil {
		pub.Q(queue).Declare(UntilUnused, false, false, nil)
		sub.Q(queue).Declare(UntilUnused, false, false, nil)

		pub.E("").Publish(queue, false, false, Publishing{Body: []byte("1")})
		pub.E("").Publish(queue, false, false, Publishing{Body: []byte("2")})
		pub.E("").Publish(queue, false, false, Publishing{Body: []byte("3")})

		messages, _ := sub.Q(queue).Consume(false, false, false, false, "", nil, nil)

		assertMessageBody(t, <-messages, []byte("1"))
		assertMessageBody(t, <-messages, []byte("2"))
		assertMessageBody(t, <-messages, []byte("3"))
	}
}
