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

type DB struct {
	config_loader mysql_generator.ConfigLoader
	database      mysql_base.Database
	db_op_manager mysql_base.DBOperateManager
	save_interval time.Duration
	state         int32
}

func (this *DB) LoadConfig(config_path string) bool {
	if !this.config_loader.Load(config_path) {
		log.Printf("load config %v failed\n", config_path)
		return false
	}
	return true
}

func (this *DB) Connect(dbhost, dbuser, dbpassword, dbname string) bool {
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

func (this *DB) SetConnectLifeTime(d time.Duration) {
	this.database.SetMaxLifeTime(d)
}

func (this *DB) SetSaveIntervalTime(d time.Duration) {
	this.save_interval = d
}

func (this *DB) GetConfigLoader() *mysql_generator.ConfigLoader {
	return &this.config_loader
}

func (this *DB) Insert(table_name string, field_pair []*mysql_base.FieldValuePair) {
	this.db_op_manager.Insert(table_name, field_pair)
}

func (this *DB) Update(table_name string, key string, value interface{}, field_pair []*mysql_base.FieldValuePair) {
	this.db_op_manager.Update(table_name, key, value, field_pair)
}

func (this *DB) Delete(table_name string, key string, value interface{}) {
	this.db_op_manager.Delete(table_name, key, value)
}

func (this *DB) AppendProcedureOpList(procedure_op_list *mysql_base.ProcedureOpList) {
	this.db_op_manager.AppendProcedure(procedure_op_list)
}

func (this *DB) Select(table_name string, key string, value interface{}, field_list []string, dest_list []interface{}) bool {
	return this.database.SelectRecord(table_name, key, value, field_list, dest_list)
}

func (this *DB) SelectStar(table_name string, key string, value interface{}, dest_list []interface{}) bool {
	return this.database.SelectRecord(table_name, key, value, nil, dest_list)
}

func (this *DB) SelectRecords(table_name string, key string, value interface{}, field_list []string, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, key, value, field_list, result_list)
}

func (this *DB) SelectStarRecords(table_name string, key string, value interface{}, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, key, value, nil, result_list)
}

func (this *DB) SelectAllRecords(table_name string, field_list []string, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, "", nil, field_list, result_list)
}

func (this *DB) SelectStarAllRecords(table_name string, result_list *mysql_base.QueryResultList) bool {
	return this.database.SelectRecords(table_name, "", nil, nil, result_list)
}

func (this *DB) Close() {
	this.database.Close()
}

func (this *DB) ToEnd() bool {
	return atomic.CompareAndSwapInt32(&this.state, DB_STATE_RUNNING, DB_STATE_TO_END)
}

func (this *DB) Save() {
	this.db_op_manager.Save()
}

func (this *DB) Run() {
	go func() {
		var last_save_time int32
		for {
			if atomic.CompareAndSwapInt32(&this.state, DB_STATE_TO_END, DB_STATE_NO_RUN) {
				break
			}

			now_time := int32(time.Now().Unix())
			if last_save_time == 0 {
				last_save_time = now_time
			} else if last_save_time > 0 && now_time-last_save_time >= int32(this.save_interval.Seconds()) {
				this.db_op_manager.Save()
				last_save_time = now_time
			}
			time.Sleep(time.Second)
		}
	}()
}
