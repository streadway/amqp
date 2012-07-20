// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

func (me *Delivery) Ack(multiple bool) error {
	return me.channel.send(&basicAck{
		DeliveryTag: me.DeliveryTag,
		Multiple:    multiple,
	})
}

func (me *Delivery) Reject(requeue bool) error {
	return me.channel.send(&basicReject{
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
	return me.channel.send(&basicNack{
		DeliveryTag: me.DeliveryTag,
		Multiple:    multiple,
		Requeue:     requeue,
	})
}
