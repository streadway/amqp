package amqp

import (
	"crypto/tls"
	"time"
)

// Option callback for connection option
type Option func(*Config) error

// SetOptions set amqp connection options
func (a *Config) SetOptions(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return err
		}
	}

	return nil
}

// SetTLS specifies the client configuration of the TLS connection
// when establishing a tls transport.
// If the URL uses an amqps scheme, then an empty tls.Config with the
// ServerName from the URL is used.
func SetTLS(val *tls.Config) Option {
	return func(t *Config) error {
		t.TLSClientConfig = val
		return nil
	}
}

// SetAuth The SASL mechanisms to try in the client request, and the successful
// mechanism used on the Connection object.
// If SASL is nil, PlainAuth from the URL is used.
func SetAuth(val []Authentication) Option {
	return func(t *Config) error {
		t.SASL = val
		return nil
	}
}

// SetVhost specifies the namespace of permissions, exchanges, queues and
// bindings on the server.  Dial sets this to the path parsed from the URL.
func SetVhost(val string) Option {
	return func(t *Config) error {
		t.Vhost = val
		return nil
	}
}

// SetChannelMax 0 max channels means 2^16 - 1
func SetChannelMax(val int) Option {
	return func(t *Config) error {
		t.ChannelMax = val
		return nil
	}
}

// SetFrameSize 0 max bytes means unlimited
func SetFrameSize(val int) Option {
	return func(t *Config) error {
		t.FrameSize = val
		return nil
	}
}

// SetHeartbeat ess than 1s uses the server's interval
func SetHeartbeat(val time.Duration) Option {
	return func(t *Config) error {
		t.Heartbeat = val
		return nil
	}
}

// SetProperties is table of properties that the client advertises to the server.
// This is an optional setting - if the application does not set this,
// the underlying library will use a generic set of client properties.
func SetProperties(val Table) Option {
	return func(t *Config) error {
		t.Properties = val
		return nil
	}
}

// SetLocale locale that we expect to always be en_US
// Even though servers must return it as per the AMQP 0-9-1 spec,
// we are not aware of it being used other than to satisfy the spec requirements
func SetLocale(val string) Option {
	return func(t *Config) error {
		t.Locale = val
		return nil
	}
}

// SetDial sets callback which returns a net.Conn prepared for a TLS handshake with TSLClientConfig,
// then an AMQP connection handshake.
func SetDial(val DialFn) Option {
	return func(t *Config) error {
		t.Dial = val
		return nil
	}
}
