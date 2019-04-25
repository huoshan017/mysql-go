set GOPATH=%cd%/../../../../../../../

cd ../../db_define
md proto
cd ../src/account

cd ../../../bin
code_generator.exe -c ../example/db_define/account_db.json -d ../example/src/account -p ../example/db_define/proto/account_db.proto
cd ../example

cd protobuf
protoc.exe --go_out=../src/account/account_db --proto_path=../db_define/proto account_db.proto
cd ../src/account

go build -i -o ../../bin/account.exe github.com/huoshan017/mysql-go/example/src/account
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok