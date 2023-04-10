#!/bin/sh
$(pwd)/../bin/code_generator -c $(pwd)/config.json -d $(pwd) -p $(pwd)/../_external
go build -o $(pwd)/bin/database_test github.com/huoshan017/mysql-go/tests/database
go build -o $(pwd)/bin/gen_db_test github.com/huoshan017/mysql-go/tests/gen_db