// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
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

		exchange := "test-basic-ops-exchange"

		if err := channel.ExchangeDeclare(
			exchange,    // name
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
		//      UntilUnused, // lifetime
		//      "direct",    // type
		//      false,       // internal
		//      false,       // nowait
		//      nil,         // args
		// ); err == nil {
		//      t.Fatalf("re-declare same exchange didn't fail (it should have!)")
		// } else {
		//      t.Logf("re-declare same exchange: got expected error: %s", err)
		// }

		if err := channel.ExchangeDelete(exchange, false, false); err != nil {
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

			if err := channel.ExchangeDeclare(
				exchangeName, // name
				UntilDeleted, // lifetime (note: not UntilUnused)
				"direct",     // type
				false,        // internal
				false,        // nowait
				nil,          // args
			); err != nil {
				t.Fatalf("declare exchange: %s", err)
			}
			t.Logf("declare exchange OK")

			if queueState, err := channel.QueueDeclare(
				queueName,    // name
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

			if err := channel.QueueBind(
				queueName,    // name
				"",           // routingKey
				exchangeName, // sourceExchange
				false,        // noWait
				nil,          // arguments
			); err != nil {
				t.Fatalf("queue bind: %s", err)
			}
			t.Logf("queue bind OK")

			if deleteQueueFirst {
				if err := channel.QueueDelete(
					queueName, // name
					false,     // ifUnused (false=be aggressive)
					false,     // ifEmpty (false=be aggressive)
					false,     // noWait
				); err != nil {
					t.Fatalf("delete queue (first): %s", err)
				}
				t.Logf("delete queue (first) OK")

				if err := channel.ExchangeDelete(exchangeName, false, false); err != nil {
					t.Fatalf("delete exchange (after delete queue): %s", err)
				}
				t.Logf("delete exchange (after delete queue) OK")

			} else { // deleteExchangeFirst
				if err := channel.ExchangeDelete(exchangeName, false, false); err != nil {
					t.Fatalf("delete exchange (first): %s", err)
				}
				t.Logf("delete exchange (first) OK")

				if queueState, err := channel.QueueInspect(queueName); err != nil {
					t.Fatalf("inspect queue state after deleting exchange: %s", err)
				} else if !queueState.Declared {
					t.Fatalf("after deleting exchange, queue disappeared")
				}
				t.Logf("queue properly remains after exchange is deleted")

				if err := channel.QueueDelete(
					queueName,
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

func TestIntegrationConnectBadCredentials(t *testing.T) {
	if uri, ok := integrationUri(t); ok {
		uri.Username = "lolwho"

		if _, err := Dial(uri.String()); err != ErrBadCredentials {
			t.Errorf("Expected ErrBadCredentials, got %s", err)
		}
	}
}

func TestIntegrationConnectBadVhost(t *testing.T) {
	if uri, ok := integrationUri(t); ok {
		uri.Vhost = "lolwat"

		if _, err := Dial(uri.String()); err != ErrBadVhost {
			t.Errorf("Expected ErrBadVhost, got %s", err)
		}
	}
}

func TestIntegrationConnectHostBrackets(t *testing.T) {
	if uri, ok := integrationUri(t); ok {
		uri.Host = "::1"

		_, err := Dial(uri.String())

		if e, ok := err.(*net.AddrError); ok {
			t.Errorf("Expected no AddrError on colon in hostname, got %v", e)
		}
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

		queue := "test.integration.blocking.close"

		_, err = ch.QueueDeclare(queue, UntilUnused, false, false, nil)
		if err != nil {
			t.Fatalf("Could not declare")
		}

		msgs, err := ch.Consume(queue, false, false, false, false, "", nil, nil)
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

func TestIntegrationPublishConsume(t *testing.T) {
	queue := "test.integration.publish.consume"

	c1 := integrationConnection(t, "pub")
	c2 := integrationConnection(t, "sub")

	if c1 != nil && c2 != nil {
		defer c1.Close()
		defer c2.Close()

		pub, _ := c1.Channel()
		sub, _ := c2.Channel()

		pub.QueueDeclare(queue, UntilUnused, false, false, nil)
		sub.QueueDeclare(queue, UntilUnused, false, false, nil)
		defer pub.QueueDelete(queue, false, false, false)

		messages, _ := sub.Consume(queue, false, false, false, false, "", nil, nil)

		pub.Publish("", queue, false, false, Publishing{Body: []byte("pub 1")})
		pub.Publish("", queue, false, false, Publishing{Body: []byte("pub 2")})
		pub.Publish("", queue, false, false, Publishing{Body: []byte("pub 3")})

		assertConsumeBody(t, messages, []byte("pub 1"))
		assertConsumeBody(t, messages, []byte("pub 2"))
		assertConsumeBody(t, messages, []byte("pub 3"))
	}
}

func TestIntegrationConsumeFlow(t *testing.T) {
	queue := "test.integration.consumer-flow"

	c1 := integrationConnection(t, "pub-flow")
	c2 := integrationConnection(t, "sub-flow")

	if c1 != nil && c2 != nil {
		defer c1.Close()
		defer c2.Close()

		pub, _ := c1.Channel()
		sub, _ := c2.Channel()

		pub.QueueDeclare(queue, UntilUnused, false, false, nil)
		sub.QueueDeclare(queue, UntilUnused, false, false, nil)
		defer pub.QueueDelete(queue, false, false, false)

		sub.Qos(1, 0, false)

		messages, _ := sub.Consume(queue, false, false, false, false, "", nil, nil)

		pub.Publish("", queue, false, false, Publishing{Body: []byte("pub 1")})
		pub.Publish("", queue, false, false, Publishing{Body: []byte("pub 2")})

		msg := assertConsumeBody(t, messages, []byte("pub 1"))

		sub.Flow(false)

		msg.Ack(false)

		select {
		case <-messages:
			t.Fatalf("message was delivered when flow was not active")
		default:
		}

		sub.Flow(true)

		msg = assertConsumeBody(t, messages, []byte("pub 2"))
		msg.Ack(false)
	}
}

// This test is driven by a private API to simulate the server sending a channelFlow message
func TestIntegrationPublishFlow(t *testing.T) {
	// TODO - no idea how to test without affecting the server or mucking internal APIs
	// i'd like to make sure the RW lock can be held by multiple publisher threads
	// and that multiple channelFlow messages do not block the dispatch thread
}

func TestIntegrationConsumeCancel(t *testing.T) {
	queue := "test.integration.consume-cancel"

	c := integrationConnection(t, "pub")

	if c != nil {
		defer c.Close()

		ch, _ := c.Channel()

		ch.QueueDeclare(queue, UntilUnused, false, false, nil)
		defer ch.QueueDelete(queue, false, false, false)

		messages, _ := ch.Consume(queue, false, false, false, false, "integration-tag", nil, nil)

		ch.Publish("", queue, false, false, Publishing{Body: []byte("1")})

		assertConsumeBody(t, messages, []byte("1"))

		err := ch.Cancel("integration-tag", false)
		if err != nil {
			t.Fatalf("error cancelling the consumer: %v", err)
		}

		ch.Publish("", queue, false, false, Publishing{Body: []byte("2")})

		select {
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Timeout on Close")
		case _, ok := <-messages:
			if ok {
				t.Fatalf("Extra message on consumer when consumer should have been closed")
			}
		}
	}
}

func (c *Connection) Generate(r *rand.Rand, _ int) reflect.Value {
	urlStr := os.Getenv("AMQP_URL")
	if urlStr == "" {
		return reflect.ValueOf(nil)
	}

	conn, err := Dial(urlStr)
	if err != nil {
		return reflect.ValueOf(nil)
	}

	return reflect.ValueOf(conn)
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
		queue := "test-publish"

		if _, err = pub.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Errorf("Failed to declare: %s", err)
			return
		}

		defer pub.QueueDelete(queue, false, false, false)

		quick.Check(func(msg Publishing) bool {
			return pub.Publish("", queue, false, false, msg) == nil
		}, nil)
	}
}

func TestPublishEmptyBody(t *testing.T) {
	c1 := integrationConnection(t, "empty")
	if c1 != nil {
		defer c1.Close()

		ch, err := c1.Channel()
		if err != nil {
			t.Errorf("Failed to create channel")
			return
		}

		queue := "test-TestPublishEmptyBody"

		if _, err := ch.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Fatalf("Could not declare")
		}

		messages, err := ch.Consume(queue, false, false, false, false, "", nil, nil)
		if err != nil {
			t.Fatalf("Could not consume")
		}

		err = ch.Publish("", queue, false, false, Publishing{})
		if err != nil {
			t.Fatalf("Could not publish")
		}

		select {
		case msg := <-messages:
			if len(msg.Body) != 0 {
				t.Errorf("Received non empty body")
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("Timeout on receive")
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

		queue := "TestPublishConsumeOnly"

		if _, err = pub.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Errorf("Failed to declare: %s", err)
			return
		}

		if _, err = sub.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Errorf("Failed to declare: %s", err)
			return
		}

		defer sub.QueueDelete(queue, false, false, false)

		ch, err := sub.Consume(queue, false, false, false, false, "", nil, nil)
		if err != nil {
			t.Errorf("Could not sub: %s", err)
		}

		quick.CheckEqual(
			func(msg Publishing) []byte {
				empty := Publishing{Body: msg.Body}
				if pub.Publish("", queue, false, false, empty) != nil {
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

		queue := "test-pubsub"

		if _, err = sub.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Errorf("Failed to declare: %s", err)
			return
		}

		ch, err := sub.Consume(queue, false, false, false, false, "", nil, nil)
		if err != nil {
			t.Errorf("Could not sub: %s", err)
		}

		fixture := Publishing{
			Body: make([]byte, 1e4+1000),
		}

		if _, err = pub.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Errorf("Failed to declare: %s", err)
			return
		}

		err = pub.Publish("", queue, false, false, fixture)
		if err != nil {
			t.Errorf("Could not publish big body")
		}

		select {
		case msg := <-ch:
			if bytes.Compare(msg.Body, fixture.Body) != 0 {
				t.Errorf("Consumed big body didn't match")
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("Timeout on receive")
		}
	}
}

func TestIntegrationGetOk(t *testing.T) {
	if c := integrationConnection(t, "getok"); c != nil {
		defer c.Close()

		queue := "test.get-ok"
		ch, _ := c.Channel()

		ch.QueueDeclare(queue, UntilUnused, false, false, nil)
		ch.Publish("", queue, false, false, Publishing{Body: []byte("ok")})

		msg, ok, err := ch.Get(queue, false)

		if err != nil {
			t.Fatalf("Failed get: %v", err)
		}

		if !ok {
			t.Fatalf("Get on a queued message did not find the message")
		}

		if string(msg.Body) != "ok" {
			t.Fatalf("Get did not get the correct message")
		}
	}
}

func TestIntegrationGetEmpty(t *testing.T) {
	if c := integrationConnection(t, "getok"); c != nil {
		defer c.Close()

		queue := "test.get-ok"
		ch, _ := c.Channel()

		ch.QueueDeclare(queue, UntilUnused, false, false, nil)

		_, ok, err := ch.Get(queue, false)

		if err != nil {
			t.Fatalf("Failed get: %v", err)
		}

		if !ok {
			t.Fatalf("Get on a queued message retrieved a message when it shouldn't have")
		}
	}
}

func TestIntegrationTxCommit(t *testing.T) {
	if c := integrationConnection(t, "txcommit"); c != nil {
		defer c.Close()

		queue := "test.tx.commit"
		ch, _ := c.Channel()

		ch.QueueDeclare(queue, UntilUnused, false, false, nil)

		if err := ch.TxSelect(); err != nil {
			t.Fatalf("tx.select failed")
		}

		ch.Publish("", queue, false, false, Publishing{Body: []byte("ok")})

		if err := ch.TxCommit(); err != nil {
			t.Fatalf("tx.commit failed")
		}

		msg, ok, err := ch.Get(queue, false)

		if err != nil || !ok {
			t.Fatalf("Failed get: %v", err)
		}

		if string(msg.Body) != "ok" {
			t.Fatalf("Get did not get the correct message from the transaction")
		}
	}
}

func TestIntegrationTxRollback(t *testing.T) {
	if c := integrationConnection(t, "txrollback"); c != nil {
		defer c.Close()

		queue := "test.tx.rollback"
		ch, _ := c.Channel()

		ch.QueueDeclare(queue, UntilUnused, false, false, nil)

		if err := ch.TxSelect(); err != nil {
			t.Fatalf("tx.select failed")
		}

		ch.Publish("", queue, false, false, Publishing{Body: []byte("ok")})

		if err := ch.TxRollback(); err != nil {
			t.Fatalf("tx.rollback failed")
		}

		_, ok, err := ch.Get(queue, false)

		if err != nil {
			t.Fatalf("Failed get: %v", err)
		}

		if ok {
			t.Fatalf("message was published when it should have been rolled back")
		}
	}
}

// https://github.com/streadway/amqp/issues/7
func TestCorruptedMessageRegression(t *testing.T) {
	messageCount := 1024

	c1 := integrationConnection(t, "")
	c2 := integrationConnection(t, "")

	if c1 != nil && c2 != nil {
		defer c1.Close()
		defer c2.Close()

		pub, err := c1.Channel()
		if err != nil {
			t.Fatalf("Cannot create Channel")
		}

		sub, err := c2.Channel()
		if err != nil {
			t.Fatalf("Cannot create Channel")
		}

		queue := "test-corrupted-message-regression"

		if _, err := pub.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Fatalf("Cannot declare")
		}

		if _, err := sub.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
			t.Fatalf("Cannot declare")
		}

		msgs, err := sub.Consume(queue, false, false, false, false, "", nil, nil)
		if err != nil {
			t.Fatalf("Cannot consume")
		}

		for i := 0; i < messageCount; i++ {
			err := pub.Publish("", queue, false, false, Publishing{
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
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("Timeout on recv")
			}
		}
	}
}

func TestExchangeDeclarePrecondition(t *testing.T) {
	c1 := integrationConnection(t, "exchange-double-declare")
	c2 := integrationConnection(t, "exchange-double-declare-cleanup")
	if c1 != nil {
		defer c2.Close()
		ch, err := c1.Channel()
		if err != nil {
			t.Fatalf("Create channel")
		}

		exchange := "test-mismatched-redeclare"

		err = ch.ExchangeDeclare(
			exchange,
			UntilUnused, // lifetime (auto-delete)
			"direct",    // exchangeType
			false,       // internal
			false,       // noWait
			nil,         // arguments
		)
		if err != nil {
			t.Fatalf("Could not initially declare exchange")
		}

		err = ch.ExchangeDeclare(
			exchange,
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

		// after the redeclaration above, channel exception is raised (406) and it is closed
		// so we use a different connection and channel to clean up.
		ch2, err := c2.Channel()
		if err != nil {
			t.Fatalf("Could not create channel")
		}

		if err = ch2.ExchangeDelete(exchange, false, false); err != nil {
			t.Fatalf("Could not delete exchange")
		}
	}
}

func TestRabbitMQQueueTTLGet(t *testing.T) {
	if c := integrationRabbitMQ(t, "ttl"); c != nil {
		defer c.Close()

		queue := "test.rabbitmq-message-ttl"
		channel, err := c.Channel()
		if err != nil {
			t.Fatalf("channel: %v", err)
		}

		if _, err = channel.QueueDeclare(
			queue,
			UntilUnused,
			false,
			false,
			Table{"x-message-ttl": int32(100)}, // in ms
		); err != nil {
			t.Fatalf("queue declare: %s", err)
		}

		channel.Publish("", queue, false, false, Publishing{Body: []byte("ttl")})

		time.Sleep(200 * time.Millisecond)

		_, ok, err := channel.Get(queue, false)

		if ok {
			t.Fatalf("Expected the message to expire in 100ms, it didn't expire after 200ms")
		}

		if err != nil {
			t.Fatalf("Failed to get on ttl queue")
		}
	}
}

func TestRabbitMQQueueNackMultipleRequeue(t *testing.T) {
	if c := integrationRabbitMQ(t, "nack"); c != nil {
		defer c.Close()

		if c.IsCapable("basic.nack") {
			queue := "test.rabbitmq-basic-nack"
			channel, err := c.Channel()
			if err != nil {
				t.Fatalf("channel: %v", err)
			}

			if _, err = channel.QueueDeclare(queue, UntilUnused, false, false, nil); err != nil {
				t.Fatalf("queue declare: %s", err)
			}

			channel.Publish("", queue, false, false, Publishing{Body: []byte("1")})
			channel.Publish("", queue, false, false, Publishing{Body: []byte("2")})

			m1, ok, err := channel.Get(queue, false)
			if !ok || err != nil || m1.Body[0] != '1' {
				t.Fatalf("could not get message", m1)
			}

			m2, ok, err := channel.Get(queue, false)
			if !ok || err != nil || m2.Body[0] != '2' {
				t.Fatalf("could not get message", m2)
			}

			m2.Nack(true, true)

			m1, ok, err = channel.Get(queue, false)
			if !ok || err != nil || m1.Body[0] != '1' {
				t.Fatalf("could not get message", m1)
			}

			m2, ok, err = channel.Get(queue, false)
			if !ok || err != nil || m2.Body[0] != '2' {
				t.Fatalf("could not get message", m2)
			}
		}
	}
}
