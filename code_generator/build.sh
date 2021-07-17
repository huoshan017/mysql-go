#!/bin/bash
export GO111MODULE=on
export GOPROXY=https://goproxy.io
#export GOPATH=$(pwd)/../../../../..
go get -v -u github.com/go-sql-driver/mysql
go get -v -u github.com/hashicorp/golang-lru/simplelru
go build -mod=mod -o ../bin/code_generator github.com/huoshan017/mysql-go/code_generator
