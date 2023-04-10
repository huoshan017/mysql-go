#!/bin/sh
$(pwd)/../../../bin/code_generator -c $(pwd)/../../db_define/game_db.json -d $(pwd)/../../src/game -p $(pwd)/../../../_external
go build -i -o $(pwd)/../../bin/game github.com/huoshan017/mysql-go/example/src/game