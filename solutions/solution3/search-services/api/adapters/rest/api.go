package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"yadro.com/course/api/core"
)

type WordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

func NewWordsHandler(log *slog.Logger, norm core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			log.Error("missing phrase")
			http.Error(w, "missing phrase", http.StatusBadRequest)
			return
		}
		words, err := norm.Norm(r.Context(), phrase)
		if err != nil {
			log.Error("bad reply from normalizer", "error", err)
			if errors.Is(err, core.ErrBadArguments) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		reply := WordsResponse{
			Words: words,
			Total: len(words),
		}
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}

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
				log.Error("one of services is not available", "service", name)
				continue
			}
			reply.Replies[name] = "ok"
		}
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}
