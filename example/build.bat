set GOPATH=%cd%/../../../../../

cd db_define
md proto
cd ..

cd ../bin
code_generator.exe -c ../example/db_define/account_db.json -d ../example/src/account -p ../example/db_define/proto/account_db.proto
cd ../example

cd protobuf
protoc.exe --go_out=../src/account/account_db --proto_path=../db_define/proto account_db.proto
cd ..

go build -i -o bin/account.exe github.com/huoshan017/mysql-go/example/src/account
if errorlevel 1 goto exit

cd ../bin
code_generator.exe -c ../example/db_define/login_db.json -d ../example/src/login -p ../example/db_define/proto/login_db.proto
cd ../example

cd protobuf
protoc.exe --go_out=../src/login/login_db --proto_path=../db_define/proto login_db.proto
cd ..

go build -i -o bin/login.exe github.com/huoshan017/mysql-go/example/src/login
if errorlevel 1 goto exit

cd ../bin
code_generator.exe -c ../example/db_define/game_db.json -d ../example/src/game -p ../example/db_define/proto/game_db.proto
cd ../example

cd protobuf
protoc.exe --go_out=../src/game/game_db --proto_path=../db_define/proto game_db.proto
cd ..

go build -i -o bin/game.exe github.com/huoshan017/mysql-go/example/src/game
if errorlevel 1 goto exit
if errorlevel 0 goto ok

:exit
echo build failed !!!

:ok
echo build ok