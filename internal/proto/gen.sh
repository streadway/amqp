#!/bin/sh
go run gen.go < amqp0-9-1.stripped.extended.xml | gofmt > spec091.go
