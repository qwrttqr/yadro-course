package logging

import (
	"io"
	"log/slog"
)

func New(levelStr string, file io.Writer) (logger *slog.Logger, err error) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(levelStr)); err != nil {
		return nil, err
	}
	h := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: logLevel})
	return slog.New(h), err
}
