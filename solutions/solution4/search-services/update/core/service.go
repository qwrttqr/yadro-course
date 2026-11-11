package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	inProgress  atomic.Bool
	lock        sync.Mutex
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
	if ok := s.lock.TryLock(); !ok {
		s.log.Error("service already runs update or drop")
		return ErrAlreadyExists
	}
	defer s.lock.Unlock()

	s.inProgress.Store(true)
	defer s.inProgress.Store(false)

	s.log.Info("update started")
	defer func(start time.Time) {
		s.log.Info("update finished", "duration", time.Since(start), "error", err)
	}(time.Now())

	// get existing IDs in DB
	IDs, err := s.db.IDs(ctx)
	if err != nil {
		s.log.Error("failed to get existing IDs in DB", "error", err)
		return fmt.Errorf("failed to get existing IDs in DB: %v", err)
	}
	s.log.Debug("existing comics in DB", "count", len(IDs))
	exists := make(map[int]bool, len(IDs))
	for _, id := range IDs {
		exists[id] = true
	}

	// get last comics ID
	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		slog.Error("failed to get last ID in XKCD", "error", err)
		return fmt.Errorf("failed to get last ID in XKCD: %v", err)
	}
	s.log.Debug("last comics ID in XKCD", "id", lastID)

	unknownIDs := generateIDs(ctx, 1, lastID, exists)
	comics := s.getComics(ctx, unknownIDs)
	return s.processComics(ctx, comics)
}

func generateIDs(ctx context.Context, first, last int, exists map[int]bool) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := first; i <= last; i++ {
			if exists[i] {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case ch <- i:
			}
		}
	}()
	return ch
}

type FetchInfo struct {
	info XKCDInfo
	err  error
}

func (s *Service) processComics(ctx context.Context, in <-chan FetchInfo) error {
	var failed int
	var added int
	for comics := range in {
		info, err := comics.info, comics.err
		if err != nil {
			failed++
			s.log.Error("failed to get comics", "id", info.ID, "error", err)
			continue
		}
		words, err := s.words.Norm(ctx, info.Description+" "+info.Title)
		if err != nil {
			failed++
			s.log.Error("failed to normalize", "id", info.ID, "error", err)
			continue
		}
		err = s.db.Add(ctx, Comics{
			ID:    info.ID,
			URL:   info.URL,
			Words: words,
		})
		if err != nil {
			failed++
			s.log.Error("failed to save comics", "id", info.ID, "error", err)
			continue
		}
		added++
	}
	s.log.Debug("comics", "added", added, "failed", failed)
	if failed != 0 {
		return fmt.Errorf("failed to fetch/store some comics: %d out of %d", failed, failed+added)
	}
	return nil
}

func (s *Service) getComics(ctx context.Context, in <-chan int) <-chan FetchInfo {
	out := make(chan FetchInfo)
	var wg sync.WaitGroup

	for i := range s.concurrency {
		wg.Go(func() {
			s.log.Debug("fetcher up", "id", i)
			defer s.log.Debug("fetcher down", "id", i)
			for id := range in {
				if id == 404 {
					// special case
					out <- FetchInfo{
						info: XKCDInfo{ID: id, Title: "404", Description: "Not found"},
					}
					continue
				}
				info, err := s.xkcd.Get(ctx, id)
				s.log.Debug("fetched", "id", id)
				if err != nil {
					info = XKCDInfo{ID: id}
				}
				out <- FetchInfo{info: info, err: err}
			}
		})
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("failed to get stats", "error", err)
		return ServiceStats{}, err
	}
	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get last comics ID", "error", err)
		return ServiceStats{}, err
	}
	return ServiceStats{
		DBStats:     dbStats,
		ComicsTotal: lastID,
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if s.inProgress.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	if ok := s.lock.TryLock(); !ok {
		s.log.Error("service already runs update or drop")
		return ErrAlreadyExists
	}
	defer s.lock.Unlock()
	err := s.db.Drop(ctx)
	if err != nil {
		s.log.Error("failed to drop db entries", "error", err)
	}
	return err
}
