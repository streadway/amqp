package main

import (
	"github.com/streadway/amqp"
	"log"
)

type publisher interface {
	Publish(msg *amqp.Publishing) error
}

type publisherPool struct {
	name       string
	publishers chan publisher
}

func newPublisherPool(name string, inflight int, publishers []publisher) *publisherPool {
	pool := &publisherPool{
		name:       name,
		publishers: make(chan publisher, inflight),
	}

	for _, publisher := range publishers {
		pool.publishers <- publisher
	}

	return pool
}

func (p *publisherPool) Publish(msg *amqp.Publishing) error {
	publisher := <-p.publishers
	err := publisher.Publish(msg)
	p.publishers <- publisher

	return err
}

type examplePublisher struct {
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

func newExamplePublisher(conn *amqp.Connection, exchange string, routingKey string) *examplePublisher {
	ch, _ := conn.Channel()
	p := &examplePublisher{
		channel:    ch,
		exchange:   exchange,
		routingKey: routingKey,
	}

	return p
}

func (p *examplePublisher) Publish(msg *amqp.Publishing) error {
	return p.channel.Publish(p.exchange, p.routingKey, false, false, *msg)
}

func main() {
	conn, err := amqp.Dial("amqp://user:password@host:port//vhost")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	publishers := make([]publisher, 10)
	inflight := 10
	for i := 0; i < inflight; i++ {
		publishers[i] = newExamplePublisher(conn, "example-exchange", "example-routing-key")
	}

	pool := newPublisherPool("example", inflight, publishers)

	for i := 0; i < 100; i ++ {
		err = pool.Publish(&amqp.Publishing{Body: []byte(`hello`)})
		if err != nil {
			log.Fatal(err)
		}
	}
}
