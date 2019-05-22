package main

import (
	"flag"
	"log"
	"os"
)

var db_list DbList

func main() {
	if len(os.Args) < 2 {
		log.Printf("args not enough, must specify a config file for db define\n")
		return
	}

	arg_config_file := flag.String("c", "", "config file path")
	flag.Parse()

	var config_path string
	if nil != arg_config_file {
		config_path = *arg_config_file
		log.Printf("config file path %v\n", config_path)
	} else {
		log.Printf("not found config file arg\n")
		return
	}

	err := db_list.Load(config_path)
	if err != nil {
		log.Printf("%v", err.Error())
		return
	}
}
