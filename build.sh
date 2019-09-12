#!/bin/bash

#go build github.com/huoshan017/mysql-go/base
#go build github.com/huoshan017/mysql-go/manager
#go build github.com/huoshan017/mysql-go/generate

cd code_generator
bash build.sh
cd ..

cd proxy
bash build.sh
cd ..