#!/bin/sh
$(pwd)/../../../bin/code_generator -c $(pwd)/../../db_define/login_db.json -d $(pwd)/../../src/login -p $(pwd)/../../../_external
go build -o $(pwd)/../../bin/login github.com/huoshan017/mysql-go/example/src/login