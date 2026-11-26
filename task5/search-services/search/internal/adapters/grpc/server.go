package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type SearchService interface {
	Search([]string) []string
}

type SearchServer struct {
	wordspb.UnimplementedWordsServer
	searchService SearchService
	logger        *slog.Logger
}

func NewSearchServer(searchService SearchService, logger *slog.Logger) *SearchServer {
	return &SearchServer{searchService: searchService, logger: logger}
}

func (s *SearchServer) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *SearchServer) Search(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
}
