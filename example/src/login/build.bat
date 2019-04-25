set GOPATH=%cd%/../../../../../../../

cd ../../db_define
md proto
cd ../src/login

cd ../../../bin
code_generator.exe -c ../example/db_define/login_db.json -d ../example/src/login -p ../example/db_define/proto/login_db.proto
cd ../example

cd protobuf
protoc.exe --go_out=../src/login/login_db --proto_path=../db_define/proto login_db.proto
cd ../src/login

go build -i -o ../../bin/login.exe github.com/huoshan017/mysql-go/example/src/login
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok