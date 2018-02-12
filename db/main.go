package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

// initialize database table in sqlite
func initDB(db *sql.DB) bool {
	// ID, username and timestamp the user was added to the cooldown list
	sqlStmt := "create table cooldown (id integer not null primary key, username text, timestamp datetime default current_timestmp);"

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

// add user to cooldown list
func addToCooldown(db *sql.DB, username string) {
	// begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into cooldown(username) values(?)")
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

// check if the user is still on cooldown
func checkCooldown(db *sql.DB, username string) bool {
	stmt, err := db.Prepare("select username, timestamp from cooldown where username = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(fmt.Sprintf(username))
	if err != nil {
		log.Fatal(err)
	}

	count, err := res.RowsAffected()
	if count > 0 {
		return true
	}
	return false
}

func resetCooldown(db *sql.DB, username string) bool {
	stmt, err := db.Prepare("delete from cooldown where username = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(fmt.Sprintf(username))
	if err != nil {
		log.Fatal(err)
	}

	count, err := res.RowsAffected()
	if count > 0 {
		return true
	}
	return false
}

func main() {
	// open DB connection
	db, err := sql.Open("sqlite3", "./instafetch.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// crate table
	init := initDB(db)
	if init {
		log.Info("Created table")
	}

}
