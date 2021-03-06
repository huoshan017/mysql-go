package mysql_generate

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	mysql_base "github.com/huoshan017/mysql-go/base"
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
		tab.AfterLoad()
	}

	log.Printf("ConfigLoader::Load loaded config file %v\n", config)

	return true
}

func _get_field_simple_type(field *mysql_base.FieldConfig) (int, bool) {
	ft := field.Type
	n := strings.IndexAny(field.Type, ": (")
	if n >= 0 {
		t := []byte(field.Type)
		ft = string(t[:n])
	}
	return mysql_base.GetMysqlFieldTypeByString(strings.ToUpper(ft))
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

	if !has_primary && !tab.SingleRow {
		log.Printf("ConfigLoader::load_table %v not found primary key\n", tab.Name)
		return false
	}

	var str string
	var strs []string
	for _, f := range tab.Fields {
		// blob类型
		if strings.Index(f.Type, ":") >= 0 {
			strs = strings.Split(f.Type, ":")
			if len(strs) < 2 {
				log.Printf("ConfigLoader::load_table %v field blob type not found\n")
				return false
			}
			str = strings.ToUpper(strs[0])
			f.Type = strs[0]
			f.StructName = strs[1]
		} else {
			str = strings.ToUpper(f.Type)
		}

		real_type, ok := _get_field_simple_type(f)
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
	}
	return true
}

func (this *ConfigLoader) GenerateFieldStructsProto(dest_path_file string) bool {
	if this.FieldStructs == nil {
		return false
	}

	var f *os.File
	f = _get_file_creater(dest_path_file)
	if f == nil {
		return false
	}

	res := gen_proto(f, this.DBPkg, this.FieldStructs)

	if !res {
		log.Printf("写文件%v失败\n", f.Name)
		return false
	}

	if !_save_and_close_file(f, dest_path_file) {
		return false
	}

	return true
}

func _get_file_creater(dest_file string) *os.File {
	var f *os.File
	var err error
	f, err = os.OpenFile(dest_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Printf("打开文件%v失败 %v\n", dest_file, err.Error())
		return nil
	}
	return f
}

func _save_and_close_file(f *os.File, dest_file string) bool {
	var err error
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

func (this *ConfigLoader) _init_pkg_dirs(dest_path string) string {
	pkg_path := dest_path + "/" + this.DBPkg
	err := mysql_base.CreateDirs(pkg_path)
	if err != nil {
		return ""
	}
	return pkg_path
}

func (this *ConfigLoader) Generate(dest_path string) bool {
	if this.Tables == nil || len(this.Tables) == 0 {
		return false
	}

	pkg_path := this._init_pkg_dirs(dest_path)
	if pkg_path == "" {
		return false
	}

	for _, table := range this.Tables {
		dest_file := pkg_path + "/" + table.Name + ".go"
		f := _get_file_creater(dest_file)
		if f == nil {
			return false
		}

		res := gen_source(f, this.DBPkg, table)
		if !res {
			log.Printf("write source to %v failed\n", f.Name)
			return false
		}

		res = gen_proxy_source(f, this.DBPkg, table)
		if !res {
			log.Printf("write proxy source to %v failed\n", f.Name)
			return false
		}

		res = gen_record_mgr_source(f, this.DBPkg, table)
		if !res {
			log.Printf("write record mgr source to %v failed\n", f.Name)
			return false
		}

		if !_save_and_close_file(f, dest_file) {
			return false
		}
	}

	return true
}

func (this *ConfigLoader) GenerateInitFunc(dest_path string) bool {
	if this.Tables == nil || len(this.Tables) == 0 {
		return false
	}

	pkg_path := this._init_pkg_dirs(dest_path)
	if pkg_path == "" {
		return false
	}

	dest_file := pkg_path + "/init.go"
	f := _get_file_creater(dest_file)
	if f == nil {
		return false
	}

	gen_init_source(f, this.DBPkg, this.Tables)

	if !_save_and_close_file(f, dest_file) {
		return false
	}

	return true
}

func (this *ConfigLoader) GetTablesName() []string {
	if this.Tables == nil {
		return nil
	}

	var tables_name []string
	for _, t := range this.Tables {
		tables_name = append(tables_name, t.Name)
	}

	return tables_name
}
