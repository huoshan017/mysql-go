package main

import (
	"log"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/game_db"
	"github.com/huoshan017/mysql-go/manager"
)

var db_mgr mysql_manager.DB

func main() {
	config_path := "../src/github.com/huoshan017/mysql-go/generator/config.json"
	if !db_mgr.LoadConfig(config_path) {
		return
	}
	if !db_mgr.Connect("localhost", "root", "", "game_db") {
		return
	}

	db_mgr.Run()

	game_db.Init(&db_mgr)

	id := 4
	var o bool
	var p *game_db.T_Player
	p, o = game_db.Get_T_Player_Table().Select("id", id)
	if !o {
		log.Printf("cant get result by id %v\n", id)
		return
	}

	log.Printf("get the result %v by id %v\n", p, id)

	var ps []*game_db.T_Player
	ps, o = game_db.Get_T_Player_Table().SelectMulti("", nil)
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
	ids = game_db.Get_T_Player_Table().SelectPrimaryField()
	if ids != nil {
		log.Printf("get primary field list:\n")
		for i, id := range ids {
			log.Printf("	%v: %v\n", i, id)
		}
	}

	var transaction *mysql_base.Transaction = db_mgr.NewTransaction()

	p.AtomicExecute(func(t *game_db.T_Player) {
		t.Set_level(444)
		t.Set_vip_level(4444)
		fvp_list := t.GetValuePairList([]string{"level", "vip_level"})
		game_db.Get_T_Player_Table().TransactionUpdateWithFieldPair(transaction, t.Get_id(), fvp_list)
		game_db.Get_T_Player_Table().UpdateWithFieldPair(t.Get_id(), fvp_list)
	})
	transaction.Done()
	db_mgr.Save()
}
