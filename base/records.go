package mysql_base

import (
	"log"
	"strconv"
)

type FieldValuePair struct {
	Name  string
	Value interface{}
}

func _gen_insert_params(field_args ...*FieldValuePair) (field_list, placehold_list string, args []interface{}) {
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
	return
}

func (this *Database) InsertRecord(table_name string, field_args ...*FieldValuePair) (res bool, last_insert_id int64) {
	fl := len(field_args)
	if fl > 0 {
		field_list, placehold_list, args := _gen_insert_params(field_args...)
		res = this.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	} else {
		res = this.Exec("INSERT INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

func (this *Database) InsertIgnoreRecord(table_name string, field_args ...*FieldValuePair) (res bool, last_insert_id int64) {
	var fl = len(field_args)
	if fl > 0 {
		field_list, placehold_list, args := _gen_insert_params(field_args...)
		res = this.ExecWith("INSERT IGNORE INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	} else {
		res = this.Exec("INSERT IGNORE INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

func _gen_insert_params_2(fields []string, values []interface{}) (field_list, placehold_list string) {
	for i := 0; i < len(fields); i++ {
		if i == 0 {
			placehold_list = "?"
			field_list = fields[i]
		} else {
			placehold_list += ",?"
			field_list += ("," + fields[i])
		}
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
		field_list, placehold_list := _gen_insert_params_2(fields, values)
		res = this.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", values, &last_insert_id, nil)
	} else {
		res = this.Exec("INSERT INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

func _gen_select_query_str(table_name string, field_list []string, key string, order_by string, descent bool, offset, limit int) string {
	var query_str string
	if field_list == nil || len(field_list) == 0 {
		query_str = "SELECT * FROM " + table_name
	} else {
		query_str = "SELECT "
		for i := 0; i < len(field_list); i++ {
			query_str += field_list[i]
			if i < len(field_list)-1 {
				query_str += ", "
			}
		}
		query_str += (" FROM " + table_name)
	}

	var limit_str string
	if order_by != "" {
		if descent {
			limit_str += "ORDER BY " + order_by + " DESC"
		} else {
			limit_str += "ORDER BY " + order_by + " ASC"
		}
	}
	if offset >= 0 {
		limit_str += " LIMIT " + strconv.Itoa(offset) + ", " + strconv.Itoa(limit)
	}

	if key != "" {
		if limit_str != "" {
			query_str += (" WHERE " + key + "=? " + limit_str + ";")
		} else {
			query_str += (" WHERE " + key + "=?;")
		}
	} else {
		query_str += (" " + limit_str + ";")
	}

	return query_str
}

func (this *Database) SelectRecord(table_name, key_name string, key_value interface{}, field_list []string, dest_list []interface{}) bool {
	if dest_list == nil || len(dest_list) == 0 {
		log.Printf("Database::SelectRecord result dest_list cant not empty\n")
		return false
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name, "", false, -1, -1)
	return this.QueryOneWith(query_str, []interface{}{key_value}, dest_list)
}

func (this *Database) SelectRecords(table_name, key_name string, key_value interface{}, field_list []string, result_list *QueryResultList) bool {
	if result_list == nil {
		log.Printf("Database::SelectRecords result_list cant not null\n")
		return false
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name, "", false, -1, -1)
	if key_name != "" {
		return this.QueryWith(query_str, []interface{}{key_value}, result_list)
	} else {
		return this.Query(query_str, result_list)
	}
}

func (this *Database) SelectRecordsOrderby(table_name, key_name string, key_value interface{}, order_by string, desc bool, offset, limit int, field_list []string, result_list *QueryResultList) bool {
	if result_list == nil {
		log.Printf("Database::SelectRecords result_list cant not null\n")
		return false
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name, order_by, desc, offset, limit)
	if key_name != "" {
		return this.QueryWith(query_str, []interface{}{key_value}, result_list)
	} else {
		return this.Query(query_str, result_list)
	}
}

func _gen_update_params(table_name string, key_name string, key_value interface{}, field_args ...*FieldValuePair) (query_str string, args []interface{}) {
	query_str = "UPDATE " + table_name + " SET "
	for i, fa := range field_args {
		if i == 0 {
			query_str += (fa.Name + "=?")
		} else {
			query_str += (", " + fa.Name + "=?")
		}
		args = append(args, fa.Value)
	}
	query_str += (" WHERE " + key_name + "=?;")
	args = append(args, key_value)
	return
}

func (this *Database) UpdateRecord(table_name string, key_name string, key_value interface{}, field_args ...*FieldValuePair) bool {
	fl := len(field_args)
	if fl <= 0 {
		return false
	}
	query_str, args := _gen_update_params(table_name, key_name, key_value, field_args...)
	return this.ExecWith(query_str, args, nil, nil)
}

func (this *Database) DeleteRecord(table_name string, key_name string, key_value interface{}) bool {
	sql_str := "DELETE FROM " + table_name + " WHERE " + key_name + "=?;"
	return this.ExecWith(sql_str, []interface{}{key_value}, nil, nil)
}

func (this *Procedure) InsertRecord(table_name string, field_args ...*FieldValuePair) (res bool, last_insert_id int64) {
	fl := len(field_args)
	if fl == 0 {
		return
	}
	field_list, placehold_list, args := _gen_insert_params(field_args...)
	res = this.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	return
}

func (this *Procedure) InsertIgnoreRecord(table_name string, field_args ...*FieldValuePair) (res bool, last_insert_id int64) {
	fl := len(field_args)
	if fl == 0 {
		return
	}
	field_list, placehold_list, args := _gen_insert_params(field_args...)
	res = this.ExecWith("INSERT IGNORE INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	return
}

func (this *Procedure) UpdateRecord(table_name string, key_name string, key_value interface{}, field_args ...*FieldValuePair) bool {
	fl := len(field_args)
	if fl <= 0 {
		return false
	}
	query_str, args := _gen_update_params(table_name, key_name, key_value, field_args...)
	return this.ExecWith(query_str, args, nil, nil)
}

/*func (this *Procedure) SelectRecord(table_name, key_name string, key_value interface{}, field_list []string, dest_list []interface{}) bool {
	if dest_list == nil || len(dest_list) == 0 {
		log.Printf("Procedure::SelectRecord result dest_list could not empty\n")
		return false
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name)
	return this.QueryOneWith(query_str, []interface{}{key_value}, dest_list)
}

func (this *Procedure) SelectRecords(table_name, key_name string, key_value interface{}, field_list []string, result_list *QueryResultList) bool {
	if result_list == nil {
		log.Printf("Procedure::SelectRecords result_list could not null\n")
		return false
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name)
	return this.QueryWith(query_str, []interface{}{key_value}, result_list)
}*/

func (this *Procedure) DeleteRecord(table_name string, key_name string, key_value interface{}) bool {
	sql_str := "DELETE FROM " + table_name + " WHERE " + key_name + "=?;"
	return this.ExecWith(sql_str, []interface{}{key_value}, nil, nil)
}
