package amqp

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"reflect"
	"testing"
	"testing/quick"
	"time"
	"os/signal"
	"syscall"
)

func init() {
	c := make(chan os.Signal)

	go func() {
		for {
			if _, ok := <-c; !ok {
				return
			}
			panic("ohai")
		}
	}()

	signal.Notify(c, syscall.SIGINFO)
}

func parseUrl(amqp string) (hostport string, username string, password string, vhost string) {
	u, err := url.Parse(amqp)
	if err != nil {
		return
	}

	host, port := "localhost", 5672
	username, password, vhost = "guest", "guest", "/"

	fmt.Sscanf(u.Host, "%s:%d", &host, &port)

	hostport = fmt.Sprintf("%s:%d", host, port)

	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	if u.Path != "" {
		vhost = u.Path
	}

	return
}

// Returns a conneciton to the AMQP if the AMQP_URL environment
// variable is set and a connnection can be established.
func integrationConnection(t *testing.T, name string) *Connection {
	u := os.Getenv("AMQP_URL")
	if u == "" {
		t.Log("Skipping integration tests, AMQP_URL not found in the environment")
		return nil
	}

	hostport, username, password, vhost := parseUrl(u)

	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		t.Error("Failed to connect to integration server:", err)
		return nil
	}

	//io := conn
	io := &logIO{t, name, conn}

	c, err := NewConnection(io, &PlainAuth{username, password}, vhost)

	if err != nil {
		t.Error("Failed to create client against integration server:", err)
		return nil
	}

	return c
}

func assertMessageBody(t *testing.T, msg Delivery, body []byte) bool {
	if bytes.Compare(msg.Body, body) != 0 {
		t.Errorf("Message body does not match have: %v expect %v", msg.Body, body)
		return false
	}
	return true
}

func TestIntegrationConnect(t *testing.T) {
	if c := integrationConnection(t, "connect"); c != nil {
		t.Log("have client")
	}
}

func TestIntegrationConnectChannel(t *testing.T) {
	if c := integrationConnection(t, "channel"); c != nil {
		if _, err := c.Channel(); err != nil {
			t.Errorf("Channel could not be open", err)
		}
	}
}

func TestIntegrationPublishConsume(t *testing.T) {
	queue := "test.integration.publish.consume"

	c1 := integrationConnection(t, "pub")
	c2 := integrationConnection(t, "sub")

	if c1 != nil && c2 != nil {
		pub, _ := c1.Channel()
		sub, _ := c2.Channel()

		pub.Q(queue).Declare(UntilUnused, false, false, nil)
		sub.Q(queue).Declare(UntilUnused, false, false, nil)

		messages, _ := sub.Q(queue).Consume(false, false, false, false, "", nil, nil)

		pub.E("").Publish(queue, false, false, Publishing{Body: []byte("pub 1")})
		pub.E("").Publish(queue, false, false, Publishing{Body: []byte("pub 2")})
		pub.E("").Publish(queue, false, false, Publishing{Body: []byte("pub 3")})

		assertMessageBody(t, <-messages, []byte("pub 1"))
		assertMessageBody(t, <-messages, []byte("pub 2"))
		assertMessageBody(t, <-messages, []byte("pub 3"))
	}
}

func (c *Connection) Generate(r *rand.Rand, _ int) reflect.Value {
	//fmt.Println("gen gen gen")
	hostport, username, password, vhost := parseUrl(os.Getenv("AMQP_URL"))

	conn, err := net.Dial("tcp", hostport)
	if err != nil {
		return reflect.ValueOf(nil)
	}

	c, err = NewConnection(conn, &PlainAuth{username, password}, vhost)
	if err != nil {
		return reflect.ValueOf(nil)
	}

	return reflect.ValueOf(c)
}

func (c Publishing) Generate(r *rand.Rand, _ int) reflect.Value {
	var ok bool
	var t reflect.Value

	p := Publishing{}
	//p.DeliveryMode = uint8(r.Intn(3))
	//p.Priority = uint8(r.Intn(8))

	if r.Intn(2) > 0 {
		p.ContentType = "application/octet-stream"
	}

	if r.Intn(2) > 0 {
		p.ContentEncoding = "gzip"
	}

	if r.Intn(2) > 0 {
		p.CorrelationId = fmt.Sprintf("%d", r.Int())
	}

	if r.Intn(2) > 0 {
		p.ReplyTo = fmt.Sprintf("%d", r.Int())
	}

	if r.Intn(2) > 0 {
		p.MessageId = fmt.Sprintf("%d", r.Int())
	}

	if r.Intn(2) > 0 {
		p.Type = fmt.Sprintf("%d", r.Int())
	}

	if r.Intn(2) > 0 {
		p.AppId = fmt.Sprintf("%d", r.Int())
	}

	if r.Intn(2) > 0 {
		p.Timestamp = time.Unix(r.Int63(), r.Int63())
	}

	if t, ok = quick.Value(reflect.TypeOf(p.Body), r); ok {
		p.Body = t.Bytes()
	}

	return reflect.ValueOf(p)
}

func TestQuickPublishOnly(t *testing.T) {
	if c := integrationConnection(t, "quick"); c != nil {
		defer c.Close()
		pub, err := c.Channel()
		q := pub.Q("test-publish")

		if _, err = q.Declare(UntilUnused, false, false, nil); err != nil {
			t.Error("Failed to declare", err)
			return
		}

		defer q.Delete(false, false, false)

		quick.Check(func(msg Publishing) bool {
			return pub.Q("test-publish").Publish(false, false, msg) == nil
		}, nil)
	}
}

func TestPublishEmptyBody(t *testing.T) {
	c1 := integrationConnection(t, "empty")
	if c1 != nil {
		defer c1.Close()

		pub, err := c1.Channel()
		if err != nil {
			t.Error("Failed to create channel")
			return
		}

		p := pub.Q("test-TestPublishEmptyBody")
		p.Declare(UntilUnused, false, false, nil)

		ch, err := p.Consume(false, false, false, false, "", nil, nil)

		err = p.Publish(false, false, Publishing{})

		if err != nil {
			t.Error("Failed to publish")
			return
		}
		if len((<-ch).Body) != 0 {
			t.Error("Received non empty body")
			return
		}
	}
}

func TestQuickPublishConsumeOnly(t *testing.T) {
	c1 := integrationConnection(t, "quick-pub")
	c2 := integrationConnection(t, "quick-sub")

	if c1 != nil && c2 != nil {
		defer c1.Close()
		defer c2.Close()

		pub, err := c1.Channel()
		sub, err := c2.Channel()

		p := pub.Q("TestPublishConsumeOnly")
		if _, err = p.Declare(UntilUnused, false, false, nil); err != nil {
			t.Error("Failed to declare", err)
			return
		}

		s := sub.Q("TestPublishConsumeOnly")
		if _, err = s.Declare(UntilUnused, false, false, nil); err != nil {
			t.Error("Failed to declare", err)
			return
		}

		defer s.Delete(false, false, false)

		ch, err := s.Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Error("Could not sub", err)
		}

		quick.CheckEqual(
			func(msg Publishing) []byte {
				//fmt.Println("check pub", msg)
				empty := Publishing{Body: msg.Body}
				if p.Publish(false, false, empty) != nil {
					return []byte{'X'}
				}
				//fmt.Println("published", empty)
				return msg.Body
			},
			func(msg Publishing) []byte {
				out := <-ch
				//fmt.Println("delivered", out)
				out.Ack(false)
				return out.Body
			},
			nil)
	}
}

func TestQuickPublishConsumeBigBody(t *testing.T) {
	c1 := integrationConnection(t, "big-pub")
	c2 := integrationConnection(t, "big-sub")

	if c1 != nil && c2 != nil {
		defer c1.Close()
		defer c2.Close()

		pub, err := c1.Channel()
		sub, err := c2.Channel()

		q := pub.Q("test-pubsub")
		if _, err = q.Declare(UntilUnused, false, false, nil); err != nil {
			t.Error("Failed to declare", err)
			return
		}

		ch, err := sub.Q("test-pubsub").Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Error("Could not sub", err)
		}

		msg := Publishing{
			Body: make([]byte, 1e6+10000),
		}

		err = pub.Q("test-pubsub").Publish(false, false, msg)
		if err != nil {
			t.Error("Could not publish big body")
		}

		if bytes.Compare((<-ch).Body, msg.Body) != 0 {
			t.Error("Consumed big body didn't match")
		}
	}
}
