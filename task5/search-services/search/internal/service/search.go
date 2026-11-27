package service

import (
	"context"
	"log/slog"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
	"yadro.com/course/search/internal/models"
	"yadro.com/course/search/internal/ports"
)

type SearchService struct {
	logger *slog.Logger
	db     ports.DB
}

func NewSearchService(db ports.DB, log *slog.Logger) *SearchService {
	return &SearchService{db: db, logger: log}
}

func (s *SearchService) Search(ctx context.Context, phrase string, limit int) (models.SearchResults, error) {
	words := s.Normalize(phrase)
	if phrase == "" {
		return models.SearchResults{}, ErrBadArguments
	}
	res, err := s.db.Search(ctx, words, limit)
	if err != nil {
		s.logger.Error("error searching in DB", "error", err)
		return models.SearchResults{}, err
	}
	return res, nil
}

func (p *SearchService) Normalize(input string) []string {
	tokens := strings.FieldsFunc(strings.ToLower(input), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	seen := make(map[string]struct{})
	var words []string
	for _, token := range tokens {
		if english.IsStopWord(token) {
			continue
		}

		stem := english.Stem(token, false)

		if _, exists := seen[stem]; exists {
			continue
		}

		seen[stem] = struct{}{}

		words = append(words, stem)
	}
	return words
}
