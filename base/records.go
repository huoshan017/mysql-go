package mysql_base

import (
	"log"
	"strconv"
)

type FieldValuePair struct {
	Name  string
	Value interface{}
}

type OpDetail struct {
	TableName string
	OpType    int32
	Key       string
	Value     interface{}
	FieldList []*FieldValuePair
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

func (this *Database) InsertRecord(table_name string, field_args ...*FieldValuePair) (err error, last_insert_id int64) {
	fl := len(field_args)
	if fl > 0 {
		field_list, placehold_list, args := _gen_insert_params(field_args...)
		err = this.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	} else {
		err = this.Exec("INSERT INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

func (this *Database) InsertIgnoreRecord(table_name string, field_args ...*FieldValuePair) (err error, last_insert_id int64) {
	var fl = len(field_args)
	if fl > 0 {
		field_list, placehold_list, args := _gen_insert_params(field_args...)
		err = this.ExecWith("INSERT IGNORE INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	} else {
		err = this.Exec("INSERT IGNORE INTO "+table_name+";", &last_insert_id, nil)
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

/*func (this *Database) InsertRecord2(table_name string, fields []string, values []interface{}) (res bool, last_insert_id int64) {
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
}*/

func _gen_select_query_str(table_name string, field_list []string, field_name string, sel_cond *SelectCondition) string {
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
	if sel_cond != nil && sel_cond.OrderBy != "" {
		if sel_cond.Desc {
			limit_str += "ORDER BY " + sel_cond.OrderBy + " DESC"
		} else {
			limit_str += "ORDER BY " + sel_cond.OrderBy + " ASC"
		}
	}
	if sel_cond != nil && sel_cond.Offset >= 0 {
		limit_str += " LIMIT " + strconv.Itoa(sel_cond.Offset) + ", " + strconv.Itoa(sel_cond.Limit)
	}

	if field_name != "" {
		if limit_str != "" {
			query_str += (" WHERE " + field_name + "=? " + limit_str + ";")
		} else {
			query_str += (" WHERE " + field_name + "=?;")
		}
	} else {
		query_str += (" " + limit_str + ";")
	}

	return query_str
}

func (this *Database) SelectRecord(table_name, key_name string, key_value interface{}, field_list []string, dest_list []interface{}) error {
	if dest_list == nil || len(dest_list) == 0 {
		log.Printf("Database::SelectRecord result dest_list cant not empty\n")
		return ErrArgumentInvalid
	}
	var sel_cond = SelectCondition{
		Offset: -1,
		Limit:  -1,
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name, &sel_cond)
	return this.QueryOneWith(query_str, []interface{}{key_value}, dest_list)
}

func (this *Database) SelectRecords(table_name, key_name string, key_value interface{}, field_list []string, result_list *QueryResultList) error {
	if result_list == nil {
		//log.Printf("Database::SelectRecords result_list cant not null\n")
		return ErrArgumentInvalid
	}
	var sel_cond = SelectCondition{
		Offset: -1,
		Limit:  -1,
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name, &sel_cond)
	if key_name != "" {
		return this.QueryWith(query_str, []interface{}{key_value}, result_list)
	} else {
		return this.Query(query_str, result_list)
	}
}

const (
	COMPARITION_EQUAL       = iota
	COMPARITION_GREAT_THAN  = 1
	COMPARITION_LESS_THAN   = 2
	COMPARITION_GREAT_EQUAL = 3
	COMPARITION_LESS_EQUAL  = 4
)

type SelectCondition struct {
	CompType int
	OrderBy  string
	Desc     bool
	Offset   int
	Limit    int
}

func (this *Database) SelectRecordsCondition(table_name, field_name string, field_value interface{}, sel_cond *SelectCondition, field_list []string, result_list *QueryResultList) error {
	if result_list == nil {
		//log.Printf("Database::SelectRecords result_list cant not null\n")
		return ErrArgumentInvalid
	}
	query_str := _gen_select_query_str(table_name, field_list, field_name, sel_cond)
	if field_name != "" {
		return this.QueryWith(query_str, []interface{}{field_value}, result_list)
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

func (this *Database) UpdateRecord(table_name string, key_name string, key_value interface{}, field_args ...*FieldValuePair) error {
	query_str, args := _gen_update_params(table_name, key_name, key_value, field_args...)
	return this.ExecWith(query_str, args, nil, nil)
}

func (this *Database) DeleteRecord(table_name string, key_name string, key_value interface{}) error {
	sql_str := "DELETE FROM " + table_name + " WHERE " + key_name + "=?;"
	return this.ExecWith(sql_str, []interface{}{key_value}, nil, nil)
}

func (this *Procedure) InsertRecord(table_name string, field_args ...*FieldValuePair) (err error, last_insert_id int64) {
	field_list, placehold_list, args := _gen_insert_params(field_args...)
	err = this.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	return
}

func (this *Procedure) InsertIgnoreRecord(table_name string, field_args ...*FieldValuePair) (err error, last_insert_id int64) {
	field_list, placehold_list, args := _gen_insert_params(field_args...)
	err = this.ExecWith("INSERT IGNORE INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	return
}

func (this *Procedure) UpdateRecord(table_name string, key_name string, key_value interface{}, field_args ...*FieldValuePair) error {
	query_str, args := _gen_update_params(table_name, key_name, key_value, field_args...)
	return this.ExecWith(query_str, args, nil, nil)
}

func (this *Procedure) DeleteRecord(table_name string, key_name string, key_value interface{}) error {
	sql_str := "DELETE FROM " + table_name + " WHERE " + key_name + "=?;"
	return this.ExecWith(sql_str, []interface{}{key_value}, nil, nil)
}
