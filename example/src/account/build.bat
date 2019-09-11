set GOPATH=%cd%/../../../../../../../

cd ../../../bin
code_generator.exe -c ../example/db_define/account_db.json -d ../example/src/account -p ../example/protobuf/protoc.exe

go build -i -o ../../bin/account.exe github.com/huoshan017/mysql-go/example/src/account
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok