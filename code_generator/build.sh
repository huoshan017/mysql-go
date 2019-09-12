#!/bin/bash

export GOPATH=$(pwd)/../../../..
go get -v -u github.com/go-sql-driver/mysql
go build -i -o ../bin/code_generator github.com/huoshan017/mysql-go/code_generator
