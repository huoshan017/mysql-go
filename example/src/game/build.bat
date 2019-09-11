set GOPATH=%cd%/../../../../../../../

%cd%/../../../bin/code_generator.exe -c %cd%/../../db_define/game_db.json -d %cd%/../../src/game -p %cd%/../../protobuf/protoc.exe

go build -i -o %cd%/../../bin/game.exe github.com/huoshan017/mysql-go/example/src/game
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok