package words

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	log    *slog.Logger
	client wordspb.WordsClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	return nil, status.Error(codes.Internal, "implement me")
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	return nil, status.Error(codes.Internal, "implement me")
}

func (c Client) Ping(ctx context.Context) error {
	return status.Error(codes.Internal, "implement me")
}
