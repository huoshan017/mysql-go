package mysql_generator

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/huoshan017/mysql-go/base"
)

type FieldStructMember struct {
	Name  string `json:"name"`
	Index int32  `json:"index"`
	Type  string `json:"type"`
}

type FieldStruct struct {
	Name    string               `json:"name"`
	Members []*FieldStructMember `json:"members"`
}

type ConfigLoader struct {
	DBPkg        string                    `json:"db_pkg"`
	Charset      string                    `json:"charset"`
	Tables       []*mysql_base.TableConfig `json:"tables"`
	FieldStructs []*FieldStruct            `json:"field_structs"`
}

func (this *ConfigLoader) GetTable(table_name string) *mysql_base.TableConfig {
	if this.Tables == nil {
		return nil
	}
	for _, t := range this.Tables {
		if t.Name == table_name {
			return t
		}
	}
	return nil
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
	var strs []string
	for _, f := range tab.Fields {
		if strings.Index(f.Type, ":") >= 0 {
			strs = strings.Split(f.Type, ":")
			str = strings.ToUpper(strs[0])
			f.Type = strs[0]
		} else {
			str = strings.ToUpper(f.Type)
		}

		var real_type int
		real_type, ok = mysql_base.GetMysqlFieldTypeByString(str)
		if !ok {
			log.Printf("ConfigLoader::load_table %v field type %v not found\n", tab.Name, str)
			return false
		}

		f.RealType = real_type

		str = strings.ToUpper(f.IndexType)
		var real_index_type int
		real_index_type, ok = mysql_base.GetMysqlIndexTypeByString(str)
		if !ok {
			log.Printf("ConfigLoader::load_table %v index type %v not found\n", tab.Name, str)
			return false
		}

		f.RealIndexType = real_index_type

		strs = strings.Split(f.CreateFlags, ",")
		for _, s := range strs {
			str = strings.ToUpper(s)
			str = strings.Trim(str, " ")
			if str == "" {
				continue
			}
			_, ok = mysql_base.GetMysqlTableCreateFlagTypeByString(str)
			if !ok {
				log.Printf("ConfigLoader::load_table %v create flag %v not found\n", tab.Name, str)
				return false
			}
		}
	}
	return true
}

func (this *ConfigLoader) GenerateFieldStructsProto(dest_path string) bool {
	if this.FieldStructs == nil {
		return false
	}

	var f *os.File
	var err error
	dest_file := dest_path + "/" + this.DBPkg + "_field_structs.proto"
	f, err = os.OpenFile(dest_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Printf("打开文件%v失败 %v\n", dest_file, err.Error())
		return false
	}

	res := gen_proto(f, this.DBPkg, this.FieldStructs)

	if !res {
		log.Printf("写文件%v失败\n", f.Name)
		return false
	}

	if err = f.Sync(); err != nil {
		log.Printf("同步文件%v失败 %v\n", dest_file, err.Error())
		return false
	}
	if err = f.Close(); err != nil {
		log.Printf("关闭文件%v失败 %v\n", dest_file, err.Error())
		return false
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

		res := gen_source(f, this.DBPkg, table)

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
