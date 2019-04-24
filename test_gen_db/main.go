package main

import (
	"log"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/game_db"
	"github.com/huoshan017/mysql-go/manager"
)

var db_mgr mysql_manager.DB
var db_player game_db.T_Player_Table
var db_player_friend game_db.T_Player_Friend_Table

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
	db_player_friend.Init(&db_mgr)

	id := 3
	var o bool
	var p *game_db.T_Player
	p, o = db_player.Select("id", id)
	if !o {
		log.Printf("cant get result by id %v\n", id)
		return
	}

	log.Printf("get the result %v by id %v\n", p, id)

	var ps []*game_db.T_Player
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

	var ids []int32
	ids = db_player.SelectPrimaryField()
	if ids != nil {
		log.Printf("get primary field list:\n")
		for i, id := range ids {
			log.Printf("	%v: %v\n", i, id)
		}
	}

	var transaction *mysql_base.Transaction = db_mgr.NewTransaction()

	p.AtomicExecute(func(t *game_db.T_Player) {
		t.Set_level(333)
		t.Set_vip_level(333)
		vp_list := t.GetValuePairList([]string{"level", "vip_level"})
		db_player.TransactionUpdateWithFieldPair(transaction, t.Get_id(), vp_list)
	})
	transaction.Done()
	db_mgr.Save()
}
