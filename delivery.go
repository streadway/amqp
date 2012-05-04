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
	if err = me.channel.send(&basicCancel{
		ConsumerTag: me.ConsumerTag,
		NoWait:      noWait,
	}); err != nil {
		return
	}

	if !noWait {
		switch ok := (<-me.channel.rpc).(type) {
		case *basicCancelOk:
			return ok.ConsumerTag, nil
		case nil:
			return "", me.channel.Close()
		default:
			return "", ErrBadProtocol
		}
	}

	return
}
