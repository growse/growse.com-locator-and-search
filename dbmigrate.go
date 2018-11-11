package main

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func DoDatabaseMigrations(db *sql.DB, migrationsPath string) {
	log.Printf("Starting Database Migrations from %v", configuration.DatabaseMigrationsPath)
	driver, err := postgres.WithInstance(db, &postgres.Config{MigrationsTable: "migrations"})

	if err != nil {
		log.Fatalf("Errors encountered creating migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "www_growse_com", driver)

	if err != nil {
		log.Fatalf("Errors encountered creating migrate instance : %v", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Errors encountered migrating database: %v", err)
	}
	log.Print("Database migration done")
}
