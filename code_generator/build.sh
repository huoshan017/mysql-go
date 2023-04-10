#!/bin/bash
export GOPROXY=https://goproxy.io
go build -mod=mod -o ../bin/code_generator github.com/huoshan017/mysql-go/code_generator
