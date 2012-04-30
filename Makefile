all: build fmt test

.PHONY: gen

spec/spec: spec/gen.go
	echo "Entering directory 'spec'"
	cd spec && go build .
	echo "Entering directory '.'"
	cd ..

spec091.go: spec/spec
	spec/spec < spec/amqp0-9-1.extended.xml > spec091.go

gen: spec091.go

test: all
	go test

fmt: gen
	go fmt .

build: gen
	go build .

install: build
	go install .
