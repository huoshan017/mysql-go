package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DEFAULT_LISTEN_PORT = 19999
)

func main() {
	if len(os.Args) < 2 {
		log.Printf("args not enough, must specify a config file for db define\n")
		return
	}

	arg_config_file := flag.String("c", "", "config file path")
	arg_listen_address := flag.String("l", "", "config listen address")
	flag.Parse()

	var config_path string
	if nil != arg_config_file {
		config_path = *arg_config_file
		log.Printf("config file path %v\n", config_path)
	} else {
		log.Printf("not specified config file arg\n")
		return
	}

	var listen_address string
	if len(os.Args) >= 3 {
		if nil != arg_listen_address {
			listen_address = *arg_listen_address
		} else {
			log.Printf("not specified listen address arg\n")
			return
		}
	}

	err := db_list.Load(config_path)
	if err != nil {
		log.Printf("load db list config file failed: %v\n", err.Error())
		return
	}

	var proc_service ProcService
	if !strings.Contains(listen_address, ":") {
		listen_address += (":" + strconv.Itoa(DEFAULT_LISTEN_PORT))
	}
	proc_service.Start(listen_address)
}
