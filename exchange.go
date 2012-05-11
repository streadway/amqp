package amqp

func (me *Exchange) Declare(lifetime Lifetime, exchangeType string, internal bool, noWait bool, arguments Table) (err error) {
	if err = me.channel.send(&exchangeDeclare{
		Exchange:   me.name,
		Type:       exchangeType,
		Passive:    false,
		Durable:    lifetime.durable(),
		AutoDelete: lifetime.autoDelete(),
		Internal:   internal,
		NoWait:     noWait,
		Arguments:  arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.channel.rpc).(type) {
		case *exchangeDeclareOk:
			return
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Exchange) Delete(ifUnused bool, noWait bool) (err error) {
	if err = me.channel.send(&exchangeDelete{
		Exchange: me.name,
		IfUnused: ifUnused,
		NoWait:   noWait,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.channel.rpc).(type) {
		case *exchangeDeleteOk:
			return
		case nil:
			return me.channel.Close()
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Exchange) Bind(routingKey string, destinationExchange string, noWait bool, arguments Table) (err error) {
	if err = me.channel.send(&exchangeBind{
		Destination: destinationExchange,
		Source:      me.name,
		RoutingKey:  routingKey,
		NoWait:      noWait,
		Arguments:   arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.channel.rpc).(type) {
		case *exchangeBindOk:
			return
		case nil:
			return me.channel.Close()
		default:
			return ErrBadProtocol
		}
	}

	return
}

func (me *Exchange) Unbind(routingKey string, destinationExchange string, noWait bool, arguments Table) (err error) {
	if err = me.channel.send(&exchangeUnbind{
		Destination: destinationExchange,
		Source:      me.name,
		RoutingKey:  routingKey,
		NoWait:      noWait,
		Arguments:   arguments,
	}); err != nil {
		return
	}

	if !noWait {
		switch (<-me.channel.rpc).(type) {
		case *exchangeUnbindOk:
			return
		case nil:
			return me.channel.Close()
		default:
			return ErrBadProtocol
		}
	}
	return
}

// Publishes a message to an exchange.
func (me *Exchange) Publish(routingKey string, mandatory bool, immediate bool, msg Publishing) (err error) {
	return me.channel.send(&basicPublish{
		Exchange:   me.name,
		RoutingKey: routingKey,
		Mandatory:  mandatory,
		Immediate:  immediate,
		Body:       msg.Body,
		Properties: properties{
			Headers:         msg.Headers,
			ContentType:     msg.ContentType,
			ContentEncoding: msg.ContentEncoding,
			DeliveryMode:    msg.DeliveryMode,
			Priority:        msg.Priority,
			CorrelationId:   msg.CorrelationId,
			ReplyTo:         msg.ReplyTo,
			Expiration:      msg.Expiration,
			MessageId:       msg.MessageId,
			Timestamp:       msg.Timestamp,
			Type:            msg.Type,
			UserId:          msg.UserId,
			AppId:           msg.AppId,
		},
	})
}
