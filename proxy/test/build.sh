#!/bin/bash

export GOPATH=$(pwd)/../../../../../../
export PATH=$PATH:$GOPATH/bin

go get -u -v -t github.com/golang/protobuf/protoc-gen-go

$(pwd)/../../bin/./code_generator -c $(pwd)/db_define/game_db.json -d $(pwd)

go build -i -o ../bin/test_client github.com/huoshan017/mysql-go/proxy/test
