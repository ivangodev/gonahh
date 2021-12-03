package data

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

var Db *sql.DB

func init() {
	host := os.Getenv("DATABASE_HOSTNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	port := 5432
	user := "postgres"
	dbname := "test"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password)
	var err error
	Db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	if err = Db.Ping(); err != nil {
		panic(err)
	}
}
