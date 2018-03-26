package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

var (
	testDB *sql.DB
)

func TestMain(m *testing.M) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	os.Exit(m.Run())
}

func Test_initDB(t *testing.T) {

	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Create database", args{db: testDB}, true},
		{"Recreate existing database", args{db: testDB}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initDB(tt.args.db); got != tt.want {
				t.Errorf("initDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addToCooldown(t *testing.T) {
	type args struct {
		db       *sql.DB
		username string
	}
	tests := []struct {
		name string
		args args
	}{
		{"Add user1", args{username: "user1", db: testDB}},
		{"Add user2", args{username: "user2", db: testDB}},
		{"Add user3", args{username: "user3", db: testDB}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addToCooldown(tt.args.db, tt.args.username)
		})
	}
}

func Test_checkCooldown(t *testing.T) {
	type args struct {
		db       *sql.DB
		username string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Check cooldown for user1", args{username: "user1", db: testDB}, true},
		{"Check cooldown for user2", args{username: "user2", db: testDB}, true},
		{"Check cooldown for user3", args{username: "user3", db: testDB}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkCooldown(tt.args.db, tt.args.username); got != tt.want {
				t.Errorf("checkCooldown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resetCooldown(t *testing.T) {
	type args struct {
		db       *sql.DB
		username string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resetCooldown(tt.args.db, tt.args.username); got != tt.want {
				t.Errorf("resetCooldown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_expireCooldown(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expireCooldown(tt.args.db)
		})
	}
}
