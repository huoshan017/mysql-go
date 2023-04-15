package mysql_generate

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	mysql_base "github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/log"
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
	configBytes  []byte
}

func (c *ConfigLoader) GetTable(table_name string) *mysql_base.TableConfig {
	if c.Tables == nil {
		return nil
	}
	for _, t := range c.Tables {
		if t.Name == table_name {
			return t
		}
	}
	return nil
}

func (c *ConfigLoader) LoadConfigBytes(bytes []byte) bool {
	err := json.Unmarshal(bytes, c)
	if nil != err {
		log.Infof("ConfigLoader::Load json unmarshal failed err(%s)!", err.Error())
		return false
	}

	if c.DBPkg == "" {
		log.Infof("ConfigLoader::Load db_pkg is empty")
		return false
	}

	for _, tab := range c.Tables {
		if !c.load_table(tab) {
			return false
		}
		tab.AfterLoad()
	}

	return true
}

func (c *ConfigLoader) Load(config string) bool {
	data, err := ioutil.ReadFile(config)
	if nil != err {
		log.Infof("ConfigLoader::Load failed to readfile err(%s)!", err.Error())
		return false
	}

	if !c.LoadConfigBytes(data) {
		return false
	}

	c.configBytes = data

	log.Infof("ConfigLoader::Load loaded config file %v", config)

	return true
}

func _get_field_simple_type(field *mysql_base.FieldConfig) (int, bool) {
	ft := field.TypeStr
	n := strings.IndexAny(field.TypeStr, ": (")
	if n >= 0 {
		t := []byte(field.TypeStr)
		ft = string(t[:n])
	}
	return mysql_base.GetMysqlFieldTypeByString(strings.ToUpper(ft))
}

func (c *ConfigLoader) load_table(tab *mysql_base.TableConfig) bool {
	engine := strings.ToUpper(tab.Engine)
	var ok bool
	if _, ok = mysql_base.GetMysqlEngineTypeByString(engine); !ok {
		log.Infof("ConfigLoader::load_table unsupported engine type %v", engine)
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
		log.Infof("ConfigLoader::load_table %v not found primary key", tab.Name)
		return false
	}

	var str string
	var strs []string
	for _, f := range tab.Fields {
		// blob类型
		if strings.Contains(f.TypeStr, ":") {
			strs = strings.Split(f.TypeStr, ":")
			if len(strs) < 2 {
				log.Infof("ConfigLoader::load_table %v field blob type not found", tab.Name)
				return false
			}
			str = strings.ToUpper(strs[0])
			f.TypeStr = strs[0]
			f.StructName = strs[1]
		} else {
			str = strings.ToUpper(f.TypeStr)
		}

		real_type, ok := _get_field_simple_type(f)
		if !ok {
			log.Infof("ConfigLoader::load_table %v field type %v not found", tab.Name, str)
			return false
		}

		f.Type = real_type

		str = strings.ToUpper(f.IndexStr)
		var real_index_type int
		real_index_type, ok = mysql_base.GetMysqlIndexTypeByString(str)
		if !ok {
			log.Infof("ConfigLoader::load_table %v index type %v not found", tab.Name, str)
			return false
		}

		f.Index = real_index_type
	}
	return true
}

func (c *ConfigLoader) GenerateFieldStructsProto(dest_path_file string) bool {
	if c.FieldStructs == nil {
		return false
	}

	f := _get_file_creater(dest_path_file)
	if f == nil {
		return false
	}

	err := gen_proto(f, c.DBPkg, c.FieldStructs)

	if err != nil {
		log.Fatalf("gen_proto err: %v", err)
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
		log.Infof("打开文件%v失败 %vn", dest_file, err.Error())
		return nil
	}
	return f
}

func _save_and_close_file(f *os.File, dest_file string) bool {
	var err error
	if err = f.Sync(); err != nil {
		log.Infof("同步文件%v失败 %v", dest_file, err.Error())
		return false
	}
	if err = f.Close(); err != nil {
		log.Infof("关闭文件%v失败 %v", dest_file, err.Error())
		return false
	}
	return true
}

func (c *ConfigLoader) _init_pkg_dirs(dest_path string) string {
	pkg_path := dest_path + "/" + c.DBPkg
	err := mysql_base.CreateDirs(pkg_path)
	if err != nil {
		return ""
	}
	return pkg_path
}

func (c *ConfigLoader) Generate(dest_path string) bool {
	if c.Tables == nil || len(c.Tables) == 0 {
		return false
	}

	pkg_path := c._init_pkg_dirs(dest_path)
	if pkg_path == "" {
		return false
	}

	for _, table := range c.Tables {
		dest_file := pkg_path + "/" + table.Name + ".go"
		f := _get_file_creater(dest_file)
		if f == nil {
			return false
		}

		res := gen_source(f, c.DBPkg, table)
		if !res {
			log.Infof("write source to %v failed", f.Name())
			return false
		}

		res = gen_proxy_source(f, c.DBPkg, table)
		if !res {
			log.Infof("write proxy source to %v failed", f.Name())
			return false
		}

		res = gen_record_mgr_source(f, c.DBPkg, table)
		if !res {
			log.Infof("write record mgr source to %v failed", f.Name())
			return false
		}

		if !_save_and_close_file(f, dest_file) {
			return false
		}
	}

	return true
}

func (c *ConfigLoader) GenerateInitFunc(dest_path string) bool {
	if c.Tables == nil || len(c.Tables) == 0 {
		return false
	}

	pkg_path := c._init_pkg_dirs(dest_path)
	if pkg_path == "" {
		return false
	}

	dest_file := pkg_path + "/init.go"
	f := _get_file_creater(dest_file)
	if f == nil {
		return false
	}

	gen_init_source(f, c.DBPkg, c.configBytes, c.Tables)

	return _save_and_close_file(f, dest_file)
}

func (c *ConfigLoader) GetTablesName() []string {
	if c.Tables == nil {
		return nil
	}

	var tables_name []string
	for _, t := range c.Tables {
		tables_name = append(tables_name, t.Name)
	}

	return tables_name
}
