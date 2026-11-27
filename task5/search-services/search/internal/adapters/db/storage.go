package db

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/search/internal/models"
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

func (db *DB) Search(ctx context.Context, words []string, limit int) (models.SearchResults, error) {
	const query = `
		   SELECT c.id, c.url
		   FROM comics AS c 
		   JOIN words AS w ON c.id = w.comics_id 
		   WHERE w.word = ANY($1::text[])
		   GROUP BY c.id, c.url
		   ORDER BY COUNT(DISTINCT w.word) DESC, c.id ASC
		   LIMIT $2
		`
	results := make([]models.SearchResultUnit, 0)
	if err := db.conn.SelectContext(ctx, &results, query, words, limit); err != nil {
		return models.SearchResults{}, fmt.Errorf("search comics: %w", err)
	}
	return models.SearchResults{Result: results, Total: len(results)}, nil
}
