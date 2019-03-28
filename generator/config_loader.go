package mysql_generator

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/huoshan017/mysql-go/base"
)

type ConfigLoader struct {
	DBName string                    `json:"db_name"`
	Tables []*mysql_base.TableConfig `json:"tables"`
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

func (this *ConfigLoader) load_table(tab *mysql_base.TableConfig) bool {
	engine := strings.ToUpper(tab.Engine)
	var ok bool
	if _, ok = mysql_base.GetMysqlEngineTypeByString(engine); !ok {
		log.Printf("ConfigLoader::load_table unsupported engine type %v", engine)
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

	var str string
	for _, f := range tab.Fields {
		str = strings.ToUpper(f.Type)
		_, ok = mysql_base.GetMysqlFieldTypeByString(str)
		if !ok {
			log.Printf("ConfigLoader::load_table %v field type %v not found", tab.Name, str)
			return false
		}

		str = strings.ToUpper(f.IndexType)
		_, ok = mysql_base.GetMysqlIndexTypeByString(str)
		if !ok {
			log.Printf("ConfigLoader::load_table %v index type %v not found", tab.Name, str)
			return false
		}

		strs := strings.Split(f.CreateFlags, ",")
		for _, s := range strs {
			str = strings.ToUpper(s)
			_, ok = mysql_base.GetMysqlTableCreateFlagTypeByString(str)
			if !ok {
				log.Printf("ConfigLoader::load_table %v create flag %v not found", tab.Name, str)
				return false
			}
		}
	}
	return true
}

func (this *ConfigLoader) Generate() bool {
	return true
}