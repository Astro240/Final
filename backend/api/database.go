package api

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
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
		user_type INTEGER,
		profile_picture TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_type) REFERENCES user_type(id)
	);
	CREATE TABLE IF NOT EXISTS verification_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		code TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		session_token TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE TABLE IF NOT EXISTS user_type (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS stores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		color_scheme TEXT,
		logo TEXT,
		owner_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (owner_id) REFERENCES users(id)
	);
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		store_id INTEGER,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		image TEXT,
		quantity INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (store_id) REFERENCES stores(id)
	);
	INSERT OR IGNORE INTO user_type (id, name) VALUES
	(1, 'admin'),
	(2, 'user');
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
