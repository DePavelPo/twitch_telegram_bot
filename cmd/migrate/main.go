package main

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

const migrationDir = "./db/migrations"

func main() {

	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	migrateDown := false

	for _, arg := range os.Args {
		if arg == "--down" {
			migrateDown = true
		}
	}

	err = migrate("postgres", os.Getenv("DB_CONN"), migrateDown)
	if err != nil {
		logrus.Fatal(err)
	}

}

func migrate(dialect string, creds string, migrateDown bool) error {
	db, err := sql.Open(dialect, creds)
	if err != nil {
		return errors.Errorf("cannot open postgres db connection: %v", err)
	}

	defer db.Close()

	err = goose.SetDialect(dialect)
	if err != nil {
		return errors.Errorf("cannot set %s dialect: %v", dialect, err)
	}

	if migrateDown {
		err = goose.Down(db, migrationDir)
		if err != nil {
			return errors.Errorf("cannot down %s migrations: %v", dialect, err)
		}
		return nil
	}

	err = goose.Up(db, migrationDir, goose.WithAllowMissing())
	if err != nil {
		return errors.Errorf("cannot up %s migrations: %v", dialect, err)
	}
	return nil

}
