set GOPATH=%cd%/../../../../../

go build -i -o ../bin/code_generator.exe github.com/huoshan017/mysql-go/code_generator
if errorlevel 1 goto exit
if errorlevel 0 goto ok

:exit
echo build failed !!!

:ok
echo build ok