package main

import (
    "github.com/tanel/dbmigrate"
    "log"
)

func DoDatabaseMigrations() {
    error := dbmigrate.Run(db, configuration.DatabaseMigrationsPath)
    if error!=nil {
        log.Fatalf("Errors encountered migrating database: %v", error)
    }
}
