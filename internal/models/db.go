package models

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	MIGRATIONS_DIR_URI = "file://internal/models/migrations"
)

var DB *gorm.DB

var (
	user       string
	password   string
	host       string
	port       string
	db         string
	disableSSL string
)

// SetDBConfiguration reads the database connection configuration from env vars.
// This NEEDS to be called before InitDB().
func SetDBConfiguration() error {
	user = os.Getenv("POSTGRES_USER")
	if user == "" {
		return fmt.Errorf("unable to read env var POSTGRES_USER")
	}
	password = os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return fmt.Errorf("unable to read env var POSTGRES_PASSWORD")
	}
	host = os.Getenv("POSTGRES_HOST")
	if host == "" {
		return fmt.Errorf("unable to read env var POSTGRES_HOST")
	}
	db = os.Getenv("POSTGRES_DB_NAME")
	if db == "" {
		return fmt.Errorf("unable to read env var POSTGRES_DB_NAME")
	}
	port = os.Getenv("POSTGRES_PORT")
	if port == "" {
		return fmt.Errorf("unable to read env var POSTGRES_PORT")
	}
	disableSSL = os.Getenv("POSTGRES_DISABLE_SSL")
	return nil
}

// InitDB initializes the database handler.
func InitDB() error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", host, user, password, db, port)
	if disableSSL == "true" {
		dsn = fmt.Sprintf("%s sslmode=disable", dsn)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

// Migrate runs the migrations present in the specified URI. If destroy is true,
// the migrations are run downwards than upwards.
func Migrate(migrationsDirUri string, destroy bool) error {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, db)
	if disableSSL == "true" {
		connStr = fmt.Sprintf("%s?sslmode=disable", connStr)
	}

	if migrationsDirUri == "" {
		migrationsDirUri = MIGRATIONS_DIR_URI
	}
	m, err := migrate.New(migrationsDirUri, connStr)
	if err != nil {
		return err
	}
	if destroy {
		if err := m.Down(); err != nil && err.Error() != "no change" {
			return err
		}
	} else {
		if err := m.Up(); err != nil && err.Error() != "no change" {
			return err
		}
	}

	return nil
}
