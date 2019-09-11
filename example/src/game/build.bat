set GOPATH=%cd%/../../../../../../../

cd ../../../bin
code_generator.exe -c ../example/db_define/game_db.json -d ../example/src/game -p ../example/protobuf/protoc.exe

go build -i -o ../../bin/game.exe github.com/huoshan017/mysql-go/example/src/game
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok