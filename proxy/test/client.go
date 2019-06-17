package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/huoshan017/mysql-go/proxy/client"
	"github.com/huoshan017/mysql-go/proxy/test/game_db"
)

func main() {
	if len(os.Args) < 1 {
		log.Printf("args not enough, must specify a config file for db define\n")
		return
	}

	host_arg := flag.String("h", "", "config host server")
	flag.Parse()

	var host string
	if nil != host_arg {
		host = *host_arg
		log.Printf("config file path %v\n", host)
	} else {
		log.Printf("not specified config file arg\n")
		return
	}

	var proxy_addr string = host
	var db_host_id int32 = 1
	var db_host_alias string = "main"
	var db_name = "game_db"
	var db_proxy mysql_proxy.DB
	err := db_proxy.Connect(proxy_addr, db_host_id, db_host_alias, db_name)
	if err != nil {
		log.Printf("db proxy connect err %v\n", err.Error())
		return
	}

	db_proxy.GoRun()

	tb_mgr := game_db.NewTablesProxyManager(&db_proxy)
	player_table_proxy := tb_mgr.Get_T_Player_Table_Proxy()

	field_name := "id"
	go func() {
		var err error
		var p *game_db.T_Player
		var ps []*game_db.T_Player
		var id int = 1
		for ; id < 10; id++ {
			p, err = player_table_proxy.Select(field_name, id)
			if err != nil {
				log.Printf("select id %v err %v\n", id, err.Error())
				continue
			}

			log.Printf("selected player: %v\n", p)
		}

		for i := 0; i < 1000; i++ {
			ps, err = player_table_proxy.SelectRecords("level", 1)
			if err != nil {
				log.Printf("selected player records err %v\n", err.Error())
			} else {
				log.Printf("selected players: %v\n", ps)
			}

			ps, err = player_table_proxy.SelectAllRecords()
			if err != nil {
				log.Printf("selected all player records err %v\n", err.Error())
			} else {
				log.Printf("selected all players: %v\n", ps)
			}

			ps, err = player_table_proxy.SelectRecordsCondition("vip_level", 1, nil)
			if err != nil {
				log.Printf("selected records condition err %v\n", err.Error())
			} else {
				log.Printf("selected records condition players: %v\n", ps)
			}

			time.Sleep(time.Millisecond * 50)
		}
	}()

	for {
		time.Sleep(time.Second)
	}
}
