package data

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

var Db *sql.DB

func OpenDB() (*sql.DB, error) {
	host := os.Getenv("DATABASE_HOSTNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	port := 5432
	user := "postgres"
	dbname := "test"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func init() {
	var err error
	Db, err = OpenDB()
	if err != nil {
		panic(err)
	}

	if err = Db.Ping(); err != nil {
		panic(err)
	}
}
