# AMQP

Alpha-class AMQP 0.9.1 client in Go.

# Status

[![Build Status](https://secure.travis-ci.org/streadway/amqp.png)](http://travis-ci.org/streadway/amqp)

Under heavy development, including API changes.  Check the latest
[issues](https://github.com/streadway/amqp/issues) for bugs and enhancements.

# Goals

Provide an low level interface that abstracts the wire protocol and IO,
exposing methods specific to the 0.9.1 specification targeted to RabbitMQ.

# Usage

See the 'examples' subdirectory for simple producers and consumers. If you have
a use-case in mind which isn't well-represented by the examples, please file an
issue, and I'll try to create one.

# Documentation

See [the gopkgdoc page](http://gopkgdoc.appspot.com/github.com/streadway/amqp)
for up-to-the-minute documentation and usage.

# Areas that need work

## Shutdown

* Check leak of go routines after clean shutdown

## Tests

* concurrency
* interruption of synchronous messages
* handle "noise on the line" safely

# Non-goals

Things not intended to be supported (at this stage).

* Auto reconnect and re-establishment of state
* Multiple consumers on a single channel

# Credits

 * [Sean Treadway](https://github.com/streadway)
 * [Peter Bourgon](https://github.com/peterbourgon)
 * [Michael Klishin](https://github.com/michaelklishin)

# License

BSD 2 clause - see LICENSE file for more details.
