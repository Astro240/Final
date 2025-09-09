package api

import (
	"database/sql"
	"log"
)

const DATABASEPATH = "./database/database.db"

func createDatabase() {
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
		password TEXT NOT NULL
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}