package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"yadro.com/course/api/core"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := PingResponse{
			Replies: make(map[string]string),
		}
		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				reply.Replies[name] = "unavailable"
				log.Error("one of services is not available", "service", name, "error", err)
				continue
			}
			reply.Replies[name] = "ok"
		}
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			log.Error("error while update", "error", err)
			if errors.Is(err, core.ErrAlreadyExists) {
				http.Error(w, err.Error(), http.StatusAccepted)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type UpdateStats struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("error while stats", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reply := UpdateStats{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}
		if err = json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("encoding failed", "error", err)
		}
	}
}

type UpdateStatus struct {
	Status string `json:"status"`
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("error while status", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reply := UpdateStatus{Status: string(status)}
		if err = json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("encoding failed", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("error while drop", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
