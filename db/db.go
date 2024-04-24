package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func GetDb() *sql.DB {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS WeightRecords (
		id INTEGER PRIMARY KEY,
		date DATE UNIQUE,
		weight REAL);
		`,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS HeartRateRecords (
		id INTEGER PRIMARY KEY,
		time TIME UNIQUE,
		heartRate REAL);
		`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS StepsRecords (
		id INTEGER PRIMARY KEY,
		time TIME UNIQUE,
		steps REAL);
		`,
	)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
