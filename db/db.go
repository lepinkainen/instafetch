package db

import (
	"database/sql"
	"fmt"

	// sqlite libraries on top of database/sql
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

// initialize database table in sqlite
func initDB(db *sql.DB) bool {
	// ID, username and timestamp the user was added to the cooldown list
	sqlStmt := "create table cooldown (id integer not null primary key, username text unique, start_time datetime not null);"

	_, err := db.Exec(sqlStmt)
	if err != nil {
		if err.Error() == "table cooldown already exists" {
			return false
		}
		log.Infof("%q: %s\n", err, sqlStmt)
		return false
	}

	return true
}

// AddCooldown adds user to cooldown list if they aren't on it already
func AddCooldown(db *sql.DB, username string) {
	if !CheckCooldown(db, username) {
		addToCooldown(db, username)
	}
}

// add user to cooldown list
func addToCooldown(db *sql.DB, username string) {
	// begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into cooldown(username, start_time) values(?, datetime('now'))")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	// actual transaction
	_, err = stmt.Exec(fmt.Sprintf(username))
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

// CheckCooldown returns true if the user is still on cooldown
func CheckCooldown(db *sql.DB, username string) bool {
	stmt, err := db.Prepare("select username, start_time from cooldown where username = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		return true
	}

	return false
}

// ResetCooldown deletes users who have been on cooldown long enough
func ResetCooldown(db *sql.DB) bool {
	stmt, err := db.Prepare("delete from cooldown where start_time < datetime('now', '-1 day');")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	count, err := res.RowsAffected()
	if count > 0 {
		return true
	}
	return false
}

// GetConnection returns a connection to the sqlite DB
func GetConnection() *sql.DB {
	// open DB connection
	db, err := sql.Open("sqlite3", "./instafetch.db")
	if err != nil {
		log.Fatal(err)
	}

	// crate table
	init := initDB(db)
	if init {
		log.Info("Created table")
	}

	return db
}
