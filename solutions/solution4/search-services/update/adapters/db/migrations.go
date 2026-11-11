package db

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Logger struct {
	logger *slog.Logger
}

func (l Logger) Verbose() bool {
	return true
}

func (l Logger) Printf(format string, v ...any) {
	l.logger.Debug("migrate", "msg", fmt.Sprintf(format, v...))
}

func (db *DB) Migrate() error {
	db.log.Debug("running migration")
	files, err := iofs.New(migrationFiles, "migrations") // get migrations from
	if err != nil {
		return err
	}
	driver, err := pgx.WithInstance(db.conn.DB, &pgx.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("myiofs", files, "mypg", driver)
	if err != nil {
		return err
	}
	m.Log = Logger{logger: db.log}

	err = m.Up()

	if err != nil {
		if err != migrate.ErrNoChange {
			db.log.Error("migration failed", "error", err)
			return err
		}
		db.log.Debug("migration did not change anything")
	}

	db.log.Debug("migration finished")
	return nil
}
