#!/bin/bash

export GOPATH=$(pwd)/../../../..
echo $GOPATH
go get -v -u -t github.com/go-sql-driver/mysql
go build -i -o bin/code_generator github.com/huoshan017/mysql-go/code_generator
