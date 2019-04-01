package mysql_generator

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/huoshan017/mysql-go/base"
)

type ConfigLoader struct {
	DBPkg  string                    `json:"db_pkg"`
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

	if this.DBPkg == "" {
		log.Printf("ConfigLoader::Load db_pkg is empty\n")
		return false
	}

	for _, tab := range this.Tables {
		if !this.load_table(tab) {
			return false
		}
	}

	log.Printf("ConfigLoader::Load loaded config file %v\n", config)

	return true
}

func (this *ConfigLoader) load_table(tab *mysql_base.TableConfig) bool {
	engine := strings.ToUpper(tab.Engine)
	var ok bool
	if _, ok = mysql_base.GetMysqlEngineTypeByString(engine); !ok {
		log.Printf("ConfigLoader::load_table unsupported engine type %v\n", engine)
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
		log.Printf("ConfigLoader::load_table %v not found primary key\n", tab.Name)
		return false
	}

	var str string
	for _, f := range tab.Fields {
		str = strings.ToUpper(f.Type)
		_, ok = mysql_base.GetMysqlFieldTypeByString(str)
		if !ok {
			log.Printf("ConfigLoader::load_table %v field type %v not found\n", tab.Name, str)
			return false
		}

		str = strings.ToUpper(f.IndexType)
		_, ok = mysql_base.GetMysqlIndexTypeByString(str)
		if !ok {
			log.Printf("ConfigLoader::load_table %v index type %v not found\n", tab.Name, str)
			return false
		}

		strs := strings.Split(f.CreateFlags, ",")
		for _, s := range strs {
			str = strings.ToUpper(s)
			_, ok = mysql_base.GetMysqlTableCreateFlagTypeByString(str)
			if !ok {
				log.Printf("ConfigLoader::load_table %v create flag %v not found\n", tab.Name, str)
				return false
			}
		}
	}
	return true
}

func create_dirs(dest_path string) (err error) {
	if err = os.MkdirAll(dest_path, os.ModePerm); err != nil {
		log.Printf("创建目录结构%v错误 %v\n", dest_path, err.Error())
		return
	}
	if err = os.Chmod(dest_path, os.ModePerm); err != nil {
		log.Printf("修改目录%v权限错误 %v\n", dest_path, err.Error())
		return
	}
	return
}

func (this *ConfigLoader) Generate(dest_path string) bool {
	if this.Tables == nil || len(this.Tables) == 0 {
		return false
	}

	pkg_path := dest_path + "/" + this.DBPkg
	err := create_dirs(pkg_path)
	if err != nil {
		return false
	}

	for _, table := range this.Tables {
		dest_file := pkg_path + "/" + table.Name + ".go"
		var f *os.File
		f, err = os.OpenFile(dest_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Printf("打开文件%v失败 %v\n", dest_file, err.Error())
			return false
		}

		res := gen_source(f, pkg_path, table)

		if !res {
			log.Printf("写文件%v失败\n", f.Name)
			return false
		}

		var err error
		if err = f.Sync(); err != nil {
			log.Printf("同步文件%v失败 %v\n", dest_file, err.Error())
			return false
		}
		if err = f.Close(); err != nil {
			log.Printf("关闭文件%v失败 %v\n", dest_file, err.Error())
			return false
		}
	}

	return true
}
