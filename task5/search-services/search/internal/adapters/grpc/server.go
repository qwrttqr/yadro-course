package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/internal/models"
	"yadro.com/course/search/internal/service"
)

type SearchService interface {
	Search(ctx context.Context, words string, limit int) (models.SearchResults, error)
}

type SearchServer struct {
	searchpb.UnimplementedSearchServer
	searchService SearchService
	logger        *slog.Logger
}

func NewSearchServer(searchService SearchService, logger *slog.Logger) *SearchServer {
	return &SearchServer{searchService: searchService, logger: logger}
}

func (s *SearchServer) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *SearchServer) Search(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	phrase := in.GetPhrase()
	limit := in.GetLimit()
	res, err := s.searchService.Search(ctx, phrase, int(limit))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBadArguments):
			return nil, status.Error(codes.InvalidArgument, "phrase must not be empty")
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, status.FromContextError(err).Err()
		default:
			return nil, status.Error(codes.Internal, "internal service error")
		}
	}
	pbRes := make([]*searchpb.SearchReplyUnit, 0, len(res.Result))
	for _, unit := range res.Result {
		pbRes = append(pbRes, &searchpb.SearchReplyUnit{
			Id:  int64(unit.ID),
			Url: unit.URL,
		})
	}
	return &searchpb.SearchReply{Results: pbRes, Total: int32(res.Total)}, nil
}
