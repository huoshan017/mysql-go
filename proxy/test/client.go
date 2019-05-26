package main

import (
	"log"
	"time"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/proxy/client"
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

	var table_name string = "t_player"

	/*go func() {
		var select_names = []string{"account", "role_id", "nick_name", "sex", "level", "vip_level", "exp", "head", "create_time", "token", "items", "skills", "tasks", "role_common", "roles"}
		var account, nick_name, token string
		var role_id int64
		var sex int8
		var level, vip_level, exp, head, create_time int32
		var items, skills, tasks, role_common, roles []byte
		var dest_list = []interface{}{&account, &role_id, &nick_name, &sex, &level, &vip_level, &exp, &head, &create_time, &token, &items, &skills, &tasks, &role_common, &roles}
		for {
			var result_list mysql_proxy.QueryResultList
			if !db_proxy.SelectAllRecords(table_name, select_names, &result_list) {
				log.Printf("db proxy select table %v with select_names %v failed\n", table_name, select_names)
				return
			}

			var idx int
			log.Printf("db proxy selected all: \n")
			for {
				if !result_list.Get(dest_list...) {
					break
				}
				log.Printf("  %v	account:%v  role_id:%v  nick_name:%v  sex:%v  level:%v  vip_level:%v  exp:%v  head:%v  create_time:%v  token:%v  items:%v  skills:%v  tasks:%v  role_common:%v  roles:%v\n",
					idx+1, account, role_id, nick_name, sex, level, vip_level, exp, head, create_time, token, items, skills, tasks, role_common, roles)
				idx += 1
			}
			time.Sleep(time.Minute * 50)
		}
	}()*/

	for i := 0; i < 10000; i++ {
		var field_pairs = []*mysql_base.FieldValuePair{
			&mysql_base.FieldValuePair{
				Name:  "id",
				Value: 888 + i,
			},
			&mysql_base.FieldValuePair{
				Name:  "role_id",
				Value: 676767 + i,
			},
		}
		db_proxy.Insert(table_name, field_pairs)

		field_pairs = []*mysql_base.FieldValuePair{
			&mysql_base.FieldValuePair{
				Name:  "vip_level",
				Value: 200 + i,
			},
		}
		db_proxy.Update(table_name, "id", 1, field_pairs)

		time.Sleep(time.Millisecond * 100)
	}

	for {
		time.Sleep(time.Second)
	}
}
