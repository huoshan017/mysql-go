set GOPATH=%cd%/../../../../../../../

cd ../../db_define
md proto
cd ../src/game

cd ../../../bin
code_generator.exe -c ../example/db_define/game_db.json -d ../example/src/game -p ../example/db_define/proto/game_db.proto
cd ../example

cd protobuf
protoc.exe --go_out=../src/game/game_db --proto_path=../db_define/proto game_db.proto
cd ../src/game

go build -i -o ../../bin/game.exe github.com/huoshan017/mysql-go/example/src/game
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok