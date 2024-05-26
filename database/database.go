package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var once sync.Once
var database_connection *sql.DB

func new() *sql.DB {
	var err error
	var db *sql.DB

	db, err = sql.Open("mysql", "root:toor@tcp(localhost:3306)/portfolio?parseTime=true")
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

func GetDatabaseConnection() *sql.DB {
	if database_connection != nil {
		fmt.Println("koneksi database sudah pernah dibuat.")

		return database_connection
	}

	once.Do(func() {
		database_connection = new()
	})

	return database_connection
}
