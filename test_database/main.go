package main

import (
	"log"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/generator"
)

func main() {
	var config_loader mysql_generator.ConfigLoader
	var database mysql_base.Database

	config_path := "../src/github.com/huoshan017/mysql-go/generator/config.json"
	if !config_loader.Load(config_path) {
		log.Printf("load config %v failed\n", config_path)
		return
	}

	err := database.Open("localhost", "root", "", config_loader.DBPkg)
	if err != nil {
		log.Printf("open database err %v\n", err.Error())
		return
	}

	if config_loader.Tables != nil {
		for _, t := range config_loader.Tables {
			if !database.LoadTable(t) {
				log.Printf("load table %v config failed\n", t.Name)
				return
			}
		}
	}

	log.Printf("database loaded\n")
}
