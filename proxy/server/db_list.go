package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/huoshan017/mysql-go/generate"
	"github.com/huoshan017/mysql-go/manager"
)

type DbDefine struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
}

type Db struct {
	ConfigId int32    `json:"config_id"`
	Name     string   `json:"name"`
	NameList []string `json:"name_list"`
	Disable  bool     `json:"enable"`
}

type DbHost struct {
	Id       int32  `json:"id"`
	Enable   bool   `json:"enable"`
	Ip       string `json:"ip"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbList   []*Db  `json:"db_list"`
}

type DbList struct {
	DefineList []*DbDefine `json:"config_list"`
	MysqlHosts []*DbHost   `json:"mysql_hosts"`

	config_loaders       map[int32]*mysql_generate.ConfigLoader
	db_mgr_list          map[int32]map[string]*mysql_manager.DB
	db_mgr_list_by_alias map[string]map[string]*mysql_manager.DB
}

func (this *DbList) Load(config string) error {
	data, err := ioutil.ReadFile(config)
	if nil != err {
		s := fmt.Sprintf("mysql-proxy-server: DbList failed to readfile err(%s)", err.Error())
		return errors.New(s)
	}

	err = json.Unmarshal(data, this)
	if nil != err {
		s := fmt.Sprintf("mysql-proxy-server: DbList json unmarshal failed err(%s)!\n", err.Error())
		return errors.New(s)
	}

	root_path, _ := path.Split(config)

	for _, d := range this.DefineList {
		var config_loader mysql_generate.ConfigLoader
		if !config_loader.Load(root_path + d.Name) {
			return fmt.Errorf("mysql-proxy-server: DbList failed to load db define %v", root_path+d.Name)
		}
		if this.config_loaders == nil {
			this.config_loaders = make(map[int32]*mysql_generate.ConfigLoader)
		}
		this.config_loaders[d.Id] = &config_loader
	}

	if this.config_loaders == nil {
		return errors.New("mysql-proxy-server: DbList not found any db define")
	}

	for _, h := range this.MysqlHosts {
		if !h.Enable {
			continue
		}
		for _, d := range h.DbList {
			if d.Disable {
				continue
			}
			var c *mysql_generate.ConfigLoader
			if c = this.config_loaders[d.ConfigId]; c == nil {
				return fmt.Errorf("mysql-proxy-server: DbList not found db define by id %v ", d.ConfigId)
			}
			if d.Name != "" {
				var db_mgr mysql_manager.DB
				err := this.connect_db(&db_mgr, c, h, d.Name)
				if err != nil {
					return err
				}
				this.insert_db_mgr_list(&db_mgr, h.Id, d.Name)
			} else if d.NameList != nil {
				for _, name := range d.NameList {
					var db_mgr mysql_manager.DB
					err := this.connect_db(&db_mgr, c, h, name)
					if err != nil {
						return err
					}
					this.insert_db_mgr_list(&db_mgr, h.Id, name)
				}
			} else {
				return fmt.Errorf("mysql-proxy-server: DbList not found db host %v name or name list", h.Id)
			}
		}
	}

	return nil
}

func (this *DbList) connect_db(db_mgr *mysql_manager.DB, attach_define *mysql_generate.ConfigLoader, host *DbHost, db_name string) error {
	db_mgr.AttachConfig(attach_define)
	err := db_mgr.Connect(host.Ip, host.User, host.Password, db_name)
	if err != nil {
		return err
	}
	db_mgr.GoRun()
	return nil
}

func (this *DbList) insert_db_mgr_list(db_mgr *mysql_manager.DB, db_host_id int32, db_name string) {
	if this.db_mgr_list == nil {
		this.db_mgr_list = make(map[int32]map[string]*mysql_manager.DB)
	}
	dml := this.db_mgr_list[db_host_id]
	if dml == nil {
		dml = make(map[string]*mysql_manager.DB)
		this.db_mgr_list[db_host_id] = dml
	}
	dml[db_name] = db_mgr
}

func (this *DbList) GetDB(db_host_id int32, db_name string) *mysql_manager.DB {
	if this.db_mgr_list == nil {
		return nil
	}
	db_list := this.db_mgr_list[db_host_id]
	if db_list == nil {
		return nil
	}
	return db_list[db_name]
}

func (this *DbList) GetDB2(db_host_alias string, db_name string) *mysql_manager.DB {
	if this.db_mgr_list_by_alias == nil {
		return nil
	}
	db_list := this.db_mgr_list_by_alias[db_host_alias]
	if db_list == nil {
		return nil
	}
	return db_list[db_name]
}

var db_list DbList
