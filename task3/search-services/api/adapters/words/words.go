package words

import (
	"context"
	"log/slog"

	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	log    *slog.Logger
	client wordspb.WordsClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	return nil, nil
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	return nil, nil
}

func (c Client) Ping(ctx context.Context) error {
	return nil
}
