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

const (
	COMPARITION_EQUAL       = iota
	COMPARITION_GREAT_THAN  = 1
	COMPARITION_LESS_THAN   = 2
	COMPARITION_GREAT_EQUAL = 3
	COMPARITION_LESS_EQUAL  = 4
)

var comp_type = []string{
	"=",
	">",
	"<",
	">=",
	"<=",
}

type SelectCondition struct {
	CompType int
	OrderBy  string
	Desc     bool
	Offset   int
	Limit    int
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

func (db *Database) InsertRecord(table_name string, field_args ...*FieldValuePair) (last_insert_id int64, err error) {
	fl := len(field_args)
	if fl > 0 {
		field_list, placehold_list, args := _gen_insert_params(field_args...)
		err = db.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	} else {
		err = db.Exec("INSERT INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

func (db *Database) InsertIgnoreRecord(table_name string, field_args ...*FieldValuePair) (last_insert_id int64, err error) {
	var fl = len(field_args)
	if fl > 0 {
		field_list, placehold_list, args := _gen_insert_params(field_args...)
		err = db.ExecWith("INSERT IGNORE INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	} else {
		err = db.Exec("INSERT IGNORE INTO "+table_name+";", &last_insert_id, nil)
	}
	return
}

/*func _gen_insert_params_2(fields []string, values []interface{}) (field_list, placehold_list string) {
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
}*/

func _gen_select_query_str(table_name string, field_list []string, field_name string, sel_cond *SelectCondition) string {
	var query_str string
	if len(field_list) == 0 {
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

	if field_name != "" && sel_cond != nil {
		if limit_str != "" {
			query_str += (" WHERE " + field_name + comp_type[sel_cond.CompType] + "? " + limit_str + ";")
		} else {
			query_str += (" WHERE " + field_name + comp_type[sel_cond.CompType] + "?;")
		}
	} else {
		query_str += (" " + limit_str + ";")
	}

	return query_str
}

func (db *Database) SelectRecord(table_name, key_name string, key_value interface{}, field_list []string, dest_list []interface{}) error {
	if len(dest_list) == 0 {
		log.Printf("Database::SelectRecord result dest_list cant not empty\n")
		return ErrArgumentInvalid
	}
	var sel_cond = SelectCondition{
		Offset: -1,
		Limit:  -1,
	}
	query_str := _gen_select_query_str(table_name, field_list, key_name, &sel_cond)
	return db.QueryOneWith(query_str, []interface{}{key_value}, dest_list)
}

func (db *Database) SelectRecords(table_name, field_name string, field_value interface{}, field_list []string, result_list *QueryResultList) error {
	if result_list == nil {
		//log.Printf("Database::SelectRecords result_list cant not null\n")
		return ErrArgumentInvalid
	}
	var sel_cond = SelectCondition{
		Offset: -1,
		Limit:  -1,
	}
	query_str := _gen_select_query_str(table_name, field_list, field_name, &sel_cond)
	if field_name != "" {
		return db.QueryWith(query_str, []interface{}{field_value}, result_list)
	} else {
		return db.Query(query_str, result_list)
	}
}

func (db *Database) SelectRecordsCondition(table_name, field_name string, field_value interface{}, sel_cond *SelectCondition, field_list []string, result_list *QueryResultList) error {
	if result_list == nil {
		//log.Printf("Database::SelectRecords result_list cant not null\n")
		return ErrArgumentInvalid
	}
	query_str := _gen_select_query_str(table_name, field_list, field_name, sel_cond)
	if field_name != "" {
		return db.QueryWith(query_str, []interface{}{field_value}, result_list)
	} else {
		return db.Query(query_str, result_list)
	}
}

func (db *Database) SelectRecordsCount(table_name, field_name string, field_value interface{}) (count int32, err error) {
	query_str := "SELECT COUNT(*) FROM " + table_name
	if field_name != "" {
		query_str += " WHERE " + field_name + "=?;"
		count, err = db.QueryCountWith(query_str, field_value)
	} else {
		count, err = db.QueryCount(query_str)
	}
	return
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

func (db *Database) UpdateRecord(table_name string, key_name string, key_value interface{}, field_args ...*FieldValuePair) error {
	query_str, args := _gen_update_params(table_name, key_name, key_value, field_args...)
	return db.ExecWith(query_str, args, nil, nil)
}

func (db *Database) DeleteRecord(table_name string, key_name string, key_value interface{}) error {
	sql_str := "DELETE FROM " + table_name + " WHERE " + key_name + "=?;"
	return db.ExecWith(sql_str, []interface{}{key_value}, nil, nil)
}

func (db *Procedure) InsertRecord(table_name string, field_args ...*FieldValuePair) (last_insert_id int64, err error) {
	field_list, placehold_list, args := _gen_insert_params(field_args...)
	err = db.ExecWith("INSERT INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	return
}

func (db *Procedure) InsertIgnoreRecord(table_name string, field_args ...*FieldValuePair) (last_insert_id int64, err error) {
	field_list, placehold_list, args := _gen_insert_params(field_args...)
	err = db.ExecWith("INSERT IGNORE INTO "+table_name+"("+field_list+") VALUES ("+placehold_list+");", args, &last_insert_id, nil)
	return
}

func (db *Procedure) UpdateRecord(table_name string, key_name string, key_value interface{}, field_args ...*FieldValuePair) error {
	query_str, args := _gen_update_params(table_name, key_name, key_value, field_args...)
	return db.ExecWith(query_str, args, nil, nil)
}

func (db *Procedure) DeleteRecord(table_name string, key_name string, key_value interface{}) error {
	sql_str := "DELETE FROM " + table_name + " WHERE " + key_name + "=?;"
	return db.ExecWith(sql_str, []interface{}{key_value}, nil, nil)
}
