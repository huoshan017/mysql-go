#!/bin/bash
export GO111MODULE=on
export GOPROXY=https://goproxy.io
#export GOPATH=$(pwd)/../../../../../../
set -x
#go mod vendor 
go build -mod=mod -o ../bin/proxy_server github.com/huoshan017/mysql-go/proxy/server
