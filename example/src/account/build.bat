set GOPATH=%cd%/../../../../../../../

%cd%/../../../bin/code_generator.exe -c %cd%/../../db_define/account_db.json -d %cd%/../../src/account -p %cd%/../../protobuf/protoc.exe

go build -i -o %cd%/../../bin/account.exe github.com/huoshan017/mysql-go/example/src/account
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok