package main

import (
	"NeuroNest/internal/config"
	"NeuroNest/internal/db"
	"NeuroNest/internal/router"
	"log"
)

func main() {
	config.LoadEnv()

	db.ConnectDBPostgres()
	db.AutoMigrateTables()

	r := router.RouterConfig()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
