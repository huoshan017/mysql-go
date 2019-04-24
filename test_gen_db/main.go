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

	tb_mgr := game_db.NewTablesManager(&db_mgr)
	db_player_table := tb_mgr.Get_T_Player_Table()
	db_global_table := tb_mgr.Get_T_Global_Table()

	id := 5
	var o bool
	var gd *game_db.T_Global
	gd, o = db_global_table.Select()
	if !o {
		log.Printf("cant get global table data\n")
		return
	}

	gd.Set_curr_guild_id(2)
	gd.Set_curr_mail_id(3)
	gd.Set_curr_player_id(4)
	db_global_table.UpdateAll(gd)

	var p *game_db.T_Player
	p, o = db_player_table.Select("id", id)
	if !o {
		log.Printf("cant get result by id %v\n", id)
		return
	}

	log.Printf("get the result %v by id %v\n", p, id)

	var ps []*game_db.T_Player
	ps, o = db_player_table.SelectMulti("", nil)
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
	ids = db_player_table.SelectPrimaryField()
	if ids != nil {
		log.Printf("get primary field list:\n")
		for i, id := range ids {
			log.Printf("	%v: %v\n", i, id)
		}
	}

	var transaction *mysql_base.Transaction = db_mgr.NewTransaction()

	p.AtomicExecute(func(t *game_db.T_Player) {
		t.Set_level(555)
		t.Set_vip_level(5555)
		fvp_list := t.GetValuePairList([]string{"level", "vip_level"})
		db_player_table.TransactionUpdateWithFieldPair(transaction, t.Get_id(), fvp_list)
	})
	transaction.Done()
	db_mgr.Save()
}
