# AMQP client for Go

Work in progress.  Check the development branch for how much work is in progress.

# Goals

Provide an interface that makes managing the queue/exchange space easy.
Provide a high level interface to capture the common task of sending and receiving small messages.

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
