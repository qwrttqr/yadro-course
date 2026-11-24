package db

import (
	"context"
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
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
			INSERT INTO comics (id, url) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING
    `, comics.ID, comics.URL)

	if err != nil {
		return fmt.Errorf("insert comics: %w", err)
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check whether comics: %d, was inserted: %w", comics.ID, err)
	}
	if inserted == 0 {
		return nil
	}
	if len(comics.Words) > 0 {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO words (comics_id, word)
		SELECT $1, word
		FROM unnest($2::text[]) AS w(word);
		`)
		if err != nil {
			return fmt.Errorf("insert words for comics: %d, %w", comics.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit comic %d: %w", comics.ID, err)
	}
	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	const query = `SELECT COUNT(id), COUNT(DISTINCT word), (SELECT COUNT(*) FROM comics) FROM words`

	var stats core.DBStats

	err := db.conn.QueryRowContext(ctx, query).Scan(
		&stats.WordsTotal,
		&stats.WordsUnique,
		&stats.ComicsFetched,
	)
	if err != nil {
		return core.DBStats{}, fmt.Errorf("get db stats: %w", err)
	}
	return stats, err
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	const query = `SELECT id FROM comics`
	ids := make([]int, 0)
	if err := db.conn.SelectContext(ctx, &ids, query); err != nil {
		return nil, fmt.Errorf("get comic IDs: %w", err)
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, `TRUNCATE TABLE comics, words`); err != nil {
		return fmt.Errorf("clear tables: %w", err)
	}
	return nil
}
