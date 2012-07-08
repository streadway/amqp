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

func TestIntegrationConnect(t *testing.T) {
	if c := integrationConnection(t, "connect"); c != nil {
		t.Logf("have connection")
	}
}

func TestIntegrationConnectClose(t *testing.T) {
	if c := integrationConnection(t, "close"); c != nil {
		t.Logf("calling connection close")
		if err := c.Close(); err != nil {
			t.Fatalf("connection close: %s", err)
		}
		t.Logf("connection close OK")
	}
}

func TestIntegrationConnectChannel(t *testing.T) {
	if c := integrationConnection(t, "channel"); c != nil {
		if _, err := c.Channel(); err != nil {
			t.Errorf("Channel could not be opened: %s", err)
		}
	}
}

func TestIntegrationExchange(t *testing.T) {
	c1 := integrationConnection(t, "exch")
	if c1 != nil {
		defer c1.Close()

		channel, err := c1.Channel()
		if err != nil {
			t.Fatalf("create channel: %s", err)
		}
		t.Logf("create channel OK")

		exchange := channel.E("test-basic-ops-exchange")
		if err := exchange.Declare(
			UntilUnused, // lifetime
			"direct",    // type
			false,       // internal
			false,       // nowait
			nil,         // args
		); err != nil {
			t.Fatalf("declare exchange: %s", err)
		}
		t.Logf("declare exchange OK")

		// I'm not sure if this behavior is actually well-defined.
		// if err := exchange.Declare(
		// 	UntilUnused, // lifetime
		// 	"direct",    // type
		// 	false,       // internal
		// 	false,       // nowait
		// 	nil,         // args
		// ); err == nil {
		// 	t.Fatalf("re-declare same exchange didn't fail (it should have!)")
		// } else {
		// 	t.Logf("re-declare same exchange: got expected error: %s", err)
		// }

		if err := exchange.Delete(false, false); err != nil {
			t.Fatalf("delete exchange: %s", err)
		}
		t.Logf("delete exchange OK")

		if err := channel.Close(); err != nil {
			t.Fatalf("close channel: %s", err)
		}
		t.Logf("close channel OK")
	}
}

func TestIntegrationBasicQueueOperations(t *testing.T) {
	c1 := integrationConnection(t, "queue")
	if c1 != nil {
		defer c1.Close()

		channel, err := c1.Channel()
		if err != nil {
			t.Fatalf("create channel: %s", err)
		}
		t.Logf("create channel OK")

		exchangeName := "test-basic-ops-exchange"
		queueName := "test-basic-ops-queue"

		deleteQueueFirstOptions := []bool{true, false}
		for _, deleteQueueFirst := range deleteQueueFirstOptions {

			exchange := channel.E(exchangeName)
			if err := exchange.Declare(
				UntilDeleted, // lifetime (note: not UntilUnused)
				"direct",     // type
				false,        // internal
				false,        // nowait
				nil,          // args
			); err != nil {
				t.Fatalf("declare exchange: %s", err)
			}
			t.Logf("declare exchange OK")

			queue := channel.Q(queueName)
			if queueState, err := queue.Declare(
				UntilDeleted, // lifetime (note: not UntilUnused)
				false,        // exclusive
				false,        // noWait
				nil,          // arguments
			); err != nil {
				t.Fatalf("queue declare: %s", err)
			} else if !queueState.Declared {
				t.Fatalf("queue declare: state indicates not-declared")
			}
			t.Logf("declare queue OK")

			if err := queue.Bind(
				"",           // routingKey
				exchangeName, // sourceExchange
				false,        // noWait
				nil,          // arguments
			); err != nil {
				t.Fatalf("queue bind: %s", err)
			}
			t.Logf("queue bind OK")

			if deleteQueueFirst {

				if err := queue.Delete(
					false, // ifUnused (false=be aggressive)
					false, // ifEmpty (false=be aggressive)
					false, // noWait
				); err != nil {
					t.Fatalf("delete queue (first): %s", err)
				}
				t.Logf("delete queue (first) OK")

				if err := exchange.Delete(false, false); err != nil {
					t.Fatalf("delete exchange (after delete queue): %s", err)
				}
				t.Logf("delete exchange (after delete queue) OK")

			} else { // deleteExchangeFirst

				if err := exchange.Delete(false, false); err != nil {
					t.Fatalf("delete exchange (first): %s", err)
				}
				t.Logf("delete exchange (first) OK")

				if queueState, err := queue.Inspect(); err != nil {
					t.Fatalf("inspect queue state after deleting exchange: %s", err)
				} else if !queueState.Declared {
					t.Fatalf("after deleting exchange, queue disappeared")
				}
				t.Logf("queue properly remains after exchange is deleted")

				if err := queue.Delete(
					false, // ifUnused
					false, // ifEmpty
					false, // noWait
				); err != nil {
					t.Fatalf("delete queue (after delete exchange): %s", err)
				}
				t.Logf("delete queue (after delete exchange) OK")

			}
		}

		if err := channel.Close(); err != nil {
			t.Fatalf("close channel: %s", err)
		}
		t.Logf("close channel OK")
	}
}

func TestIntegrationChannelClosing(t *testing.T) {
	c1 := integrationConnection(t, "closings")
	if c1 != nil {
		defer c1.Close()

		// This function is run on every channel after it is successfully
		// opened. It can do something to verify something. It should be
		// quick; many channels may be opened!
		f := func(t *testing.T, c *Channel) {
			return
		}

		// open and close
		channel, err := c1.Channel()
		if err != nil {
			t.Fatalf("basic create channel: %s", err)
		}
		t.Logf("basic create channel OK")

		if err := channel.Close(); err != nil {
			t.Fatalf("basic close channel: %s", err)
		}
		t.Logf("basic close channel OK")

		// deferred close
		signal := make(chan bool)
		go func() {
			channel, err := c1.Channel()
			if err != nil {
				t.Fatalf("second create channel: %s", err)
			}
			t.Logf("second create channel OK")

			<-signal // a bit of synchronization
			f(t, channel)

			defer func() {
				if err := channel.Close(); err != nil {
					t.Fatalf("deferred close channel: %s", err)
				}
				t.Logf("deferred close channel OK")
				signal <- true
			}()
		}()
		signal <- true
		select {
		case <-signal:
			t.Logf("(got close signal OK)")
			break
		case <-time.After(250 * time.Millisecond):
			t.Fatalf("deferred close: timeout")
		}

		// multiple channels
		for _, n := range []int{2, 4, 8, 16, 32, 64, 128, 256} {
			channels := make([]*Channel, n)
			for i := 0; i < n; i++ {
				var err error
				if channels[i], err = c1.Channel(); err != nil {
					t.Fatalf("create channel %d/%d: %s", i+1, n, err)
				}
			}
			f(t, channel)
			for i, channel := range channels {
				if err := channel.Close(); err != nil {
					t.Fatalf("close channel %d/%d: %s", i+1, n, err)
				}
			}
			t.Logf("created/closed %d channels OK", n)
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
		defer c1.Close()

		ch, err := c1.Channel()
		if err != nil {
			t.Fatalf("Could not create channel")
		}

		queue := ch.Q("test.integration.blocking.close")

		_, err = queue.Declare(UntilUnused, false, false, nil)
		if err != nil {
			t.Fatalf("Could not declare")
		}

		msgs, err := queue.Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Fatalf("Could not consume")
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
				t.Fatalf("Close produced an error when it shouldn't")
			}
			succeed <- true
		}()

		select {
		case <-succeed:
			break
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
		defer c1.Close()
		defer c2.Close()

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
			t.Errorf("Failed to declare: %s", err)
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
			t.Errorf("Failed to create channel")
			return
		}

		p := pub.Q("test-TestPublishEmptyBody")
		p.Declare(UntilUnused, false, false, nil)

		ch, err := p.Consume(false, false, false, false, "", nil, nil)

		err = p.Publish(false, false, Publishing{})

		if err != nil {
			t.Errorf("Failed to publish")
			return
		}
		if len((<-ch).Body) != 0 {
			t.Errorf("Received non empty body")
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
			t.Errorf("Failed to declare: %s", err)
			return
		}

		s := sub.Q("TestPublishConsumeOnly")
		if _, err = s.Declare(UntilUnused, false, false, nil); err != nil {
			t.Errorf("Failed to declare: %s", err)
			return
		}

		defer s.Delete(false, false, false)

		ch, err := s.Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Errorf("Could not sub: %s", err)
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
			t.Errorf("Failed to declare: %s", err)
			return
		}

		ch, err := sub.Q("test-pubsub").Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Errorf("Could not sub: %s", err)
		}

		msg := Publishing{
			Body: make([]byte, 1e4+1000),
		}

		err = pub.Q("test-pubsub").Publish(false, false, msg)
		if err != nil {
			t.Errorf("Could not publish big body")
		}

		if bytes.Compare((<-ch).Body, msg.Body) != 0 {
			t.Errorf("Consumed big body didn't match")
		}
	}
}

// https://github.com/streadway/amqp/issues/7
func TestCorruptedMessageRegression(t *testing.T) {
	messageCount := 1024

	c1 := integrationConnection(t, "corrupt-pub")
	c2 := integrationConnection(t, "corrupt-sub")

	if c1 != nil && c2 != nil {
		defer c1.Close()
		defer c2.Close()

		ch1, err := c1.Channel()
		if err != nil {
			t.Fatalf("Cannot create Channel")
		}

		ch2, err := c2.Channel()
		if err != nil {
			t.Fatalf("Cannot create Channel")
		}

		queue := "test-corrupted-message-regression"

		pub := ch1.Q(queue)
		if _, err := pub.Declare(UntilUnused, false, false, nil); err != nil {
			t.Fatalf("Cannot declare")
		}

		sub := ch2.Q(queue)
		if _, err := pub.Declare(UntilUnused, false, false, nil); err != nil {
			t.Fatalf("Cannot declare")
		}

		msgs, err := sub.Consume(false, false, false, false, "", nil, nil)
		if err != nil {
			t.Fatalf("Cannot consume")
		}

		for i := 0; i < messageCount; i++ {
			err := pub.Publish(false, false, Publishing{
				Body: generateCrc32Random(7 * i),
			})

			if err != nil {
				t.Fatalf("Failed to publish")
			}
		}

		for i := 0; i < messageCount; i++ {
			select {
			case msg := <-msgs:
				assertMessageCrc32(t, msg.Body, fmt.Sprintf("missed match at %d", i))
			case <-time.After(1 * time.Second):
				t.Fatalf("Timed out after 1s")
			}
		}
	}
}

func TestExchangeDeclarePrecondition(t *testing.T) {
	c1 := integrationConnection(t, "exchange-double-declare")
	if c1 != nil {
		// defer c1.Close() // TODO stalls; see TODO below

		ch, err := c1.Channel()
		if err != nil {
			t.Fatalf("Create channel")
		}

		e := ch.E("test-mismatched-redeclare")

		err = e.Declare(
			UntilUnused, // lifetime (auto-delete)
			"direct",    // exchangeType
			false,       // internal
			false,       // noWait
			nil,         // arguments
		)
		if err != nil {
			t.Fatalf("Could not initially declare exchange")
		}

		err = e.Declare(
			UntilDeleted, // lifetime
			"direct",
			false,
			false,
			nil,
		)
		if err == nil {
			t.Fatalf("Expected to fail a redeclare with different lifetime, didn't receive an error")
		}
		t.Logf("good: got error: %s", err)

		// TODO stalls
		// (I'm guessing the connection becomes invalid after the redeclare.)
		// e.Delete(false, false)
	}
}
