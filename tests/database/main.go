package main

import (
	"fmt"
	"log"

	"reflect"
	"strings"
	"time"

	mysql_base "github.com/huoshan017/mysql-go/base"
	mysql_manager "github.com/huoshan017/mysql-go/manager"
)

var db_mgr mysql_manager.DB
var transaction *mysql_manager.Transaction

func main() {
	//config_path := "../src/github.com/huoshan017/mysql-go/example/config.json"

	/*if !db_mgr.LoadConfig(config_path) {
		return
	}*/

	if err := db_mgr.Connect("localhost", "root", "", "game_db"); err != nil {
		log.Panicf("connect db err: %v", err)
		return
	}
	go db_mgr.Run()

	for {
		on_tick()
		time.Sleep(time.Second)
	}
}

func do_insert(strs []string) {
	table_name := strs[1]
	var field_list []*mysql_base.FieldValuePair
	for i := 2; i < len(strs); i += 2 {
		field_list = append(field_list, &mysql_base.FieldValuePair{
			Name:  strs[i],
			Value: strs[i+1],
		})
	}
	db_mgr.Insert(table_name, field_list)
}

func do_select(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]

	var field_list []string
	if len(strs) > 4 {
		for i := 4; i < len(strs); i++ {
			field_list = append(field_list, strs[i])
		}
	}
	table := db_mgr.GetConfigLoader().GetTable(table_name)
	if table == nil {
		log.Printf("没有表%v\n", table_name)
		return
	}

	var dest_list []interface{}
	for _, field_name := range field_list {
		field := table.GetField(field_name)
		if mysql_base.IsMysqlFieldIntType(field.Type) {
			dest_list = append(dest_list, new(int))
		} else if mysql_base.IsMysqlFieldTextType(field.Type) {
			dest_list = append(dest_list, new(string))
		} else if mysql_base.IsMysqlFieldBinaryType(field.Type) || mysql_base.IsMysqlFieldBlobType(field.Type) {
			dest_list = append(dest_list, new([]byte))
		} else {
			log.Printf("不支持的select字段类型 %v\n", field.Type)
		}
	}
	if db_mgr.Select(table_name, key, value, field_list, dest_list) == nil {
		log.Printf("select结果: \n")
		for i := 0; i < len(field_list); i++ {
			/*if len(dest_list) <= i {
				break
			}*/
			v := reflect.ValueOf(dest_list[i])
			e := v.Elem()
			t := e.Kind()
			if t == reflect.Int {
				log.Printf("		%v: %v\n", field_list[i], *dest_list[i].(*int))
			} else if t == reflect.String {
				log.Printf("		%v: %v\n", field_list[i], *dest_list[i].(*string))
			} else if t == reflect.Slice {
				log.Printf("		%v: %v\n", field_list[i], dest_list[i].([]byte))
			} else {
				log.Printf("		unprocessed reflect kind %v with index %v\n", t, i)
			}
		}
	}

}

func do_select_star(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	var dest_list []interface{}
	if err := db_mgr.SelectStar(table_name, key, value, dest_list); err != nil {
		log.Panicf("select star err: %v", err)
		return
	}
}

func do_update(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	var field_list []*mysql_base.FieldValuePair
	for i := 4; i < len(strs); i += 2 {
		field_list = append(field_list, &mysql_base.FieldValuePair{Name: strs[i], Value: strs[i+1]})
	}
	db_mgr.Update(table_name, key, value, field_list)
}

func do_delete(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	db_mgr.Delete(table_name, key, value)
}

func do_selects(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]

	var field_list []string
	if len(strs) > 4 {
		for i := 4; i < len(strs); i++ {
			field_list = append(field_list, strs[i])
		}
	}
	table := db_mgr.GetConfigLoader().GetTable(table_name)
	if table == nil {
		log.Printf("没有表%v\n", table_name)
		return
	}

	var dest_list []interface{}
	for _, field_name := range field_list {
		field := table.GetField(field_name)
		if mysql_base.IsMysqlFieldIntType(field.Type) {
			dest_list = append(dest_list, new(int))
		} else if mysql_base.IsMysqlFieldTextType(field.Type) {
			dest_list = append(dest_list, new(string))
		} else if mysql_base.IsMysqlFieldBinaryType(field.Type) || mysql_base.IsMysqlFieldBlobType(field.Type) {
			dest_list = append(dest_list, new([]byte))
		} else {
			log.Printf("不支持的select字段类型 %v\n", field.Type)
		}
	}

	log.Printf("field_list: %v, dest_list: %v\n", field_list, dest_list)

	var result_list mysql_base.QueryResultList
	if err := db_mgr.SelectRecords(table_name, key, value, field_list, &result_list); err == nil {
		log.Printf("select结果: \n")
		var cnt int
		for {
			if !result_list.Get(dest_list...) {
				log.Printf("!!!!!!!!!!!!!!\n")
				break
			}
			log.Printf("	index: %v\n", cnt)
			for i := 0; i < len(field_list); i++ {
				if len(dest_list) <= i {
					break
				}
				v := reflect.ValueOf(dest_list[i])
				e := v.Elem()
				t := e.Kind()
				if t == reflect.Int {
					log.Printf("		%v: %v\n", field_list[i], *dest_list[i].(*int))
				} else if t == reflect.String {
					log.Printf("		%v: %v\n", field_list[i], *dest_list[i].(*string))
				} else if t == reflect.Slice {
					log.Printf("		%v: %v\n", field_list[i], dest_list[i].(*[]byte))
				} else {
					log.Printf("		unprocessed reflect kind %v with index %v\n", t, i)
				}
			}
			cnt += 1
		}
		result_list.Close()
	}

}

func do_pinsert(strs []string) {
	if transaction == nil {
		log.Printf("还没有创建新事务\n")
		return
	}
	table_name := strs[1]
	var field_list []*mysql_base.FieldValuePair
	for i := 2; i < len(strs); i += 2 {
		field_list = append(field_list, &mysql_base.FieldValuePair{
			Name:  strs[i],
			Value: strs[i+1],
		})
	}
	transaction.Insert(table_name, field_list)
}

func do_pupdate(strs []string) {
	if transaction == nil {
		log.Printf("还没有创建新事务\n")
		return
	}
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	var field_list []*mysql_base.FieldValuePair
	for i := 4; i < len(strs); i += 2 {
		field_list = append(field_list, &mysql_base.FieldValuePair{Name: strs[i], Value: strs[i+1]})
	}
	transaction.Update(table_name, key, value, field_list)
}

func do_pdelete(strs []string) {
	if transaction == nil {
		log.Printf("还没有创建新事务\n")
		return
	}
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	transaction.Delete(table_name, key, value)
}

func on_tick() {
	fmt.Printf("请输入命令:\n")
	var cmd_str string
	fmt.Scanf("%s\n", &cmd_str)

	strs := strings.Split(cmd_str, ",")
	if len(strs) == 0 {
		log.Printf("命令不能为空\n")
		return
	}

	cmd := strs[0]
	if cmd == "insert" {
		if len(strs) < 4 {
			log.Printf("insert命令参数不够\n")
			return
		}
		do_insert(strs)
	} else if cmd == "select" {
		if len(strs) < 4 {
			log.Printf("select命令参数不够\n")
			return
		}
		do_select(strs)
	} else if cmd == "select_star" {
		if len(strs) < 2 {
			log.Printf("select_star命令参数不够\n")
			return
		}
		do_select_star(strs)
	} else if cmd == "update" {
		if len(strs) < 6 {
			log.Printf("update命令参数不够\n")
			return
		}
		do_update(strs)
	} else if cmd == "delete" {
		if len(strs) < 4 {
			log.Printf("delete命令参数不够\n")
			return
		}
		do_delete(strs)
	} else if cmd == "selects" {
		if len(strs) < 4 {
			return
		}
		do_selects(strs)
	} else if cmd == "save" {
		db_mgr.Save()
	} else if cmd == "new_procedure" {
		transaction = db_mgr.NewTransaction()
		log.Printf("创建了一个新的事务\n")
	} else if cmd == "commit_procedure" {
		if transaction == nil {
			log.Printf("还没有创建事务\n")
			return
		}
		transaction.Done()
		transaction = nil
		log.Printf("提交了新事务\n")
	} else if cmd == "pinsert" {
		if len(strs) < 4 {
			log.Printf("pinsert命令参数不够\n")
			return
		}
		do_pinsert(strs)
	} else if cmd == "pupdate" {
		if len(strs) < 6 {
			log.Printf("pupdate命令参数不够\n")
			return
		}
		do_pupdate(strs)
	} else if cmd == "pdelete" {
		if len(strs) < 4 {
			log.Printf("pdelete命令参数不够\n")
			return
		}
		do_pdelete(strs)
	} else {
		log.Printf("不支持的命令")
	}
}
