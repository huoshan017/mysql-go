package main

import (
	"flag"
	"log"
	"os"

	"github.com/huoshan017/mysql-go/generate"
)

var config_loader mysql_generate.ConfigLoader

func main() {
	if len(os.Args) < 4 {
		log.Printf("args num not enough\n")
		return
	}

	arg_config_file := flag.String("c", "", "config file path")
	arg_dest_path := flag.String("d", "", "dest source path")
	arg_dest_proto := flag.String("p", "", "dest proto file")
	flag.Parse()

	var config_path string
	if nil != arg_config_file {
		//flag.Parse()
		config_path = *arg_config_file
		log.Printf("config file path %v\n", config_path)
	} else {
		log.Printf("not found config file arg\n")
		return
	}

	var dest_path string
	if nil != arg_dest_path {
		//flag.Parse()
		dest_path = *arg_dest_path
		log.Printf("dest path %v\n", dest_path)
	} else {
		log.Printf("not found dest path arg\n")
		return
	}

	var dest_proto string
	if nil != arg_dest_proto {
		//flag.Parse()
		dest_proto = *arg_dest_proto
		log.Printf("dest proto %v\n", dest_proto)
	} else {
		log.Printf("not found dest proto arg\n")
		return
	}

	if !config_loader.Load(config_path) {
		return
	}

	if !config_loader.Generate(dest_path) {
		return
	}

	log.Printf("generated source\n")

	if !config_loader.GenerateFieldStructsProto(dest_proto) {
		return
	}

	log.Printf("generated proto\n")

	if !config_loader.GenerateInitFunc(dest_path) {
		return
	}

	log.Printf("generated init funcs\n")

	log.Printf("generated all\n")
}
