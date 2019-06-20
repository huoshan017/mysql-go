package main

import (
	"flag"
	"log"
	"os"

	//"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/manager"

	"github.com/huoshan017/mysql-go/example/src/login/login_db"
)

var db_mgr mysql_manager.DB

func main() {
	if len(os.Args) < 2 {
		log.Printf("args not enough\n")
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

	if !db_mgr.LoadConfig(config_path) {
		return
	}

	err := db_mgr.Connect("localhost", "root", "", "login_db")
	if err != nil {
		log.Printf("connect db err: %v\n", err.Error())
		return
	}

	defer db_mgr.Close()

	db_mgr.Run()

	tables := login_db.NewTablesManager(&db_mgr)
	ban_player_table := tables.GetT_Ban_PlayerTable()
	ban_player := ban_player_table.NewRecord("bbb")
	ban_player_table.Insert(ban_player)

	db_mgr.Save()
}
