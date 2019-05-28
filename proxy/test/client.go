package main

import (
	"log"
	"time"

	"github.com/huoshan017/mysql-go/proxy/client"
	"github.com/huoshan017/mysql-go/proxy/test/game_db"
)

func main() {
	var proxy_addr string = "127.0.0.1:19999"
	var db_host_id int32 = 1
	var db_host_alias string = "main"
	var db_name = "game_db"
	var db_proxy mysql_proxy.DB
	if !db_proxy.Connect(proxy_addr, db_host_id, db_host_alias, db_name) {
		log.Printf("db proxy connect %v failed\n", proxy_addr)
		return
	}

	db_proxy.Run()

	tb_mgr := game_db.NewTableProxysManager(&db_proxy)
	player_table_proxy := tb_mgr.Get_T_Player_Table_Proxy()

	field_name := "id"
	var p *game_db.T_Player
	var ps []*game_db.T_Player
	var o bool
	go func() {
		var id int = 1
		for ; id < 10; id++ {
			p, o = player_table_proxy.Select(field_name, id)
			if !o {
				log.Printf("select id %v failed\n", id)
				continue
			}

			log.Printf("selected player: %v\n", p)
		}

		ps, o = player_table_proxy.SelectRecords("level", 1)
		if !o {
			log.Printf("selected player records failed\n")
		} else {
			log.Printf("selected players: %v\n", ps)
		}

		ps, o = player_table_proxy.SelectAllRecords()
		if !o {
			log.Printf("selected all player records failed\n")
		} else {
			log.Printf("selected all players: %v\n", ps)
		}

		ps, o = player_table_proxy.SelectRecordsCondition("vip_level", 1, nil)
		if !o {
			log.Printf("selected records condition failed\n")
		} else {
			log.Printf("selected records condition players: %v\n", ps)
		}
	}()

	for {
		time.Sleep(time.Second)
	}
}
