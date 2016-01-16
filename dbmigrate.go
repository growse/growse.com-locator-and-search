package main

import (
	"github.com/tanel/dbmigrate"
	"log"
)

func DoDatabaseMigrations() {
	log.Printf("Starting Database Migrations from %v", configuration.DatabaseMigrationsPath)
	error := dbmigrate.Run(db, configuration.DatabaseMigrationsPath)
	if error != nil {
		log.Fatalf("Errors encountered migrating database: %v", error)
	}
	log.Printf("Done")
}
