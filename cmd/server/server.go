package main

import (
	"fmt"
	"github.com/ivangodev/gonahh/delivery/http"
	"github.com/ivangodev/gonahh/repository/psql"
	"github.com/ivangodev/gonahh/service"
)

func main() {
	db, err := psql.OpenDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to open DB: %s", err))
	}
	repo := psql.NewPSql(db)
	service := service.NewService(repo)
	delivery := http.NewDelivery(service)
	delivery.RegisterEndpoints()
	delivery.Start()
}
