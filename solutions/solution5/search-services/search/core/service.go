package core

import (
	"cmp"
	"context"
	"log/slog"
	"maps"
	"slices"
)

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
}

func NewService(log *slog.Logger, db DB, words Words) (*Service, error) {
	return &Service{
		log:   log,
		db:    db,
		words: words,
	}, nil
}

func (s *Service) Search(ctx context.Context, phrase string, limit int) ([]Comics, error) {
	keywords, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("failed to find keywords", "error", err)
		return nil, err
	}
	s.log.Debug("normalized query", "keywords", keywords)

	// comics ID -> number of findings
	scores := map[int]int{}
	for _, keyword := range keywords {
		IDs, err := s.db.Search(ctx, keyword)
		if err != nil {
			s.log.Error("failed to search keyword in DB", "error", err)
			return nil, err
		}
		for _, ID := range IDs {
			scores[ID]++
		}
	}
	s.log.Debug("relevant comics", "count", len(scores))

	// sort by number of findings
	sorted := slices.SortedFunc(maps.Keys(scores), func(a, b int) int {
		return cmp.Compare(scores[b], scores[a]) // desc
	})

	// limit results
	if len(sorted) < limit {
		limit = len(sorted)
	}
	sorted = sorted[:limit]

	// fetch comics
	result := make([]Comics, 0, len(sorted))
	for _, ID := range sorted {
		comics, err := s.db.Get(ctx, ID)
		if err != nil {
			s.log.Error("failed to fetch comics", "id", ID, "error", err)
			return nil, err
		}
		result = append(result, comics)
	}
	s.log.Debug("returning comics", "count", len(result))

	return result, nil
}
