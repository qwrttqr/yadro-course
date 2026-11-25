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

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()
		replies := make(map[string]string, len(pingers))

		for name, pinger := range pingers {
			if requestCtx.Err() != nil {
				return
			}

			pingCtx, cancel := context.WithTimeout(requestCtx, 5*time.Second)
			err := pinger.Ping(pingCtx)
			cancel()

			if err != nil {
				replies[name] = "unavailable"
				log.ErrorContext(requestCtx, "service ping failed", "service", name, "error", err)
				continue
			}

			replies[name] = "ok"
		}

		if requestCtx.Err() != nil {
			return
		}

		response := struct {
			Replies map[string]string `json:"replies"`
		}{Replies: replies}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.ErrorContext(requestCtx, "writing ping response failed", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()
		err := updater.Update(requestCtx)
		if requestCtx.Err() != nil {
			return
		}
		if err != nil {
			if errors.Is(err, core.ErrAlreadyExists) {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.ErrorContext(requestCtx, "call update failed", "error", err)
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()
		statsCtx, cancel := context.WithTimeout(requestCtx, 5*time.Second)
		stats, err := updater.Stats(statsCtx)
		cancel()
		if requestCtx.Err() != nil {
			return
		}
		if err != nil {
			log.ErrorContext(requestCtx, "call update failed", "error", err)
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}
		response := struct {
			WordsTotal    int `json:"words_total"`
			WordsUnique   int `json:"words_unique"`
			ComicsFetched int `json:"comics_fetched"`
			ComicsTotal   int `json:"comics_total"`
		}{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.ErrorContext(statsCtx, "writing ping response failed", "error", err)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()
		statusCtx, cancel := context.WithTimeout(requestCtx, 5*time.Second)
		status, err := updater.Status(statusCtx)
		cancel()
		if requestCtx.Err() != nil {
			return
		}
		if err != nil {
			log.ErrorContext(requestCtx, "call update failed", "error", err)
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}
		response := struct {
			Status string `json:"status"`
		}{
			Status: string(status),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.ErrorContext(statusCtx, "writing ping response failed", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := updater.Drop(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.ErrorContext(ctx, "drop comics failed", "error", err)
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
