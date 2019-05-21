set GOPATH=%cd%/../../../../../../

go build -i -o ../bin/proxy_server.exe github.com/huoshan017/mysql-go/proxy/server
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok