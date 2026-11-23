package service

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
)

type WordsNormalizationService struct {
}

func NewWordsNormalizationService() *WordsNormalizationService {
	return &WordsNormalizationService{}
}

func (p *WordsNormalizationService) Normalize(input string) []string {
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
