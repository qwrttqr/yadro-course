package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	updating    atomic.Bool
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
	if !s.updating.CompareAndSwap(false, true) {
		return ErrAlreadyExists
	}
	defer s.updating.Store(false)
	lastId, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("get last xkcd id", "error", err)
		return fmt.Errorf("error get last id: %w", err)
	}
	ids, err := s.db.IDs(ctx)
	if err != nil {
		s.log.Error("get stored ids from db", "error", err)
		return fmt.Errorf("get stored ids from db: %w", err)
	}

	stored := make(map[int]struct{}, len(ids))

	for _, id := range ids {
		stored[id] = struct{}{}
	}

	group, workerCtx := errgroup.WithContext(ctx)
	jobs := make(chan int)
	for range s.concurrency {
		group.Go(func() error {
			for id := range jobs {
				if err := s.updateComic(ctx, id); err != nil {
					return err
				}
			}
			return nil
		})
	}

	group.Go(func() error {
		defer close(jobs)

		for id := 1; id <= lastId; id++ {
			if _, exists := stored[id]; exists {
				continue
			}
			select {
			case jobs <- id:
			case <-workerCtx.Done():
				return workerCtx.Err()
			}
		}
		return nil
	})
	return group.Wait()
}

func (s *Service) updateComic(ctx context.Context, id int) error {
	info, err := s.xkcd.Get(ctx, id)

	if errors.Is(err, ErrNotFound) {
		if err = s.db.Add(ctx, Comics{ID: info.ID}); err != nil && !errors.Is(err, ErrAlreadyExists) {
			return fmt.Errorf("record missing comic %d: %w", id, err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("fetch comic %d: %w", id, err)
	}

	words, err := s.words.Norm(ctx, info.Title+" "+info.Description)
	if err != nil {
		return fmt.Errorf("normalize words for comic %d: %w", id, err)
	}

	err = s.db.Add(ctx, Comics{
		ID:    info.ID,
		URL:   info.URL,
		Words: words,
	})
	if errors.Is(err, ErrAlreadyExists) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("save comic %d: %w", id, err)
	}
	return nil
}
func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	stats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("error getting stats from db", "error", err)
		return ServiceStats{}, fmt.Errorf("error getting stats from db: %w", err)
	}
	lastId, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("get last xkcd id", "error", err)
		return ServiceStats{}, fmt.Errorf("error get last id: %w", err)
	}
	return ServiceStats{
		WordsTotal:    stats.WordsTotal,
		WordsUnique:   stats.WordsUnique,
		ComicsFetched: stats.ComicsFetched,
		ComicsTotal:   lastId,
	}, nil

}

func (s *Service) Status(_ context.Context) ServiceStatus {
	if s.updating.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	err := s.db.Drop(ctx)
	if err != nil {
		s.log.Error("drop comics", "error", err)
		return fmt.Errorf("drop comics: %w", err)
	}
	return nil
}
