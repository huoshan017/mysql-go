package mysql_manager

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/generate"
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
	config_loader *mysql_generate.ConfigLoader
	database      mysql_base.Database
	op_mgr        OperateManager
	save_interval time.Duration
	state         int32
}

func (this *DB) LoadConfig(config_path string) bool {
	config_loader := &mysql_generate.ConfigLoader{}
	if !config_loader.Load(config_path) {
		log.Printf("load config %v failed\n", config_path)
		return false
	}
	this.config_loader = config_loader
	return true
}

func (this *DB) AttachConfig(config_loader *mysql_generate.ConfigLoader) {
	this.config_loader = config_loader
}

func (this *DB) GetConfigLoader() *mysql_generate.ConfigLoader {
	return this.config_loader
}

func (this *DB) Connect(dbhost, dbuser, dbpassword, dbname string) error {
	err := this.database.Open(dbhost, dbuser, dbpassword, dbname)
	if err != nil {
		log.Printf("open database err %v\n", err.Error())
		return err
	}
	this.database.SetMaxLifeTime(DEFAULT_CONN_MAX_LIFE_SECONDS)
	if this.config_loader.Tables != nil {
		for _, t := range this.config_loader.Tables {
			err = this.database.LoadTable(t)
			if err != nil {
				log.Printf("load table %v config failed\n", t.Name)
				return err
			}
		}
	}
	this.op_mgr.Init(&this.database, this.config_loader)
	this.save_interval = DEFAULT_SAVE_INTERVAL_TIME
	return nil
}

func (this *DB) SetConnectLifeTime(d time.Duration) {
	this.database.SetMaxLifeTime(d)
}

func (this *DB) SetSaveIntervalTime(d time.Duration) {
	this.save_interval = d
}

func (this *DB) Insert(table_name string, field_pair []*mysql_base.FieldValuePair) {
	this.op_mgr.Insert(table_name, field_pair, false)
}

func (this *DB) InsertIgnore(table_name string, field_pair []*mysql_base.FieldValuePair) {
	this.op_mgr.Insert(table_name, field_pair, true)
}

func (this *DB) Update(table_name string, field_name string, field_value interface{}, field_pair []*mysql_base.FieldValuePair) {
	this.op_mgr.Update(table_name, field_name, field_value, field_pair)
}

func (this *DB) Delete(table_name string, field_name string, field_value interface{}) {
	this.op_mgr.Delete(table_name, field_name, field_value)
}

func (this *DB) SelectUseSql(query_sql string, result_list *mysql_base.QueryResultList) error {
	return this.database.Query(query_sql, result_list)
}

func (this *DB) SelectRecordsCount(table_name string) (count int32, err error) {
	return this.database.SelectRecordsCount(table_name, "", nil)
}

func (this *DB) SelectRecordsCountByField(table_name, field_name string, field_value interface{}) (count int32, err error) {
	return this.database.SelectRecordsCount(table_name, field_name, field_value)
}

func (this *DB) Select(table_name string, field_name string, field_value interface{}, field_list []string, dest_list []interface{}) error {
	return this.database.SelectRecord(table_name, field_name, field_value, field_list, dest_list)
}

func (this *DB) SelectStar(table_name string, field_name string, field_value interface{}, dest_list []interface{}) error {
	return this.database.SelectRecord(table_name, field_name, field_value, nil, dest_list)
}

func (this *DB) SelectRecords(table_name string, field_name string, field_value interface{}, field_list []string, result_list *mysql_base.QueryResultList) error {
	return this.database.SelectRecords(table_name, field_name, field_value, field_list, result_list)
}

func (this *DB) SelectStarRecords(table_name string, field_name string, field_value interface{}, result_list *mysql_base.QueryResultList) error {
	return this.database.SelectRecords(table_name, field_name, field_value, nil, result_list)
}

func (this *DB) SelectAllRecords(table_name string, field_list []string, result_list *mysql_base.QueryResultList) error {
	return this.database.SelectRecords(table_name, "", nil, field_list, result_list)
}

func (this *DB) SelectStarAllRecords(table_name string, result_list *mysql_base.QueryResultList) error {
	return this.database.SelectRecords(table_name, "", nil, nil, result_list)
}

func (this *DB) SelectFieldNoKey(table_name string, field_name string, result_list *mysql_base.QueryResultList) error {
	return this.database.SelectRecords(table_name, "", nil, []string{field_name}, result_list)
}

func (this *DB) SelectRecordsCondition(table_name string, field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition, field_list []string, result_list *mysql_base.QueryResultList) error {
	return this.database.SelectRecordsCondition(table_name, field_name, field_value, sel_cond, field_list, result_list)
}

func (this *DB) NewTransaction() *Transaction {
	return this.op_mgr.NewTransaction()
}

func (this *DB) Close() {
	this.database.Close()
}

func (this *DB) EndRun() bool {
	return atomic.CompareAndSwapInt32(&this.state, DB_STATE_RUNNING, DB_STATE_TO_END)
}

func (this *DB) IsEnd() bool {
	return atomic.LoadInt32(&this.state) == DB_STATE_NO_RUN
}

func (this *DB) Save() {
	this.op_mgr.Save()
}

func (this *DB) Run() {
	this.state = DB_STATE_RUNNING
	var last_save_time int32
	for {
		if atomic.CompareAndSwapInt32(&this.state, DB_STATE_TO_END, DB_STATE_NO_RUN) {
			break
		}

		now_time := int32(time.Now().Unix())
		if last_save_time == 0 {
			last_save_time = now_time
		} else if last_save_time > 0 && now_time-last_save_time >= int32(this.save_interval.Seconds()) {
			this.op_mgr.Save()
			last_save_time = now_time
		}
		time.Sleep(time.Second)
	}
	this.database.Close()
}

func (this *DB) GoRun() {
	go this.Run()
}
