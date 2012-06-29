package amqp

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"testing"
	"testing/quick"
	"time"
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

func TestIntegrationConnect(t *testing.T) {
	if c := integrationConnection(t, "connect"); c != nil {
		t.Logf("have client")
	}
}

func TestIntegrationConnectChannel(t *testing.T) {
	if c := integrationConnection(t, "channel"); c != nil {
		if _, err := c.Channel(); err != nil {
			t.Errorf("Channel could not be opened: %s", err)
		}
	}
}

func TestIntegrationConnectBadVhost(t *testing.T) {
	urlStr := os.Getenv("AMQP_URL")
	if urlStr == "" {
		t.Logf("Skipping; AMQP_URL not found in the environment")
		return
	}

	uri, err := ParseURI(urlStr)
	if err != nil {
		t.Fatalf("Failed to parse URI: %s", err)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", uri.Host, uri.Port))
	if err != nil {
		t.Fatalf("Dial: %s", err)
	}

	vhost := "lolwat_not_found"
	_, err = NewConnection(&logIO{t, "badauth", conn}, uri.PlainAuth(), vhost)
	if err != ErrBadVhost {
		t.Errorf("Expected ErrBadVhost, got %s", err)
	}
}

// https://github.com/streadway/amqp/issues/6
func TestIntegrationNonBlockingClose(t *testing.T) {
	c1 := integrationConnection(t, "pub")
	if c1 != nil {
		ch, err := c1.Channel()
		if err != nil {
			t.Fatal("Could not create channel")
		}

		queue := ch.Q("test.integration.blocking.close")

		_, err = queue.Declare(UntilUnused, false, false, nil)
		if err != nil {
			t.Fatal("Could not declare")
		}

		msgs, err := queue.Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Fatal("Could not consume")
		}

		// Simulate a consumer
		go func() {
			for _ = range msgs {
				t.Logf("Oh my, received message on an empty queue")
			}
		}()

		time.Sleep(2 * time.Second)

		succeed := make(chan bool)
		fail := time.After(1 * time.Second)

		go func() {
			if err = ch.Close(); err != nil {
				t.Fatal("Close produced an error when it shouldn't")
			}
			succeed <- true
		}()

		select {
		case <-succeed:
		case <-fail:
			t.Fatalf("Close timed out after 1s")
		}
	}
}

func TestIntegrationConnectBadCredentials(t *testing.T) {
	urlStr := os.Getenv("AMQP_URL")
	if urlStr == "" {
		t.Logf("Skipping; AMQP_URL not found in the environment")
		return
	}

	uri, err := ParseURI(urlStr)
	if err != nil {
		t.Fatalf("Failed to parse URI: %s", err)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", uri.Host, uri.Port))
	if err != nil {
		t.Fatalf("Dial: %s", err)
	}

	if _, err = NewConnection(
		&logIO{t, "badauth", conn},
		&PlainAuth{
			Username: "",
			Password: "",
		},
		uri.Vhost,
	); err != ErrBadCredentials {
		t.Errorf("Expected ErrBadCredentials, got %s", err)
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
		defer pub.Q(queue).Delete(false, false, false)

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
	urlStr := os.Getenv("AMQP_URL")
	if urlStr == "" {
		return reflect.ValueOf(nil)
	}

	uri, err := ParseURI(urlStr)
	if err != nil {
		return reflect.ValueOf(nil)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", uri.Host, uri.Port))
	if err != nil {
		return reflect.ValueOf(nil)
	}

	c, err = NewConnection(conn, uri.PlainAuth(), uri.Vhost)
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
				empty := Publishing{Body: msg.Body}
				if p.Publish(false, false, empty) != nil {
					return []byte{'X'}
				}
				return msg.Body
			},
			func(msg Publishing) []byte {
				out := <-ch
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
			Body: make([]byte, 1e6+1000),
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

// https://github.com/streadway/amqp/issues/7
func TestCorruptedMessageRegression(t *testing.T) {
	messageCount := 1024

	c1 := integrationConnection(t, "corrupt-pub")
	c2 := integrationConnection(t, "corrupt-sub")

	if c1 != nil && c2 != nil {
		//defer c1.Close()
		//defer c2.Close()

		ch1, err := c1.Channel()
		if err != nil {
			t.Fatal("Cannot create Channel")
		}

		ch2, err := c2.Channel()
		if err != nil {
			t.Fatal("Cannot create Channel")
		}

		queue := "test-corrupted-message-regression"

		pub := ch1.Q(queue)
		if _, err := pub.Declare(UntilUnused, false, false, nil); err != nil {
			t.Fatal("Cannot declare")
		}

		sub := ch2.Q(queue)
		if _, err := pub.Declare(UntilUnused, false, false, nil); err != nil {
			t.Fatal("Cannot declare")
		}

		msgs, err := sub.Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Fatal("Cannot consume")
		}

		for i := 0; i < messageCount; i++ {
			err := pub.Publish(false, false, Publishing{
				Body: generateCrc32Random(7 * i),
			})

			if err != nil {
				t.Fatal("Failed to publish")
			}
		}

		for i := 0; i < messageCount; i++ {
			select {
			case msg := <-msgs:
				assertMessageCrc32(t, msg.Body, fmt.Sprintf("missed match at %d", i))
			case <-time.After(1 * time.Second):
				t.Fatal("Timed out after 1s")
			}
		}
	}
}

func TestExchangeDeclarePrecondition(t *testing.T) {
	c1 := integrationConnection(t, "exchange-double-declare")
	if c1 != nil {
		//defer c1.Close()

		ch, err := c1.Channel()
		if err != nil {
			t.Fatal("Create channel")
		}

		e := ch.E("test-mismatched-redeclare")

		err = e.Declare(UntilUnused, "direct", false, false, nil)
		if err != nil {
			t.Fatal("Could not initially declare exchange")
		}
		// TODO currently stalls
		// defer e.Delete(false, false)

		err = e.Declare(UntilDeleted, "direct", false, false, nil)
		if err == nil {
			t.Fatal("Expected to fail a redeclare with different lifetime, didn't receive an error")
		}
	}
}
