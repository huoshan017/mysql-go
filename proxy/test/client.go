package main

import (
	"log"
	"time"

	"github.com/huoshan017/mysql-go/proxy/client"
	"github.com/huoshan017/mysql-go/proxy/test/game_db"
)

func main() {
	var proxy_addr string = "127.0.0.1:1999"
	var db_host_id int32 = 1
	var db_host_alias string = "main"
	var db_name = "game_db"
	var db_proxy mysql_proxy.DB
	if !db_proxy.Connect(proxy_addr, db_host_id, db_host_alias, db_name) {
		log.Printf("db proxy connect %v failed\n", proxy_addr)
		return
	}

	tb_mgr := game_db.NewTableProxysManager(&db_proxy)
	player_table_proxy := tb_mgr.Get_T_Player_Table_Proxy()

	field_name := "id"
	go func() {
		var id int = 1
		for ; id < 10; id++ {
			p, o := player_table_proxy.Select(field_name, id)
			if !o {
				log.Printf("select id %v failed\n", id)
				continue
			}

			log.Printf("selected player: %v\n", p)
		}
	}()

	for {
		time.Sleep(time.Second)
	}
}
