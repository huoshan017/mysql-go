#!/bin/bash
export GO111MODULE=on
export GOPROXY=https://goproxy.io
#export GOPATH=$(pwd)/../../../../../../
export PATH=$PATH:$GOPATH/bin

go get -u -v github.com/golang/protobuf/protoc-gen-go

$(pwd)/../../bin/./code_generator -c $(pwd)/db_define/game_db.json -d $(pwd) -p $(pwd)/../../_external/

go build -race -mod=mod -o ../bin/test_client github.com/huoshan017/mysql-go/proxy/test
