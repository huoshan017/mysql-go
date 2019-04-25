package main

import (
	"flag"
	"log"
	"os"
	"time"

	//"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/manager"

	"github.com/huoshan017/mysql-go/example/src/account/account_db"
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
	if !db_mgr.Connect("localhost", "root", "", "account_db") {
		return
	}

	db_mgr.Run()

	log.Printf("db running...\n")

	tables := account_db.NewTablesManager(&db_mgr)
	account_table := tables.Get_T_Account_Table()

	var account_list = []string{
		"aaa", "bbb", "ccc", "ddd",
	}

	for _, a := range account_list {
		account := account_db.Create_T_Account()
		account.Set_account(a)
		account_table.Insert(account)
	}

	db_mgr.Save()

	for {
		time.Sleep(time.Second)
	}
}
