package api

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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
		email TEXT NOT NULL,
		password TEXT NOT NULL,
		first_name TEXT,
		last_name TEXT,
		age INTEGER,
		user_type INTEGER,
		profile_picture TEXT,
		store_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_type) REFERENCES user_type(id) ON DELETE CASCADE,
		FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS verification_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		code TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		session_token TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS user_type (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS stores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		template TEXT,
		color_scheme TEXT,
		logo TEXT,
		banner TEXT,
		owner_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		store_id INTEGER,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		image TEXT,
		quantity INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
	);
	INSERT OR IGNORE INTO user_type (id, name) VALUES
	(1, 'admin'),
	(2, 'user');
	CREATE TABLE IF NOT EXISTS payment_methods (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user_id INTEGER,
    	method_name TEXT NOT NULL,
    	account_details TEXT, -- This could be a card number or account ID
    	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS transactions (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	order_id INTEGER,
    	payment_method_id INTEGER,
    	amount REAL NOT NULL,
    	transaction_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    	status TEXT NOT NULL, -- e.g., 'completed', 'pending', 'failed'
    	FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    	FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id) ON DELETE CASCADE
	);

	-- Cart table
	CREATE TABLE IF NOT EXISTS cart (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user_id INTEGER NOT NULL,
    	product_id INTEGER NOT NULL,
    	quantity INTEGER NOT NULL DEFAULT 1,
    	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    	UNIQUE(user_id, product_id)
	);

	-- Orders table
	CREATE TABLE IF NOT EXISTS orders (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user_id INTEGER NOT NULL,
    	store_id INTEGER NOT NULL,
    	total_amount REAL NOT NULL,
    	status TEXT NOT NULL DEFAULT 'pending',
    	shipping_info TEXT NOT NULL,
    	payment_info TEXT,
    	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    	FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
	);

	-- Order product table
	CREATE TABLE IF NOT EXISTS order_products (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	order_id INTEGER NOT NULL,
    	product_id INTEGER NOT NULL,
    	quantity INTEGER NOT NULL,
    	price REAL NOT NULL,
    	FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS favorites (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user_id INTEGER NOT NULL,
    	product_id INTEGER NOT NULL,
    	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);
	insert or ignore into users (id, email, password, first_name, user_type) values
	(0, 'astropify@gmail.com', '$2a$10$MleK0bpPilssP8IyMai7A.5azKNbMcz8bJrLxWmbLnhcKtRHY87V2', 'admin', 1);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
