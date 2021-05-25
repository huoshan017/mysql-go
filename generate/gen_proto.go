package mysql_generate

import (
	"log"
	"os"
	"strconv"
)

func gen_proto(f *os.File, pkg_name string, field_structs []*FieldStruct) bool {
	str := "syntax = \"proto3\";\n"
	//str += "package " + pkg_name + ";\n"
	str += "option go_package=\"./" + pkg_name + "\";\n\n"

	for _, fs := range field_structs {
		str += _gen_struct(fs)
		str += "\n"
	}

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}
	return true
}

func _gen_struct(field_struct *FieldStruct) string {
	var str string
	str = "message " + field_struct.Name + "{\n"
	for _, m := range field_struct.Members {
		str += ("	" + m.Type + " " + m.Name + " = " + strconv.Itoa(int(m.Index)) + ";\n")
	}
	str += "}\n"
	return str
}
