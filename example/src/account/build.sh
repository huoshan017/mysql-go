#!/bin/sh
$(pwd)/../../../bin/code_generator -c $(pwd)/../../db_define/account_db.json -d $(pwd)/../../src/account -p $(pwd)/../../../_external

go build -o $(pwd)/../../bin/account github.com/huoshan017/mysql-go/example/src/account