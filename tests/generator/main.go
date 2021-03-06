package main

import (
	"log"
	//"os/exec"

	"github.com/huoshan017/mysql-go/generate"
)

var config_loader mysql_generate.ConfigLoader

func main() {
	config_file := "../src/github.com/huoshan017/mysql-go/generate/config.json"
	if !config_loader.Load(config_file) {
		return
	}

	if !config_loader.Generate("../src/github.com/huoshan017/mysql-go") {
		return
	}

	log.Printf("generated source\n")

	if !config_loader.GenerateFieldStructsProto("../src/github.com/huoshan017/mysql-go/test_generate") {
		return
	}

	log.Printf("generated proto\n")

	if !config_loader.GenerateInitFunc("../src/github.com/huoshan017/mysql-go") {
		return
	}

	log.Printf("generated init funcs\n")

	log.Printf("generated all\n")
}
