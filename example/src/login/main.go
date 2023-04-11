package main

import (
	"log"

	//"github.com/huoshan017/mysql-go/base"
	mysql_manager "github.com/huoshan017/mysql-go/manager"

	"github.com/huoshan017/mysql-go/example/src/login/login_db"
)

var db_mgr mysql_manager.DB

func main() {
	err := db_mgr.Connect("localhost", "root", "", "login_db")
	if err != nil {
		log.Printf("connect db err: %v\n", err.Error())
		return
	}

	db_mgr.Run()

	tables := login_db.NewTablesManager(&db_mgr)
	ban_player_table := tables.GetT_Ban_PlayerTable()
	ban_player := ban_player_table.NewRecord("bbb")
	ban_player_table.Insert(ban_player)

	db_mgr.Save()
}
