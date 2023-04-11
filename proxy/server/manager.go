package main

import (
	"fmt"
	"log"
	"strings"

	mysql_base "github.com/huoshan017/mysql-go/base"
	mysql_manager "github.com/huoshan017/mysql-go/manager"
	mysql_proxy_common "github.com/huoshan017/mysql-go/proxy/common"
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

func _get_db(head *mysql_proxy_common.ArgsHead) (db *mysql_manager.DB, err error) {
	host_id := head.GetDBHostId()
	db_name := head.GetDBName()
	db = db_list.GetDB(host_id, db_name)
	if db == nil {
		err = fmt.Errorf("mysql-proxy-server: not found db with host_id(%v) and db_name(%v)", head.GetDBHostId(), head.GetDBName())
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
		err = fmt.Errorf("mysql-proxy-server: db host_id(%v) db_name(%v) not get config loader", head.GetDBHostId(), head.GetDBName())
		return
	}
	table_config = config_loader.GetTable(table_name)
	if table_config == nil {
		err = fmt.Errorf("mysql-proxy-server: db host_id(%v) db_name(%v) not found table name %v", head.GetDBHostId(), head.GetDBName(), table_name)
		return
	}
	return
}

func _get_new_value_with_field_name(table_config *mysql_base.TableConfig, field_name string) (new_value interface{}, err error) {
	fc := table_config.GetField(field_name)
	if fc == nil {
		err = fmt.Errorf("mysql-proxy-server: get table %v field %v not found", table_config.Name, field_name)
		return
	}
	is_unsigned := strings.Contains(strings.ToLower(fc.TypeStr), "unsigned")
	go_type := mysql_base.MysqlFieldType2GoTypeStr(fc.Type, is_unsigned)
	if go_type == "" {
		err = fmt.Errorf("mysql-proxy-server: table %v field %v type %v transfer to go type failed", table_config.Name, field_name, fc.Type)
		return
	}
	new_value_func := type2new_value[go_type]
	if new_value_func == nil {
		err = fmt.Errorf("mysql-proxy-server: table %v field %v type %v transfer to go type %v not get new value func", table_config.Name, field_name, fc.Type, go_type)
	}
	new_value = new_value_func()
	return
}

func _make_dest_list_with_field_names(table_config *mysql_base.TableConfig, field_names []string) (dest_list []interface{}, primary_value interface{}, err error) {
	dest_list = make([]interface{}, len(field_names))
	for i, fn := range field_names {
		var new_value interface{}
		new_value, err = _get_new_value_with_field_name(table_config, fn)
		if err != nil {
			return
		}
		if new_value == nil {
			err = fmt.Errorf("mysql-proxy-server: table %v field %v cant get new value", table_config.Name, fn)
			return
		}
		dest_list[i] = new_value
		if fn == table_config.PrimaryKey {
			primary_value = new_value
		}
	}
	return
}

func _gen_dest_lists(result_list *mysql_base.QueryResultList, table_config *mysql_base.TableConfig, select_field_names []string) (dest_lists [][]interface{}, err error) {
	for {
		var dest_list []interface{}
		dest_list, _, err = _make_dest_list_with_field_names(table_config, select_field_names)
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

func _gen_dest_list_map(result_list *mysql_base.QueryResultList, table_config *mysql_base.TableConfig, select_field_names []string) (dest_list_map map[interface{}][]interface{}, err error) {
	for {
		var dest_list []interface{}
		var primary_value interface{}
		dest_list, primary_value, err = _make_dest_list_with_field_names(table_config, select_field_names)
		if err != nil {
			return
		}
		if !result_list.Get(dest_list...) {
			break
		}
		if dest_list_map == nil {
			dest_list_map = make(map[interface{}][]interface{})
		}
		dest_list_map[primary_value] = dest_list
	}
	return
}

func output_critical(err interface{}) {
	mysql_proxy_common.OutputCriticalStack(mysql_proxy_common.ServerLogErr, err)
}

type ProxyReadProc struct {
}

func (p *ProxyReadProc) Select(args *mysql_proxy_common.SelectArgs, reply *mysql_proxy_common.SelectReply) error {
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
	dest_list, _, err = _make_dest_list_with_field_names(table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	err = db.Select(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelectFieldNames, dest_list)
	if err != nil {
		return err
	}
	reply.Result = dest_list
	return nil
}

func (p *ProxyReadProc) SelectRecordsCount(args *mysql_proxy_common.SelectRecordsCountArgs, reply *mysql_proxy_common.SelectRecordsCountReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()

	db, err := _get_db(args.Head)
	if err != nil {
		return err
	}

	var count int32
	if args.WhereFieldName == "" {
		count, err = db.SelectRecordsCount(args.TableName)
	} else {
		count, err = db.SelectRecordsCountByField(args.TableName, args.WhereFieldName, args.WhereFieldValue)
	}

	if err != nil {
		reply.Count = count
	}

	return err
}

func (p *ProxyReadProc) SelectRecords(args *mysql_proxy_common.SelectRecordsArgs, reply *mysql_proxy_common.SelectRecordsReply) error {
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
	err = db.SelectRecords(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelectFieldNames, &result_list)
	if err != nil {
		return err
	}
	var dest_lists [][]interface{}
	dest_lists, err = _gen_dest_lists(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultList = dest_lists
	return nil
}

func (p *ProxyReadProc) SelectRecordsMap(args *mysql_proxy_common.SelectRecordsArgs, reply *mysql_proxy_common.SelectRecordsMapReply) error {
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
	err = db.SelectRecords(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelectFieldNames, &result_list)
	if err != nil {
		return err
	}
	var dest_list_map map[interface{}][]interface{}
	dest_list_map, err = _gen_dest_list_map(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultMap = dest_list_map
	return nil
}

func (p *ProxyReadProc) SelectAllRecords(args *mysql_proxy_common.SelectAllRecordsArgs, reply *mysql_proxy_common.SelectAllRecordsReply) error {
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
	err = db.SelectAllRecords(args.TableName, args.SelectFieldNames, &result_list)
	if err != nil {
		return err
	}
	var dest_lists [][]interface{}
	dest_lists, err = _gen_dest_lists(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultList = dest_lists
	return nil
}

func (p *ProxyReadProc) SelectAllRecordsMap(args *mysql_proxy_common.SelectAllRecordsArgs, reply *mysql_proxy_common.SelectAllRecordsMapReply) error {
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
	err = db.SelectAllRecords(args.TableName, args.SelectFieldNames, &result_list)
	if err != nil {
		return err
	}
	var dest_list_map map[interface{}][]interface{}
	dest_list_map, err = _gen_dest_list_map(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultMap = dest_list_map
	return nil
}

func (p *ProxyReadProc) SelectField(args *mysql_proxy_common.SelectFieldArgs, reply *mysql_proxy_common.SelectFieldReply) error {
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
	err = db.SelectFieldNoKey(args.TableName, args.SelectFieldName, &result_list)
	if err != nil {
		return err
	}
	var dest_list []interface{}
	for {
		var new_value interface{}
		new_value, err = _get_new_value_with_field_name(table_config, args.SelectFieldName)
		if err != nil {
			continue
		}
		if !result_list.Get(new_value) {
			break
		}
		dest_list = append(dest_list, new_value)
	}
	reply.ResultList = dest_list
	return nil
}

func (p *ProxyReadProc) SelectFieldMap(args *mysql_proxy_common.SelectFieldArgs, reply *mysql_proxy_common.SelectFieldMapReply) error {
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
	err = db.SelectFieldNoKey(args.TableName, args.SelectFieldName, &result_list)
	if err != nil {
		return err
	}
	var dest_map = make(map[interface{}]bool)
	for {
		var new_value interface{}
		new_value, err = _get_new_value_with_field_name(table_config, args.SelectFieldName)
		if err != nil {
			continue
		}
		if !result_list.Get(new_value) {
			break
		}
		dest_map[new_value] = true
	}
	reply.ResultMap = dest_map
	return nil
}

func (p *ProxyReadProc) SelectRecordsCondition(args *mysql_proxy_common.SelectRecordsConditionArgs, reply *mysql_proxy_common.SelectRecordsConditionReply) error {
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
	err = db.SelectRecordsCondition(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelCond, args.SelectFieldNames, &result_list)
	if err != nil {
		return err
	}
	var dest_lists [][]interface{}
	dest_lists, err = _gen_dest_lists(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultList = dest_lists
	return nil
}

func (p *ProxyReadProc) SelectRecordsMapCondition(args *mysql_proxy_common.SelectRecordsConditionArgs, reply *mysql_proxy_common.SelectRecordsMapConditionReply) error {
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
	err = db.SelectRecordsCondition(args.TableName, args.WhereFieldName, args.WhereFieldValue, args.SelCond, args.SelectFieldNames, &result_list)
	if err != nil {
		return err
	}
	var dest_list_map map[interface{}][]interface{}
	dest_list_map, err = _gen_dest_list_map(&result_list, table_config, args.SelectFieldNames)
	if err != nil {
		return err
	}
	reply.ResultMap = dest_list_map
	return nil
}

type ProxyWriteProc struct {
}

func (p *ProxyWriteProc) InsertRecord(args *mysql_proxy_common.InsertRecordArgs, reply *mysql_proxy_common.InsertRecordReply) error {
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

func (p *ProxyWriteProc) UpdateRecord(args *mysql_proxy_common.UpdateRecordArgs, reply *mysql_proxy_common.UpdateRecordReply) error {
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

func (p *ProxyWriteProc) DeleteRecord(args *mysql_proxy_common.DeleteRecordArgs, reply *mysql_proxy_common.DeleteRecordReply) error {
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

func (p *ProxyWriteProc) CommitTransaction(args *mysql_proxy_common.CommitTransactionArgs, reply *mysql_proxy_common.CommitTransactionReply) error {
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

func (p *ProxyWriteProc) Save(args *mysql_proxy_common.SaveImmidiateArgs, reply *mysql_proxy_common.SaveImmidiateReply) error {
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

func (p *ProxyWriteProc) End(args *mysql_proxy_common.EndArgs, reply *mysql_proxy_common.EndReply) error {
	defer func() {
		if err := recover(); err != nil {
			output_critical(err)
		}
	}()
	return nil
}

type ProcService struct {
	service *Service
}

func (p *ProcService) init() {
	p.service = &Service{}
	p.service.Init()
	p.service.Register(&ProxyReadProc{})
	p.service.Register(&ProxyWriteProc{})
	RegisterUserType(&mysql_base.FieldValuePair{})
	RegisterUserType(&mysql_base.OpDetail{})
	RegisterUserType(&mysql_base.SelectCondition{})
}

func (p *ProcService) Start(addr string) error {
	p.init()
	err := p.service.Listen(addr)
	if err != nil {
		return err
	}
	p.service.Serve()
	return nil
}

var (
	IsDebug = false
)

func SetDebug(debug bool) {
	IsDebug = debug
}
