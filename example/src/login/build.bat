set GOPATH=%cd%/../../../../../../../

%cd%/../../../bin/code_generator.exe -c %cd%/../../db_define/login_db.json -d %cd%/../../src/login -p %cd%/../../protobuf/protoc.exe

go build -i -o %cd%/../../bin/login.exe github.com/huoshan017/mysql-go/example/src/login
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok