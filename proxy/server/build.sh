#!/bin/bash

set GOPATH=$(pwd)/../../../../../../
go build -i -o ../bin/proxy_server github.com/huoshan017/mysql-go/proxy/server
