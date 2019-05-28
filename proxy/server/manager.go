package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/manager"
	"github.com/huoshan017/mysql-go/proxy/common"
)

type new_value_func func() interface{}

var type2new_value = map[string]new_value_func{
	"bool": func() interface{} {
		return new(bool)
	},
	"int8": func() interface{} {
		return new(int8)
	},
	"int16": func() interface{} {
		return new(int16)
	},
	"int32": func() interface{} {
		return new(int32)
	},
	"int64": func() interface{} {
		return new(int64)
	},
	"uint8": func() interface{} {
		return new(uint8)
	},
	"uint16": func() interface{} {
		return new(uint16)
	},
	"uint32": func() interface{} {
		return new(uint32)
	},
	"uint64": func() interface{} {
		return new(uint64)
	},
	"float32": func() interface{} {
		return new(float32)
	},
	"float64": func() interface{} {
		return new(float64)
	},
	"string": func() interface{} {
		return new(string)
	},
	"byte": func() interface{} {
		return new(byte)
	},
	"[]byte": func() interface{} {
		return &[]byte{}
	},
}

type ProxyReadProc struct {
}

func _get_db(head *mysql_proxy_common.ArgsHead) (db *mysql_manager.DB, err error) {
	host_id := head.GetDBHostId()
	db_name := head.GetDBName()
	db = db_list.GetDB(host_id, db_name)
	if db == nil {
		err = errors.New(fmt.Sprintf("mysql-proxy-server: not found db with host_id(%v) and db_name(%v)", head.GetDBHostId(), head.GetDBName()))
		return
	}
	return
}

func _get_db_and_table_config(head *mysql_proxy_common.ArgsHead, table_name string) (db *mysql_manager.DB, table_config *mysql_base.TableConfig, err error) {
	db, err = _get_db(head)
	if err != nil {
		return
	}
	config_loader := db.GetConfigLoader()
	if config_loader == nil {
		err = errors.New(fmt.Sprintf("mysql-proxy-server: db host_id(%v) db_name(%v) not get config loader", head.GetDBHostId(), head.GetDBName()))
		return
	}
	table_config = config_loader.GetTable(table_name)
	if table_config == nil {
		err = errors.New(fmt.Sprintf("mysql-proxy-server: db host_id(%v) db_name(%v) not found table name %v", head.GetDBHostId(), head.GetDBName(), table_name))
		return
	}
	return
}

func _get_new_value_with_field_name(table_config *mysql_base.TableConfig, field_name string) (new_value interface{}, err error) {
	fc := table_config.GetField(field_name)
	if fc == nil {
		err = errors.New(fmt.Sprintf("mysql-proxy-server: get table %v field %v not found", table_config.Name, field_name))
		return
	}
	go_type := mysql_base.MysqlFieldType2GoTypeStr(fc.RealType)
	if go_type == "" {
		err = errors.New(fmt.Sprintf("mysql-proxy-server: table %v field %v type %v transfer to go type failed", table_config.Name, field_name, fc.Type))
		return
	}
	new_value_func := type2new_value[go_type]
	if new_value_func == nil {
		err = errors.New(fmt.Sprintf("mysql-proxy-server: table %v field %v type %v transfer to go type %v not get new value func", table_config.Name, field_name, fc.Type, go_type))
	}
	new_value = new_value_func()
	return
}

func _make_dest_list_with_field_names(table_config *mysql_base.TableConfig, field_names []string) (dest_list []interface{}, err error) {
	dest_list = make([]interface{}, len(field_names))
	for i, fn := range field_names {
		var new_value interface{}
		new_value, err = _get_new_value_with_field_name(table_config, fn)
		if err != nil {
			return
		}
		if new_value == nil {
			err = errors.New(fmt.Sprintf("mysql-proxy-server: table %v field %v cant get new value", table_config.Name, fn))
			return
		}
		dest_list[i] = new_value
	}
	return
}

func _gen_dest_lists(result_list *mysql_base.QueryResultList, table_config *mysql_base.TableConfig, select_field_names []string) (dest_lists [][]interface{}, err error) {
	for {
		var dest_list []interface{}
		dest_list, err = _make_dest_list_with_field_names(table_config, select_field_names)
		if err != nil {
			return
		}
		if !result_list.Get(dest_list...) {
			break
		}
		dest_lists = append(dest_lists, dest_list)
	}
	return
}

func output_critical(err interface{}) {
	mysql_proxy_common.OutputCriticalStack(mysql_proxy_common.ServerLogErr, err)
}

func (this *ProxyReadProc) Select(args *mysql_proxy_common.SelectArgs, reply *mysql_proxy_common.SelectReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, table_config, err := _get_db_and_table_config(args.Head, args.TableName)
	if err != nil {
		return err
	}
	var dest_list []interface{}
	dest_list, err = _make_dest_list_with_field_names(table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	if !db.Select(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelectFieldNames, dest_list) {
		return errors.New(fmt.Sprintf("mysql-proxy-server: select with table_name(%v) where_field_name(%v) select_field_names(%v) failed", args.TableName, args.WhereFieldName, args.SelectFieldNames))
	}
	reply.Result = dest_list
	return nil
}

func (this *ProxyReadProc) SelectRecords(args *mysql_proxy_common.SelectRecordsArgs, reply *mysql_proxy_common.SelectRecordsReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, table_config, err := _get_db_and_table_config(args.Head, args.TableName)
	if err != nil {
		return err
	}
	var result_list mysql_base.QueryResultList
	if !db.SelectRecords(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelectFieldNames, &result_list) {
		return errors.New(fmt.Sprintf("mysql-proxy-server: select records with table_name(%v) where_field_name(%v) select_field_names(%v) failed", args.TableName, args.WhereFieldName, args.SelectFieldNames))
	}
	var dest_lists [][]interface{}
	dest_lists, err = _gen_dest_lists(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultList = dest_lists
	return nil
}

func (this *ProxyReadProc) SelectAllRecords(args *mysql_proxy_common.SelectAllRecordsArgs, reply *mysql_proxy_common.SelectAllRecordsReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, table_config, err := _get_db_and_table_config(args.Head, args.TableName)
	if err != nil {
		return err
	}
	var result_list mysql_base.QueryResultList
	if !db.SelectAllRecords(args.TableName, args.SelectFieldNames, &result_list) {
		return errors.New(fmt.Sprintf("mysql-proxy-server: select all records with table_name(%v) select_field_names(%v) failed", args.TableName, args.SelectFieldNames))
	}
	var dest_lists [][]interface{}
	dest_lists, err = _gen_dest_lists(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultList = dest_lists
	return nil
}

func (this *ProxyReadProc) SelectField(args *mysql_proxy_common.SelectFieldArgs, reply *mysql_proxy_common.SelectFieldReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, table_config, err := _get_db_and_table_config(args.Head, args.TableName)
	if err != nil {
		return err
	}
	var result_list mysql_base.QueryResultList
	if !db.SelectFieldNoKey(args.TableName, args.SelectFieldName, &result_list) {
		return errors.New(fmt.Sprintf("mysql-proxy-server: select field with table_name(%v) select_field_name(%v) failed", args.TableName, args.SelectFieldName))
	}
	var dest_list []interface{}
	for {
		var new_value interface{}
		new_value, err = _get_new_value_with_field_name(table_config, args.SelectFieldName)
		if !result_list.Get(new_value) {
			break
		}
		dest_list = append(dest_list, new_value)
	}
	reply.ResultList = dest_list
	return nil
}

func (this *ProxyReadProc) SelectRecordsCondition(args *mysql_proxy_common.SelectRecordsConditionArgs, reply *mysql_proxy_common.SelectRecordsConditionReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, table_config, err := _get_db_and_table_config(args.Head, args.TableName)
	if err != nil {
		return err
	}
	var result_list mysql_base.QueryResultList
	if !db.SelectRecordsCondition(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelCond, args.SelectFieldNames, &result_list) {
		return errors.New(fmt.Sprintf("mysql-proxy-server: select records order by with table_name(%v) where_field_name(%v) sel_cond(%v) select_field_names(%v) failed", args.TableName, args.WhereFieldName, args.SelCond, args.SelectFieldNames))
	}
	var dest_lists [][]interface{}
	dest_lists, err = _gen_dest_lists(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultList = dest_lists
	return nil
}

type ProxyWriteProc struct {
}

func (this *ProxyWriteProc) InsertRecord(args *mysql_proxy_common.InsertRecordArgs, reply *mysql_proxy_common.InsertRecordReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, err := _get_db(args.Head)
	if err != nil {
		return err
	}
	if args.Ignore {
		db.InsertIgnore(args.TableName, args.FieldValuePairs)
	} else {
		db.Insert(args.TableName, args.FieldValuePairs)
	}
	if IsDebug {
		log.Printf("ProxyWriteProc.InsertRecord: table_name(%v) field_pairs(%v)\n", args.TableName, args.FieldValuePairs)
	}
	return nil
}

func (this *ProxyWriteProc) UpdateRecord(args *mysql_proxy_common.UpdateRecordArgs, reply *mysql_proxy_common.UpdateRecordReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, err := _get_db(args.Head)
	if err != nil {
		return err
	}
	db.Update(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.FieldValuePairs)
	if IsDebug {
		log.Printf("ProxyWriteProc.UpdateRecord: table_name(%v) where_field_name(%v) where_field_value(%v) field_pairs(%v)\n", args.TableName, args.WhereFieldName, args.WhereFieldValue, args.FieldValuePairs)
	}
	return nil
}

func (this *ProxyWriteProc) DeleteRecord(args *mysql_proxy_common.DeleteRecordArgs, reply *mysql_proxy_common.DeleteRecordReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, err := _get_db(args.Head)
	if err != nil {
		return err
	}
	db.Delete(args.TableName, args.WhereFieldName, args.WhereFieldValue)
	if IsDebug {
		log.Printf("ProxyWriteProc.DeleteRecord: table_name(%v) where_field_name(%v) where_field_value(%v)\n", args.TableName, args.WhereFieldName, args.WhereFieldValue)
	}
	return nil
}

func (this *ProxyWriteProc) CommitTransaction(args *mysql_proxy_common.CommitTransactionArgs, reply *mysql_proxy_common.CommitTransactionReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, err := _get_db(args.Head)
	if err != nil {
		return err
	}
	transaction := db.NewTransaction()
	transaction.SetDetailList(args.Details)
	transaction.Done()
	if IsDebug {
		log.Printf("ProxyWriteProc.CommitTransaction: %v\n", args.Details)
	}
	return nil
}

func (this *ProxyWriteProc) Save(args *mysql_proxy_common.SaveImmidiateArgs, reply *mysql_proxy_common.SaveImmidiateReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	db, err := _get_db(args.Head)
	if err != nil {
		return err
	}
	db.Save()
	return nil
}

type ProcService struct {
	service *Service
}

func (this *ProcService) init() {
	this.service = &Service{}
	this.service.Init()
	this.service.Register(&ProxyReadProc{})
	this.service.Register(&ProxyWriteProc{})
	RegisterUserType(&mysql_base.FieldValuePair{})
	RegisterUserType(&mysql_base.OpDetail{})
	RegisterUserType(&mysql_base.SelectCondition{})
}

func (this *ProcService) Start(addr string) error {
	this.init()
	err := this.service.Listen(addr)
	if err != nil {
		return errors.New(fmt.Sprintf("mysql-proxy-server: start with addr %v err: %v", addr, err.Error()))
	}
	this.service.Serve()
	return nil
}

var (
	IsDebug = false
)

func SetDebug(debug bool) {
	IsDebug = debug
}
