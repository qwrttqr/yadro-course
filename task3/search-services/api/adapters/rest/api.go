package rest

import (
	"context"
	"encoding/json"
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
