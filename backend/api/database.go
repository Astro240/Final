package api

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

const DATABASEPATH = "./database/database.db"

func CreateDatabase() {
	// Implementation for creating the database if it doesn't exist
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		first_name TEXT,
		last_name TEXT,
		age INTEGER,
		profile_picture TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
