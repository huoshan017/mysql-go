package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	mysql_manager "github.com/huoshan017/mysql-go/manager"
	"github.com/huoshan017/mysql-go/tests/game_db"
)

var db_mgr mysql_manager.DB

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "args num not enough\n")
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

	//config_path := "../config.json"
	//if !db_mgr.LoadConfig(config_path) {
	//	return
	//}

	fmt.Fprintf(os.Stdout, "host=%v  user=%v  password=%v\n", host, user, password)
	if err := db_mgr.Connect(host, user, password, "game_db"); err != nil {
		log.Panicf("connect db err: %v", err)
		return
	}
	go db_mgr.Run()

	tb_mgr := game_db.NewTablesManager(&db_mgr)
	db_player_table := tb_mgr.GetT_PlayerTable()
	db_global_table := tb_mgr.GetT_GlobalTable()

	id := 5
	var e error
	var gd *game_db.T_Global
	gd, e = db_global_table.Select()
	if e != nil {
		log.Printf("select global table data err %v\n", e)
		return
	}

	gd.Set_curr_guild_id(20)
	gd.Set_curr_mail_id(30)
	gd.Set_curr_player_id(40)
	//db_global_table.UpdateAll(gd)

	db_global_table.UpdateWithFieldName(gd, []string{"curr_guild_id", "curr_mail_id", "curr_player_id"})

	var p *game_db.T_Player
	p, e = db_player_table.Select("id", id)
	if e != nil {
		log.Printf("get result by id %v err %v\n", id, e)
		return
	}

	log.Printf("get the result %v by id %v\n", p, id)

	var ps []*game_db.T_Player
	ps, e = db_player_table.SelectAllRecords()
	if e != nil {
		log.Printf("get all player list err %v\n", e)
		return
	}

	if ps != nil {
		log.Printf("get the result list:\n")
		for i, p := range ps {
			log.Printf("	%v: %v\n", i, p)
		}
	}

	var ids []uint32
	ids, e = db_player_table.SelectAllPrimaryField()
	if e == nil {
		log.Printf("get primary field list:\n")
		for i, id := range ids {
			log.Printf("	%v: %v\n", i, id)
		}
	}

	var transaction *mysql_manager.Transaction = db_mgr.NewTransaction()

	p.AtomicExecute(func(t *game_db.T_Player) {
		t.Set_level(555)
		t.Set_vip_level(5555)
		fvp_list := t.GetFVPList([]string{"level", "vip_level"})
		db_player_table.TransactionUpdateWithFVPList(transaction, t.Get_id(), fvp_list)
	})
	transaction.Done()
	db_mgr.Save()
}
