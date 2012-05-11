# AMQP client for Go

Work in progress.  Check the development branch for how much work is in progress.

# Goals

Provide an low level interface that abstracts the wire protocol and IO,
exposing methods specific to the 0.9.1 specification.

# Shutdown

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

# TODO

## Shutdown
  * Propagate closing of IO in Framing
  * leaks of go routines

## Tests

	* wire round trip equality
	* concurrency
	* interruption of synchronous messages

# Non-goals

  * Reconnect and re-establishment of state
	* Multiple consumers on a single channel

# Low level

There are 2 primary data interfaces, `Message` and `Frame`.  A `Message`
represents either a synchronous or asychronous request or response that
optionally contains `ContentProperties` and a byte array as a `ContentBody`.
`Messages` are sent and received on a `Channel`. A `Channel` is responsible for
constructing and deconstructing a `Message` into `Frames`.  The `Frames` are
multiplexed and demultiplexed on a `ReadWriteCloser` network socket by the
`Connection`.

The `Connection` and `Channel` handlers capture some basic use cases for establishing and closing the io session and logical channel.
