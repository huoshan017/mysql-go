package main

import (
	"log"

	"github.com/huoshan017/mysql-go/game_db"
	"github.com/huoshan017/mysql-go/manager"
)

var db_mgr mysql_manager.DB
var db_player game_db.T_playerTable
var db_player2 game_db.T_player2Table

func main() {
	config_path := "../src/github.com/huoshan017/mysql-go/generator/config.json"
	if !db_mgr.LoadConfig(config_path) {
		return
	}
	if !db_mgr.Connect("localhost", "root", "", "game_db") {
		return
	}

	db_mgr.Run()

	db_player.Init(&db_mgr)
	db_player2.Init(&db_mgr)

	id := 1
	var o bool
	var p *game_db.T_player
	p, o = db_player.Select("id", 1)
	if !o {
		log.Printf("cant get result by id %v\n", id)
		return
	}

	log.Printf("get the result %v by id %v\n", p, id)

	var ps []*game_db.T_player
	ps, o = db_player.SelectMulti("", nil)
	if !o {
		log.Printf("cant get result list\n")
		return
	}

	if ps != nil {
		log.Printf("get the result list:\n")
		for i, p := range ps {
			log.Printf("	%v: %v\n", i, p)
		}
	}
}
