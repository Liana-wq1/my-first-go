package storage

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitSQLite() {
	var err error
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./app.db"
	}

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Создаём таблицу users
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT,
		last_name TEXT,
		email TEXT UNIQUE,
		phone TEXT,
		avatar_url TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		is_confirmed INTEGER DEFAULT 0,
		confirm_token TEXT
	);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Создаём таблицу credentials
	//важно, чтобы user_id существовал в таблице users, Нельзя создать запись в credentials, если такого user нет
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS credentials (
		user_id INTEGER PRIMARY KEY,
		password_hash TEXT,
		last_login_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id) 
	);
	`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("SQLite initialized")
}
