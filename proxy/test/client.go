package main

import (
	"log"

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
	var select_names = []string{"account", "role_id", "nick_name", "sex", "level", "vip_level", "exp", "head", "create_time", "token", "items", "skills", "tasks", "role_common", "roles"}
	var account, nick_name, token string
	var role_id int64
	var sex int8
	var level, vip_level, exp, head, create_time int
	var items, skills, tasks, role_common, roles []byte
	var dest_list = []interface{}{&account, &role_id, &nick_name, &sex, &level, &vip_level, &exp, &head, &create_time, &token, &items, &skills, &tasks, &role_common, &roles}
	if !db_proxy.Select(table_name, "id", 1, select_names, dest_list) {
		log.Printf("db proxy select table %v where id=1 failed\n", table_name)
		return
	}
	log.Printf("db proxy selected:\n")
	log.Printf("		account: %v\n		role_id: %v\n		nick_name: %v\n		sex: %v\n		level: %v\n		vip_level: %v\n		exp: %v\n		head: %v",
		account, role_id, nick_name, sex, level, vip_level, exp, head)
	log.Printf("		create_time: %v\n		token: %v\n			items: %v\n		skills: %v\n		tasks: %v\n		role_common: %v\n		roles: %v\n",
		create_time, token, items, skills, tasks, role_common, roles)
}
