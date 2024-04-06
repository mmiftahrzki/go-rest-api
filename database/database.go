package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func New() *sql.DB {
	db, err := sql.Open("mysql", "root:toor@tcp(localhost:3306)/portfolio?parseTime=true")
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	db.SetMaxOpenConns(10)

	return db
}
