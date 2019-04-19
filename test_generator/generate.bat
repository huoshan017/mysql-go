cd ../../../../../bin
test_generator.exe
cd ../src/github.com/huoshan017/mysql-go/test_generator

cd ../../../../ih_server/third_party/protobuf

protoc.exe --go_out=../../../github.com/huoshan017/mysql-go/test_generator/game_db --proto_path=../../../github.com/huoshan017/mysql-go/test_generator game_db_field_structs.proto

if errorlevel 1 goto exit

cd ../../../github.com/huoshan017/mysql-go/test_generator

goto ok

:exit
echo gen message failed!!!!!!!!!!!!!!!!!!!!!!!!!!!!

:ok
echo gen message ok