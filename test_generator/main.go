package main

import (
	"log"
	//"os/exec"

	"github.com/huoshan017/mysql-go/generator"
)

var config_loader mysql_generator.ConfigLoader

func main() {
	config_file := "../src/github.com/huoshan017/mysql-go/generator/config.json"
	if !config_loader.Load(config_file) {
		return
	}

	if !config_loader.Generate("../src/github.com/huoshan017/mysql-go/test_generator") {
		return
	}

	if !config_loader.GenerateFieldStructsProto("../src/github.com/huoshan017/mysql-go/test_generator") {
		return
	}

	/*cmd := exec.Command("../src/ih_server/third_party/protobuf/protoc.exe", "--go-out=../src/github.com/huoshan017/mysql-go/test_generator/game_db", "--proto_path=../src/github.com/huoshan017/mysql-go/test_generator", "game_db_field_structs.proto")
	if err := cmd.Run(); err != nil {
		log.Printf("execute err: %v", err.Error())
		return
	}*/

	log.Printf("generated proto\n")
}
