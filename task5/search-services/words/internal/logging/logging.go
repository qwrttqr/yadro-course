package logging

import (
	"io"
	"log/slog"
)

func New(levelStr string, file io.Writer) (*slog.Logger, error) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(levelStr)); err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: logLevel})
	return slog.New(handler), nil
}
