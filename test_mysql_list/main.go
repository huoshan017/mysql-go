package main

import (
	"log"

	"github.com/huoshan017/mysql-go/base"
)

// test
type TestNode struct {
	op_type    int32
	table_name string
	field_list []*mysql_base.FieldValuePair
}

func output_list(list *mysql_base.List) {
	log.Printf("list node: \n")
	node := list.GetHeadNode()
	for node != nil {
		d := node.GetData()
		test_node := d.(*TestNode)
		log.Printf("		 %v\n", test_node)
		node = node.GetNext()
	}
}

func main() {
	var list mysql_base.List
	var test_node_list = []*TestNode{
		&TestNode{
			op_type:    mysql_base.DB_OPERATE_TYPE_INSERT_RECORD,
			table_name: "Players",
			field_list: []*mysql_base.FieldValuePair{&mysql_base.FieldValuePair{Name: "Id", Value: 1}},
		},
		&TestNode{
			op_type:    mysql_base.DB_OPERATE_TYPE_INSERT_RECORD,
			table_name: "Players",
			field_list: []*mysql_base.FieldValuePair{&mysql_base.FieldValuePair{Name: "Id", Value: 2}},
		},
		&TestNode{
			op_type:    mysql_base.DB_OPERATE_TYPE_DELETE_RECORD,
			table_name: "Mails",
		},
		&TestNode{
			op_type:    mysql_base.DB_OPERATE_TYPE_UPDATE_RECORD,
			table_name: "Skills",
		},
		&TestNode{
			op_type:    mysql_base.DB_OPERATE_TYPE_INSERT_RECORD,
			table_name: "Globals",
		},
		&TestNode{
			op_type:    mysql_base.DB_OPERATE_TYPE_DELETE_RECORD,
			table_name: "Guilds",
		},
		&TestNode{
			op_type:    mysql_base.DB_OPERATE_TYPE_DELETE_RECORD,
			table_name: "Servers",
		},
	}

	for _, n := range test_node_list {
		list.Append(n)
	}

	output_list(&list)

	head := list.GetHeadNode()
	if head != nil {
		list.MoveToLast(head.GetData())
		output_list(&list)
	}

	for n := 0; n < 10; n++ {
		head = list.GetHeadNode()
		next := head.GetNext()
		if next != nil {
			list.MoveToLast(next.GetData())
			output_list(&list)
		}
	}

	list.Clear()

	log.Printf("cleared...\n")

	for _, n := range test_node_list {
		list.Append(n)
	}

	for n := 0; n < 10; n++ {
		head = list.GetHeadNode()
		next := head.GetNext()
		if next != nil {
			list.MoveToLast(next.GetData())
			output_list(&list)
		}
	}
}
