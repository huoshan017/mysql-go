package main

import (
	"flag"
	"log"
	"os"

	"github.com/huoshan017/mysql-go/example/src/game/game_db"
	"github.com/huoshan017/mysql-go/manager"
)

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

	var db_mgr mysql_manager.DB

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

	var gd *game_db.T_Global
	gd = db_global_table.GetRow()
	if gd == nil {
		log.Printf("cant get global table data\n")
		return
	}

	gd.Set_curr_guild_id(20)
	gd.Set_curr_player_id(40)

	db_global_table.UpdateWithFVPList(gd.GetFVPList([]string{"curr_guild_id", "curr_mail_id", "curr_player_id"}))

	var o bool
	var p, p2 *game_db.T_Player
	id := 5
	p, o = db_player_table.Select("id", id)
	if !o {
		log.Printf("cant get result by id %v\n", id)
		return
	}

	id = 6
	p2, o = db_player_table.Select("id", id)
	if !o {
		log.Printf("cant get result by id %v\n", id)
		return
	}

	log.Printf("get the result %v by id %v\n", p, id)

	var ids []int32
	ids = db_player_table.SelectAllPrimaryField()
	if ids != nil {
		log.Printf("get primary field list:\n")
		for i, id := range ids {
			log.Printf("	%v: %v\n", i, id)
		}
	}

	var transaction *mysql_manager.Transaction = db_mgr.NewTransaction()

	p.AtomicExecute(func(t *game_db.T_Player) {
		t.Set_level(444)
		t.Set_vip_level(4444)
		t.Set_head(4545)
		db_player_table.TransactionUpdateWithFieldName(transaction, t, []string{"level", "vip_level", "head"})
	})

	p2.AtomicExecute(func(t *game_db.T_Player) {
		t.Set_level(666)
		t.Set_vip_level(6666)
		t.Set_head(6565)
		db_player_table.TransactionUpdateWithFieldName(transaction, t, []string{"level", "vip_level", "head"})
	})

	transaction.Done()

	for level := int32(1); level <= 999; level++ {
		p.Set_level(level)
		p.Set_vip_level(level)
		db_player_table.UpdateWithFieldName(p, []string{"level", "vip_level"})
	}

	db_mgr.Save()
}
