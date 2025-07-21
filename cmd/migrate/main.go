package main

import (
	"database/sql"
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

const migrationDir = "./db/migrations"

func main() {
	// Load .env if present, don't fail if missing
	_ = godotenv.Load()

	var (
		downFlag = flag.Bool("down", false, "Run migrations down instead of up")
		dbConn   = os.Getenv("DB_CONN")
	)
	flag.Parse()

	if dbConn == "" {
		logrus.Fatal("DB_CONN environment variable is required")
	}

	if err := runMigrations("postgres", dbConn, *downFlag); err != nil {
		logrus.Fatalf("Migration failed: %+v", err)
	}
}

func runMigrations(dialect, creds string, migrateDown bool) error {
	db, err := sql.Open(dialect, creds)
	if err != nil {
		return errors.Errorf("cannot open %s db connection: %v", dialect, err)
	}
	defer db.Close()

	if err := goose.SetDialect(dialect); err != nil {
		return errors.Errorf("cannot set %s dialect: %v", dialect, err)
	}

	if migrateDown {
		if err := goose.Down(db, migrationDir); err != nil {
			return errors.Errorf("cannot down %s migrations: %v", dialect, err)
		}
		logrus.Info("Migrations rolled back successfully")
		return nil
	}

	if err := goose.Up(db, migrationDir, goose.WithAllowMissing()); err != nil {
		return errors.Errorf("cannot up %s migrations: %v", dialect, err)
	}
	logrus.Info("Migrations applied successfully")
	return nil
}
