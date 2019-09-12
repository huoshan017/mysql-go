set GOPATH=%cd%/../../../../../../

%cd%/../../bin/code_generator.exe -c %cd%/db_define/game_db.json -d %cd%

go build -i -o %cd%/../bin/test_client.exe github.com/huoshan017/mysql-go/proxy/test
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok