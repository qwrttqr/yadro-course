package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/api/core"
)

type WordsNormalizerResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

type WordsClient interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

func NewWordsHandler(log *slog.Logger, wordsClient WordsClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()

		if requestCtx.Err() != nil {
			return
		}

		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "phrase is required", http.StatusBadRequest)
			return
		}
		wordsRequestCtx, cancel := context.WithTimeout(requestCtx, 5*time.Second)
		response, err := wordsClient.Norm(wordsRequestCtx, phrase)
		cancel()
		if err != nil {
			if errors.Is(err, core.ErrBadArguments) {
				http.Error(w, "phrase is too long", http.StatusBadRequest)
				return
			}
			log.Error("words normalization service error", "error", err)
			http.Error(w, "words normalization service error", http.StatusInternalServerError)
			return
		}

		if response == nil {
			response = []string{}
		}

		result := WordsNormalizerResponse{
			Words: response,
			Total: len(response),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.ErrorContext(requestCtx, "writing ping response failed", "error", err)
		}
	}
}
