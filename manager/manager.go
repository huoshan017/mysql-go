package mysql_manager

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/generator"
)

const (
	DEFAULT_CONN_MAX_LIFE_SECONDS = time.Second * 5
	DEFAULT_SAVE_INTERVAL_TIME    = time.Minute * 5
)

const (
	DB_STATE_NO_RUN  = iota
	DB_STATE_RUNNING = 1
	DB_STATE_TO_END  = 2
)

type DBManager struct {
	config_loader mysql_generator.ConfigLoader
	database      mysql_base.Database
	db_op_manager mysql_base.DBOperateManager
	save_interval time.Duration
	state         int32
}

func (this *DBManager) LoadConfig(config_path string) bool {
	if !this.config_loader.Load(config_path) {
		log.Printf("load config %v failed\n", config_path)
		return false
	}
	return true
}

func (this *DBManager) Connect(dbhost, dbuser, dbpassword, dbname string) bool {
	err := this.database.Open(dbhost, dbuser, dbpassword, this.config_loader.DBPkg)
	if err != nil {
		log.Printf("open database err %v\n", err.Error())
		return false
	}
	this.database.SetMaxLifeTime(DEFAULT_CONN_MAX_LIFE_SECONDS)
	if this.config_loader.Tables != nil {
		for _, t := range this.config_loader.Tables {
			if !this.database.LoadTable(t) {
				log.Printf("load table %v config failed\n", t.Name)
				return false
			}
		}
	}
	this.db_op_manager.Init(&this.database)
	this.save_interval = DEFAULT_SAVE_INTERVAL_TIME
	this.state = DB_STATE_RUNNING
	return true
}

func (this *DBManager) SetConnectLifeTime(d time.Duration) {
	this.database.SetMaxLifeTime(d)
}

func (this *DBManager) SetSaveIntervalTime(d time.Duration) {
	this.save_interval = d
}

func (this *DBManager) GetConfigLoader() *mysql_generator.ConfigLoader {
	return &this.config_loader
}

func (this *DBManager) Close() {
	this.database.Close()
}

func (this *DBManager) ToEnd() bool {
	return atomic.CompareAndSwapInt32(&this.state, DB_STATE_RUNNING, DB_STATE_TO_END)
}

func (this *DBManager) Save() {
	this.db_op_manager.Save()
}

func (this *DBManager) Run() {
	go func() {
		var last_save_time int32
		for {
			if atomic.CompareAndSwapInt32(&this.state, DB_STATE_TO_END, DB_STATE_NO_RUN) {
				break
			}

			if last_save_time > 0 && int32(time.Now().Unix())-last_save_time >= int32(this.save_interval) {
				this.db_op_manager.Save()
			}
			time.Sleep(time.Second)
		}
	}()
}

func (this *DBManager) Insert(table_name string, field_pair []*mysql_base.FieldValuePair) {
	this.db_op_manager.Insert(table_name, field_pair)
}

func (this *DBManager) Update(table_name string, key string, value interface{}, field_pair []*mysql_base.FieldValuePair) {
	this.db_op_manager.Update(table_name, key, value, field_pair)
}

func (this *DBManager) Delete(table_name string, key string, value interface{}) {
	this.db_op_manager.Delete(table_name, key, value)
}

func (this *DBManager) AppendProcedureOpList(procedure_op_list *mysql_base.ProcedureOpList) {
	this.db_op_manager.AppendProcedure(procedure_op_list)
}

func (this *DBManager) Select(table_name string, key string, value interface{}, field_list []string, dest_list []interface{}) bool {
	return this.database.SelectRecord(table_name, key, value, field_list, dest_list)
}

func (this *DBManager) SelectStar(table_name string, key string, value interface{}, dest_list []interface{}) bool {
	return this.database.SelectRecord(table_name, key, value, nil, dest_list)
}

func (this *DBManager) SelectRecords(table_name string, key string, value interface{}, field_list []string, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, key, value, field_list, result_list)
}

func (this *DBManager) SelectStarRecords(table_name string, key string, value interface{}, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, key, value, nil, result_list)
}

func (this *DBManager) SelectAllRecords(table_name string, field_list []string, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, "", nil, field_list, result_list)
}

func (this *DBManager) SelectStarAllRecords(table_name string, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, "", nil, nil, result_list)
}
