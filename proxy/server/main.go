package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	DEFAULT_LISTEN_PORT = 19999
)

type Config struct {
	ListenAddr       string
	DBListConfigPath string
	DBBackupPath     string
}

func (this *Config) Init(config_path string) bool {
	data, err := ioutil.ReadFile(config_path)
	if err != nil {
		log.Printf("read config file err: %v\n", config_path, err.Error())
		return false
	}
	err = json.Unmarshal(data, this)
	if err != nil {
		log.Printf("json unmarshal err: %v\n", err.Error())
		return false
	}
	return true
}

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
		log.Printf("not specified config file arg\n")
		return
	}

	var config Config
	if !config.Init(config_path) {
		return
	}

	root_path, _ := path.Split(config_path)
	err := db_list.Load(root_path + config.DBListConfigPath)
	if err != nil {
		log.Printf("%v\n", err.Error())
		return
	}

	SetDebug(true)

	listen_address := config.ListenAddr
	var proc_service ProcService
	if !strings.Contains(listen_address, ":") {
		listen_address += (":" + strconv.Itoa(DEFAULT_LISTEN_PORT))
	}
	err = proc_service.Start(listen_address)
	if err != nil {
		log.Printf("%v\n", err.Error())
	}
}
