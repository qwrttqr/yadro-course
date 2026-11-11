package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	return nil, nil
}

func (s *Service) Update(ctx context.Context) (err error) {
	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	return ServiceStats{}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	return nil
}
