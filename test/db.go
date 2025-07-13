package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SensorData struct {
	ID        string
	Type      string
	Value     float64
	Unit      string
	Timestamp string
}

func main() {
	// Kết nối đến file database SQLite
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Truy vấn dữ liệu
	rows, err := db.Query(`SELECT id, type, value, unit, timestamp FROM sensor`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var s SensorData
		if err := rows.Scan(&s.ID, &s.Type, &s.Value, &s.Unit, &s.Timestamp); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s | %s | %.2f %s | %s\n", s.ID, s.Type, s.Value, s.Unit, s.Timestamp)
	}
}
