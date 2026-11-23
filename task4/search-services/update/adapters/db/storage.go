package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Add(ctx context.Context, comics core.Comics) (err error) {
	words := comics.Words
	if words == nil {
		words = []string{}
	}

	const query = `INSERT INTO comics (id, url, words)	VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	if _, err := db.conn.ExecContext(ctx, query, comics.ID, comics.URL, words); err != nil {
		return fmt.Errorf("insert comics %d: %w", comics.ID, err)
	}
	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	return core.DBStats{}, errors.New("implement me")
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	return nil, errors.New("implement me")
}

func (db *DB) Drop(ctx context.Context) error {
	return errors.New("implement me")
}
