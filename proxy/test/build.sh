#!/bin/bash

export GOPATH=$(pwd)/../../../../../../
export PATH=$PATH:$GOPATH/bin

go get -u -v -t github.com/golang/protobuf/protoc-gen-go

cd ../../bin
./code_generator -c ../proxy/test/db_define/game_db.json -d ../proxy/test -p ../example/protobuf/protoc

go build -i -o ../bin/test_client github.com/huoshan017/mysql-go/proxy/test
