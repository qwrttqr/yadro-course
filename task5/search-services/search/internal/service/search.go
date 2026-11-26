package service

import (
	"context"
	"log/slog"

	"yadro.com/course/search/internal/models"
	"yadro.com/course/search/internal/ports"
)

type SearchService struct {
	log *slog.Logger
	db  ports.DB
}

func NewSearchService(db ports.DB, log *slog.Logger) *SearchService {
	return &SearchService{db: db, log: log}
}

func (s *SearchService) Search(ctx context.Context, words []string) (models.SearchResults, error) {
	res, err := s.db.Search(ctx, words)
}
