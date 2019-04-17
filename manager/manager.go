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
