package amqp

type Authentication interface {
}

type PlainAuth struct {
	Username string
	Password string
}
