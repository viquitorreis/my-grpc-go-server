package db

import (
	"database/sql"
	"log"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(conn *sql.DB) {
	log.Println("Starting database migration")

	driver, _ := postgres.WithInstance(conn, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Error creating migration instance: %v", err)
	}

	// run all down migrations
	if err := m.Down(); err != nil {
		if err.Error() == "no change" {
			log.Println("No down migration to run")
		} else {
			log.Fatalf("Error running down migration: %v", err)
		}
	}

	// run all up migrations
	if err := m.Up(); err != nil {
		log.Fatalf("Error running up migration: %v", err)
	}

	log.Println("Database migration completed")
}
