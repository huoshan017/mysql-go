package mysql

import (
	"log"
)

func test1() {
	var db Database
	err := db.Open("localhost", "root", "moyu123", "golib_database")
	defer db.Close()
	if err != nil {
		log.Printf("open db err %v", err.Error())
		return
	}
}
