package mysql_base

import (
	"fmt"
	"strings"

	"github.com/huoshan017/mysql-go/log"
)

func (db *Database) LoadTable(tab *TableConfig) error {
	// create table
	var sql_str string
	if !tab.SingleRow {
		primary_field := tab.GetPrimaryKeyFieldConfig()
		if primary_field == nil {
			log.Infof("Database::LoadTable %v cant get primary key field config", tab.Name)
			return ErrPrimaryFieldNotDefine
		}
		sql_str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`%s` %s, PRIMARY KEY(`%s`)) ENGINE=%s", tab.Name, tab.PrimaryKey, primary_field.TypeStr, tab.PrimaryKey, tab.Engine)
	} else {
		sql_str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`place_hold` int(11), PRIMARY KEY(`place_hold`)) ENGINE=%s", tab.Name, tab.Engine)
	}

	err := db.Exec(sql_str, nil, nil)
	if err != nil {
		return err
	}

	// add fields
	for _, f := range tab.Fields {
		if tab.PrimaryKey == f.Name {
			continue
		}
		err = db.add_field(tab.Name, f)
		if err != nil {
			return err
		}
	}

	if tab.SingleRow {
		sql_str = fmt.Sprintf("INSERT IGNORE INTO `%s` (`place_hold`) VALUES (1)", tab.Name)
		err = db.Exec(sql_str, nil, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) DropTable(table_name string) error {
	args := []interface{}{table_name}
	return db.ExecWith("DROP TABLE ?", args, nil, nil)
}

func (db *Database) add_field(table_name string, field *FieldConfig) error {
	sql_str := fmt.Sprintf("DESCRIBE %s %s", table_name, field.Name)
	if db.HasRow(sql_str) {
		return nil
	}

	sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", table_name, field.Name, field.TypeStr)
	err := db.Exec(sql_str, nil, nil)
	if err != nil {
		log.Infof("create table %v field %v failed", table_name, field.Name)
		return err
	}

	// create index
	index_type, o := GetMysqlIndexTypeByString(strings.ToUpper(field.IndexStr))
	if !o {
		log.Infof("No supported index type %v", field.IndexStr)
		return fmt.Errorf("table %v has not supported index type %v", table_name, field.IndexStr)
	}

	if index_type != MYSQL_INDEX_TYPE_NONE {
		if index_type == MYSQL_INDEX_TYPE_NORMAL {
			sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD INDEX %s_index(`%s`)", table_name, field.Name, field.Name)
		} else if index_type == MYSQL_INDEX_TYPE_UNIQUE {
			sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE (`%s`)", table_name, field.Name)
		} else if index_type == MYSQL_INDEX_TYPE_FULLTEXT {
			sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD FULLTEXT(`%s`)", table_name, field.Name)
		} else {
			log.Infof("table %v field %v index type FULLTEXT not supported", table_name, field.Name)
		}

		err = db.Exec(sql_str, nil, nil)
		if err != nil {
			log.Infof("create table %v field %v index failed", table_name, field.Name)
			return err
		}
	}

	return nil
}

/*func (db *Database) remove_field(table_name, field_name string) error {
	sql_str := "ALTER TABLE " + table_name + " DROP COLUMN " + field_name
	return db.Exec(sql_str, nil, nil)
}

func (db *Database) rename_field(table_name, old_field_name, new_field_name string) error {
	sql_str := "ALTER TABLE " + table_name + " CHANGE " + old_field_name + " " + new_field_name
	return db.Exec(sql_str, nil, nil)
}

func (db *Database) modify_field_attr(table_name string, field *FieldConfig) error {
	sql_str := "ALTER TABLE " + table_name + " MODIFY " + field.Name + " " + field.Type
	err := db.Exec(sql_str, nil, nil)
	if err != nil {
		log.Printf("modify table %v field %v attr failed\n", table_name, field.Name)
		return err
	}
	return nil
}*/
