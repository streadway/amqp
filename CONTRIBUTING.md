## Contributing

The workflow is pretty standard:

1. Fork github.com/streadway/amqp
1. Add the pre-commit hook: `ln -s ../../pre-commit .git/hooks/pre-commit`
1. Create your feature branch (`git checkout -b my-new-feature`)
1. Run integration tests (see below)
1. Implement tests.
1. Implement fixs.
1. Commit your changes (`git commit -am 'Add some feature'`)
1. Push to a branch (`git push -u origin my-new-feature`)
1. Submit a pull request

## Running Tests

The test suite assumes that

 * you have a RabbitMQ node running on localhost with all defaults
 * `AMQP_URL` is exported to `amqp://guest:guest@127.0.0.1:5672/`

### Integration Tests

  Integration tests should use the `integrationConnection(...)` test helpers defined
  in `integration_test.go` to setup the integration environment and logging.

  Run integration tests with the following:

    env AMQP_URL=amqp://guest:guest@127.0.0.1:5672/ go test -v -cpu 2 -tags integration -race
