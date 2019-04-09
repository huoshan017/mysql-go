package mysql_base

import (
	"log"
)

type FieldValuePair struct {
	Name  string
	Value interface{}
}

func (this *Database) InsertRecord(table_name string, field_args ...FieldValuePair) (res bool, last_insert_id int64) {
	fl := len(field_args)
	if fl > 0 {
		var field_list, placehold_list string
		var args []interface{}
		for i, fa := range field_args {
			if i == 0 {
				placehold_list = "?"
				field_list = fa.Name
			} else {
				placehold_list += ",?"
				field_list += ("," + fa.Name)
			}
			args = append(args, fa.Value)
		}
		res = this.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	} else {
		res = this.Exec("INSERT INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

func (this *Database) InsertRecord2(table_name string, fields []string, values []interface{}) (res bool, last_insert_id int64) {
	var fl int
	if fields != nil {
		fl := len(fields)
		if values == nil || fl != len(values) {
			log.Printf("Database:InsertRecord2 fields length must equal to values length\n")
			return false, 0
		}
	}

	if fl > 0 {
		var field_list, placehold_list string
		for i := 0; i < fl; i++ {
			if i == 0 {
				placehold_list = "?"
				field_list = fields[i]
			} else {
				placehold_list += ",?"
				field_list += ("," + fields[i])
			}
		}
		res = this.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", values, &last_insert_id, nil)
	} else {
		res = this.Exec("INSERT INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

func _gen_select_query_str(table_name string, field_list []string) string {
	var query_str string
	if field_list == nil || len(field_list) == 0 {
		query_str = "SELECT * FROM " + table_name + " WHERE ?=?"
	} else {
		query_str = "SELECT "
		for i := 0; i < len(field_list); i++ {
			query_str += field_list[i]
			if i < len(field_list)-1 {
				query_str += ", "
			}
		}
		query_str += (" FROM " + table_name + " WHERE ?=?;")
	}
	return query_str
}

func (this *Database) SelectRecord(table_name, key_name string, key_value interface{}, field_list []string, dest_list []interface{}) bool {
	if dest_list == nil || len(dest_list) == 0 {
		log.Printf("Database::SelectRecord result dest_list could not empty\n")
		return false
	}
	query_str := _gen_select_query_str(table_name, field_list)
	return this.QueryOneWith(query_str, []interface{}{key_name, key_value}, dest_list)
}

func (this *Database) SelectRecords(table_name, key_name string, key_value interface{}, field_list []string, result_list *QueryResultList) bool {
	if result_list == nil {
		log.Printf("Database::SelectRecords result_list could not null\n")
		return false
	}
	query_str := _gen_select_query_str(table_name, field_list)
	return this.QueryWith(query_str, []interface{}{key_name, key_value}, result_list)
}

func (this *Database) UpdateRecord(table_name string, key_name string, key_value interface{}, field_args ...FieldValuePair) bool {
	fl := len(field_args)
	if fl <= 0 {
		return false
	}
	var args []interface{}
	query_str := "UPDATE " + table_name + " SET "
	for _, fa := range field_args {
		query_str += (fa.Name + "=?")
		args = append(args, fa.Value)
	}
	query_str += (" WHERE " + key_name + "=?;")
	args = append(args, key_value)
	return this.ExecWith(query_str, args, nil, nil)
}

func (this *Database) DeleteRecord(table_name string, key_name string, key_value interface{}) bool {
	return this.ExecWith("DELETE FROM "+table_name+" WHERE ?=?;", []interface{}{key_name, key_value}, nil, nil)
}
