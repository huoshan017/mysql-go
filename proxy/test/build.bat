set GOPATH=%cd%/../../../../../../

go build -i -o ../bin/test_client.exe github.com/huoshan017/mysql-go/proxy/test
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok