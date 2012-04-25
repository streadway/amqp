all: build fmt test

.PHONY: gen

wire/spec/spec: wire/spec/gen.go
	echo "Entering directory 'wire/spec'"
	cd wire/spec && go build .
	echo "Entering directory '.'"
	cd ../..

wire/spec091.go: wire/spec/spec
	wire/spec/spec < wire/spec/amqp0-9-1.extended.xml > wire/spec091.go

gen: wire/spec091.go

test: all
	go test

fmt: gen
	go fmt ./...
	go fmt .

build: gen
	go build ./...
	go build .

install: build
	go install ./...
	go install .
