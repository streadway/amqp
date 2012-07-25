// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"time"
)

// A delivery from the server to a consumer created with Queue.Consume or Queue.Get.  The
// types have been promoted from the framing and the method so to provide flat,
// typed access to everything the channel knows about this message.
type Delivery struct {
	channel *Channel

	Headers Table // Application or header exchange table

	// Properties for the message
	ContentType     string    // MIME content type
	ContentEncoding string    // MIME content encoding
	DeliveryMode    uint8     // queue implemention use - non-persistent (1) or persistent (2)
	Priority        uint8     // queue implementation use - 0 to 9
	CorrelationId   string    // application use - correlation identifier
	ReplyTo         string    // application use - address to to reply to (ex: RPC)
	Expiration      string    // implementation use - message expiration spec
	MessageId       string    // application use - message identifier
	Timestamp       time.Time // application use - message timestamp
	Type            string    // application use - message type name
	UserId          string    // application use - creating user idA - should be authenticated user
	AppId           string    // application use - creating application id

	// only meaningful from a Channel.Consume or Queue.Consume
	ConsumerTag string

	// only meaningful on Channel.Get
	MessageCount uint32

	// Only meaningful when this is a result of a message consumption
	DeliveryTag uint64
	Redelivered bool
	Exchange    string
	RoutingKey  string // Message routing key

	Body []byte
}

func newDelivery(channel *Channel, msg messageWithContent) *Delivery {
	props, body := msg.getContent()

	delivery := Delivery{
		channel: channel,

		Headers:         props.Headers,
		ContentType:     props.ContentType,
		ContentEncoding: props.ContentEncoding,
		DeliveryMode:    props.DeliveryMode,
		Priority:        props.Priority,
		CorrelationId:   props.CorrelationId,
		ReplyTo:         props.ReplyTo,
		Expiration:      props.Expiration,
		MessageId:       props.MessageId,
		Timestamp:       props.Timestamp,
		Type:            props.Type,
		UserId:          props.UserId,
		AppId:           props.AppId,

		Body: body,
	}

	// Properties for the delivery types
	switch m := msg.(type) {
	case *basicDeliver:
		delivery.ConsumerTag = m.ConsumerTag
		delivery.DeliveryTag = m.DeliveryTag
		delivery.Redelivered = m.Redelivered
		delivery.Exchange = m.Exchange
		delivery.RoutingKey = m.RoutingKey

	case *basicGetOk:
		delivery.MessageCount = m.MessageCount
		delivery.DeliveryTag = m.DeliveryTag
		delivery.Redelivered = m.Redelivered
		delivery.Exchange = m.Exchange
		delivery.RoutingKey = m.RoutingKey
	}

	return &delivery
}

func (me *Delivery) Ack(multiple bool) error {
	return me.channel.send(me.channel, &basicAck{
		DeliveryTag: me.DeliveryTag,
		Multiple:    multiple,
	})
}

func (me *Delivery) Reject(requeue bool) error {
	return me.channel.send(me.channel, &basicReject{
		DeliveryTag: me.DeliveryTag,
		Requeue:     requeue,
	})
}

func (me *Delivery) Cancel(noWait bool) (consumerTag string, err error) {
	return me.ConsumerTag, me.channel.Cancel(me.ConsumerTag, noWait)
}

// RabbitMQ extension - Negatively acknowledge the delivery of message(s)
// identified by the deliveryTag.  When multiple, nack messages up to and
// including delivered messages up until the deliveryTag.
//
// This method must not be used to select or requeue messages the client wishes
// not to handle.
func (me *Delivery) Nack(multiple, requeue bool) error {
	return me.channel.send(me.channel, &basicNack{
		DeliveryTag: me.DeliveryTag,
		Multiple:    multiple,
		Requeue:     requeue,
	})
}
