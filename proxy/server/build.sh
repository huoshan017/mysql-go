#!/bin/bash
export GOPROXY=https://goproxy.io
set -x
go build -mod=mod -o ../bin/proxy_server github.com/huoshan017/mysql-go/proxy/server
