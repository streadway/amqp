package amqp

import (
	"amqp/wire"
	"errors"
)

var (
	ErrBadProtocol = errors.New("Unexpected protocol message")
)

type Table wire.Table

type Timestamp wire.Timestamp
