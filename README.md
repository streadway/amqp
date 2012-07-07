# AMQP

Beta-class AMQP 0.9.1 client in Go.

# Goals

Provide an low level interface that abstracts the wire protocol and IO,
exposing methods specific to the 0.9.1 specification.

# Usage

See the 'examples' subdirectory for simple producers and consumers. If you have
a use-case in mind which isn't well-represented by the examples, please file an
issue, and I'll try to create one.

# Documentation

See [the gopkgdoc page](http://gopkgdoc.appspot.com/github.com/streadway/amqp)
for up-to-the-minute documentation and usage.

# Areas that need work

## Shutdown

* (Better) propagate closing of IO in Framing
* leaks of go routines

    S:C ConnectionClose
    S:C ConnectionCloseOk
    S:C ChannelClose
    S:C ChannelCloseOk
    Read EOF -> ignore
    Write EOF

    close(chan(id).rpc)
    close(chan(id).deliveries)
    close(conn.in)
    close(conn.out)
    conn.rw.Close()

## Tests

* wire round trip equality
* concurrency
* interruption of synchronous messages
* handle "noise on the line" safely

# Non-goals

Things not intended to be supported (at this stage).

* Auto reconnect and re-establishment of state
* Multiple consumers on a single channel

# License

All Rights Reserved http://github.com/streadway (for now).
