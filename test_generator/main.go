package main

import (
	"log"

	"github.com/huoshan017/mysql-go/generator"
)

var config_loader mysql_generator.ConfigLoader

func main() {
	config_file := "../src/github.com/huoshan017/mysql-go/generator/config.json"
	if !config_loader.Load(config_file) {
		return
	}

	if !config_loader.GenerateFieldStructsProto("../src/github.com/huoshan017/mysql-go/generator") {
		return
	}

	log.Printf("generated proto\n")
}
