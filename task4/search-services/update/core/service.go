package core

import (
	"context"
	"errors"
	"fmt"
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
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
	}, nil
}

func (s *Service) Update(ctx context.Context) (err error) {
	return errors.New("implement me")
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	return ServiceStats{}, errors.New("implement me")

}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {

	return errors.New("implement me")
}
