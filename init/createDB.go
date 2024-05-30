package createDB

import (
	"database/sql"
	//  "log"
  
	 _ "github.com/lib/pq"
  )

func CreateAndOpen(name string) *sql.DB {

	conninfo := "user=postgres password=yourpassword host=127.0.0.1 sslmode=disable"
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
 
	_,err = db.Exec("CREATE DATABASE IF NOT EXISTS "+name)
	if err != nil {
		panic(err)
	}
	db.Close()
 
	db, err = sql.Open("postgres", "admin:admin@tcp(127.0.0.1:3306)/" + name)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	return db
 }