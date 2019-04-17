package main

import (
	"fmt"
	"log"

	"reflect"
	"strings"
	"time"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/generator"
)

var config_loader mysql_generator.ConfigLoader
var database mysql_base.Database
var db_op_manager mysql_base.DBOperateManager

func main() {

	config_path := "../src/github.com/huoshan017/mysql-go/generator/config.json"
	if !config_loader.Load(config_path) {
		log.Printf("load config %v failed\n", config_path)
		return
	}

	err := database.Open("localhost", "root", "", config_loader.DBPkg)
	if err != nil {
		log.Printf("open database err %v\n", err.Error())
		return
	}
	database.SetMaxLifeTime(time.Second * 5)
	defer database.Close()

	if config_loader.Tables != nil {
		for _, t := range config_loader.Tables {
			if !database.LoadTable(t) {
				log.Printf("load table %v config failed\n", t.Name)
				return
			}
		}
	}

	log.Printf("database loaded\n")

	db_op_manager.Init(&database)

	go func() {
		for {
			db_op_manager.Save()
			time.Sleep(time.Minute * 5)
		}
	}()

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
	db_op_manager.Insert(table_name, field_list)
}

func do_select(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	db := db_op_manager.GetDB()
	if db != nil {
		var field_list []string
		if len(strs) > 4 {
			for i := 4; i < len(strs); i++ {
				field_list = append(field_list, strs[i])
			}
		}
		table := config_loader.GetTable(table_name)
		if table == nil {
			log.Printf("没有表%v\n", table_name)
			return
		}

		var dest_list []interface{}
		for _, field_name := range field_list {
			field := table.GetField(field_name)
			if mysql_base.IsMysqlFieldIntType(field.RealType) {
				dest_list = append(dest_list, new(int))
			} else if mysql_base.IsMysqlFieldTextType(field.RealType) {
				dest_list = append(dest_list, new(string))
			} else if mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType) {
				dest_list = append(dest_list, new([]byte))
			} else {
				log.Printf("不支持的select字段类型 %v\n", field.RealType)
			}
		}
		if db.SelectRecord(table_name, key, value, field_list, dest_list) {
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
				} else if t == reflect.Array {
					log.Printf("		%v: %v\n", field_list[i], dest_list[i].([]byte))
				} else {
					log.Printf("		unprocessed reflect kind %v with index %v\n", t, i)
				}
			}
		}
	}
}

func do_update(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	var field_list []*mysql_base.FieldValuePair
	for i := 4; i < len(strs); i += 2 {
		field_list = append(field_list, &mysql_base.FieldValuePair{strs[i], strs[i+1]})
	}
	db_op_manager.Update(table_name, key, value, field_list)
}

func do_delete(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	db_op_manager.Delete(table_name, key, value)
}

func do_selects(strs []string) {
	table_name := strs[1]
	key := strs[2]
	value := strs[3]
	db := db_op_manager.GetDB()
	if db != nil {
		var field_list []string
		if len(strs) > 4 {
			for i := 4; i < len(strs); i++ {
				field_list = append(field_list, strs[i])
			}
		}
		table := config_loader.GetTable(table_name)
		if table == nil {
			log.Printf("没有表%v\n", table_name)
			return
		}

		var dest_list []interface{}
		for _, field_name := range field_list {
			field := table.GetField(field_name)
			if mysql_base.IsMysqlFieldIntType(field.RealType) {
				dest_list = append(dest_list, new(int))
			} else if mysql_base.IsMysqlFieldTextType(field.RealType) {
				dest_list = append(dest_list, new(string))
			} else if mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType) {
				dest_list = append(dest_list, new([]byte))
			} else {
				log.Printf("不支持的select字段类型 %v\n", field.RealType)
			}
		}

		log.Printf("field_list: %v, dest_list: %v\n", field_list, dest_list)

		var result_list mysql_base.QueryResultList
		if db.SelectRecords(table_name, key, value, field_list, &result_list) {
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
					} else if t == reflect.Array {
						log.Printf("		%v: %v\n", field_list[i], dest_list[i].([]byte))
					}
				}
				cnt += 1
			}
			result_list.Close()
		}
	}
}

func on_tick() {
	fmt.Printf("请输入命令:\n")
	var cmd_str string
	fmt.Scanf("%s\n", &cmd_str)

	strs := strings.Split(cmd_str, ",")
	if strs == nil || len(strs) == 0 {
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
		db_op_manager.Save()
	} else {
		log.Printf("不支持的命令")
	}
}
