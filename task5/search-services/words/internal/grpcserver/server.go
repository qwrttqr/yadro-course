package grpcserver

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type wordsService interface {
	Normalize(string) []string
}

type WordsServer struct {
	wordspb.UnimplementedWordsServer
	wordsService wordsService
	logger       *slog.Logger
}

func NewWordsServer(wordsService wordsService, logger *slog.Logger) *WordsServer {
	return &WordsServer{wordsService: wordsService, logger: logger}
}

func (s *WordsServer) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *WordsServer) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len(in.GetPhrase()) > 4096 {
		return nil, status.Error(codes.ResourceExhausted, "phrase too long")
	}
	wordsReply := wordspb.WordsReply{Words: s.wordsService.Normalize(in.GetPhrase())}
	return &wordsReply, nil
}
