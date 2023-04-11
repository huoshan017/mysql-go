package main

import (
	"log"
	"time"

	//"github.com/huoshan017/mysql-go/base"
	mysql_manager "github.com/huoshan017/mysql-go/manager"

	"github.com/huoshan017/mysql-go/example/src/account/account_db"
)

func main() {
	var db_mgr mysql_manager.DB
	err := db_mgr.Connect("localhost", "root", "", "account_db")
	if err != nil {
		log.Printf("connect db err: %v", err.Error())
		return
	}

	db_mgr.Run()

	log.Printf("db running...\n")

	tables := account_db.NewTablesManager(&db_mgr)
	account_table := tables.GetT_AccountTable()

	var account_list = []string{
		"aaa", "bbb", "ccc", "ddd",
	}

	for _, a := range account_list {
		account := account_db.CreateT_Account()
		account.Set_account(a)
		account_table.InsertIgnore(account)
	}

	db_mgr.Save()

	for {
		time.Sleep(time.Second)
	}
}
