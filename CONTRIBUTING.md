## Contributing

The workflow is pretty standard:

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Run integration tests (see below)
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Push to the branch (`git push -u origin my-new-feature`)
6. Submit a pull request

## Running Tests

The test suite assumes that

 * you have a RabbitMQ node running on localhost with all defaults
 * `AMQP_URL` is exported to `amqp://guest:guest@127.0.0.1:5672/`
 * that `GOMAXPROCS` is set to `2` (at least in low power CI environments)

### Run Tests

    AMQP_URL=amqp://guest:guest@127.0.0.1:5672/ GOMAXPROCS=2 go test -v -tags -race ./\...
