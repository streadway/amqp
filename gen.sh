#!/bin/sh

cd spec
go build . && ./spec < amqp0-9-1.extended.xml | gofmt > ../spec091.go
cd ..
