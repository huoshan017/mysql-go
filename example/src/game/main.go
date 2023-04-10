package main

import (
	"flag"
	"log"
	"os"

	"github.com/huoshan017/mysql-go/example/src/game/game_db"
	mysql_manager "github.com/huoshan017/mysql-go/manager"
)

func main() {
	if len(os.Args) < 5 {
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

	var arg_user, arg_host, arg_password *string
	// 用戶
	arg_user = flag.String("u", "", "user")
	// 主機
	arg_host = flag.String("h", "", "host")
	// 密碼
	arg_password = flag.String("p", "", "password")
	flag.Parse()

	var user string
	if nil != arg_user && *arg_user != "" {
		user = *arg_user
	} else {
		user = "root"
	}

	var host string
	if nil != arg_host && *arg_host != "" {
		host = *arg_host
	} else {
		host = "127.0.0.1"
	}

	var password string
	if nil != arg_password && *arg_password != "" {
		password = *arg_password
	}

	var db_mgr mysql_manager.DB
	if !db_mgr.LoadConfig(config_path) {
		return
	}

	var err error
	err = db_mgr.Connect(host, user, password, "game_db")
	if err != nil {
		log.Printf("connect db err: %v\n", err.Error())
		return
	}

	go db_mgr.Run()

	tb_mgr := game_db.NewTablesManager(&db_mgr)
	db_player_table := tb_mgr.GetT_PlayerTable()
	db_global_table := tb_mgr.GetT_GlobalTable()

	var gd *game_db.T_Global
	gd, err = db_global_table.GetRow()
	if err != nil {
		log.Printf("cant get global table data: %v\n", err.Error())
		return
	}

	gd.Set_curr_guild_id(20)
	gd.Set_curr_player_id(40)

	db_global_table.UpdateWithFVPList(gd.GetFVPList([]string{"curr_guild_id", "curr_mail_id", "curr_player_id"}))

	var p, p2 *game_db.T_Player
	id := 5
	p, err = db_player_table.Select("id", id)
	if err != nil {
		log.Printf("cant get result by id %v err: %v\n", id, err.Error())
		return
	}

	id = 6
	p2, err = db_player_table.Select("id", id)
	if err != nil {
		log.Printf("cant get result by id %v err: %v\n", id, err.Error())
		return
	}

	log.Printf("get the result %v by id %v\n", p, id)

	var ids []uint32
	ids, err = db_player_table.SelectAllPrimaryField()
	if err != nil {
		log.Printf("get primary field list err: %v\n", err.Error())
		return
	}

	log.Printf("get primary field list: %v\n", err.Error())
	for i, id := range ids {
		log.Printf("	%v: %v\n", i, id)
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

	for level := uint32(1); level <= 999; level++ {
		p.Set_level(level)
		p.Set_vip_level(level)
		db_player_table.UpdateWithFieldName(p, []string{"level", "vip_level"})
	}

	db_mgr.Save()
}
