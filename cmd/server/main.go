package main

import (
	_ "NeuroNest/docs"
	"NeuroNest/internal/config"
	"NeuroNest/internal/db"
	"NeuroNest/internal/router"
	"log"
)

// @Title						---
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	config.LoadEnv()

	db.ConnectDBPostgres()
	db.AutoMigrateTables()

	r := router.RouterConfig()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
