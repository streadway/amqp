# AMQP client for Go

Work in progress.  Check the development branch for how much work is in progress.

# Goals

Provide an low level interface that abstracts the wire protocol and IO,
exposing methods specific to the 0.9.1 specification.

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
