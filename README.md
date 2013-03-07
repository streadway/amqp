# AMQP

AMQP 0.9.1 client with RabbitMQ extensions in Go.

# Status

*Beta*

[![Build Status](https://secure.travis-ci.org/streadway/amqp.png)](http://travis-ci.org/streadway/amqp)

API changes unlikely and will be discussed on [Github
issues](https://github.com/streadway/amqp/issues) along with any bugs or
enhancements.

# Goals

Provide an functional interface that closely represents the AMQP 0.9.1 model
targeted to RabbitMQ as a server.

# Non-goals

Things not intended to be supported.

  * Auto reconnect and re-synchronization of client and server topologies.
  * AMQP Protocol negotiation for forward or backward compatibility.
  * Anything other than PLAIN and EXTERNAL authentication mechanisms.

# Usage

See the 'examples' subdirectory for simple producers and consumers executables.
If you have a use-case in mind which isn't well-represented by the examples,
please file an issue.

# Documentation

See [the gopkgdoc page](http://gopkgdoc.appspot.com/github.com/streadway/amqp)
for up-to-the-minute documentation and usage.

# Contributing

Pull requests are very much welcomed.  Create your pull request on a non-master
branch, make sure a test or example is included that covers your change and
your commits represent coherent changes that include a reason for the change.

To run the integration tests, make sure you have RabbitMQ running on any host,
export the environment variable `AMQP_URL=amqp://host/` and run `go test`.
TravisCI will also run the integration tests.

# Credits

 * [Sean Treadway](https://github.com/streadway)
 * [Peter Bourgon](https://github.com/peterbourgon)
 * [Michael Klishin](https://github.com/michaelklishin)
 * [Richard Musiol](https://github.com/neelance)
 * [Dan Markham](https://github.com/dmarkham)

# License

BSD 2 clause - see LICENSE for more details.


