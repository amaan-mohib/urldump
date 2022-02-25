package internal

import (
	"cli/utils"
	"database/sql"
)

type SQLite struct {
	DB *sql.DB
}

var DBGlobal *sql.DB

func SQLConnInit(dbPath string) *SQLite {
	db, dbErr := sql.Open("sqlite3", dbPath)
	utils.CheckError(dbErr)

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS
		links (
			ID	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			url	TEXT,
			title TEXT,
			tags TEXT,
			status TEXT DEFAULT "unread"
		);
	`)
	utils.CheckError(err)
	stmt.Exec()
	defer stmt.Close()

	return &SQLite{
		DB: db,
	}
}

func SetSQLConn(db *sql.DB) {
	DBGlobal = db
}

func GetSQLConn() *SQLite {
	return &SQLite{
		DB: DBGlobal,
	}
}
