#!/bin/bash

export GOPATH=$(pwd)/../../../../../../
export PATH=$PATH:$GOPATH/bin

go get -u -v -t github.com/golang/protobuf/protoc-gen-go

cd db_define
mkdir -p proto
cd ..

cd ../../bin
./code_generator -c ../proxy/test/db_define/game_db.json -d ../proxy/test -p ../proxy/test/db_define/proto/game_db.proto
cd ../example

cd protobuf
./protoc --go_out=../../proxy/test/game_db --proto_path=../../proxy/test/db_define/proto game_db.proto
cd ../../proxy/test

go build -i -o ../bin/test_client github.com/huoshan017/mysql-go/proxy/test
