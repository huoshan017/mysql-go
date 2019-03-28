package mysql

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/huoshan017/golib/mysql/base"
)

type FieldConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Length      int    `json:"length"`
	IndexType   string `json:"index_type"`
	CreateFlags string `json:"create_flags"`
}

type TableConfig struct {
	Name       string         `json:"name"`
	PrimaryKey string         `json:"primary_key"`
	Engine     string         `json:"engine"`
	Fields     []*FieldConfig `json:"fields"`
}

type ConfigLoader struct {
	DBName string         `json:"db_name"`
	Tables []*TableConfig `json:"tables"`
}

func (this *ConfigLoader) Load(config string) bool {
	data, err := ioutil.ReadFile(config)
	if nil != err {
		log.Printf("ConfigLoader::Load failed to readfile err(%s)!\n", err.Error())
		return false
	}

	err = json.Unmarshal(data, this)
	if nil != err {
		log.Printf("ConfigLoader::Load json unmarshal failed err(%s)!\n", err.Error())
		return false
	}

	if this.DBName == "" {
		log.Printf("ConfigLoader::Load db_name is empty\n")
		return false
	}

	for _, tab := range this.Tables {
		if !this.load_table(tab) {
			return false
		}
	}

	return true
}

func (this *ConfigLoader) load_table(tab *TableConfig) bool {
	engine := strings.ToLower(tab.Engine)
	if engine != "myisam" && engine != "innodb" {
		return false
	}

	var has_primary bool
	for _, f := range tab.Fields {
		if f.Name == tab.PrimaryKey {
			has_primary = true
			break
		}
	}

	if !has_primary {
		log.Printf("ConfigLoader::load_table %v not found primary key", tab.Name)
		return false
	}

	for _, f := range tab.Fields {

	}
	return true
}

func (this *ConfigLoader) Generate() bool {
	return true
}
