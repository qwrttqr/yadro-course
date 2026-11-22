package grpcserver

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	petnamepb "yadro.com/course/proto"
)

type generationService interface {
	Generate(words int64, separator string) string
}
type PetNameGeneratorServer struct {
	petnamepb.UnimplementedPetnameGeneratorServer
	generationService generationService
	logger            *slog.Logger
}

func NewPetNameGeneratorServer(gs generationService, logger *slog.Logger) *PetNameGeneratorServer {
	return &PetNameGeneratorServer{generationService: gs, logger: logger}
}

func (p *PetNameGeneratorServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (p *PetNameGeneratorServer) Generate(_ context.Context, req *petnamepb.PetnameRequest) (*petnamepb.PetnameResponse, error) {
	wordsCount := req.GetWords()
	if wordsCount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "words count must be positive")
	}
	name := p.generationService.Generate(wordsCount, req.GetSeparator())
	return &petnamepb.PetnameResponse{Name: name}, nil
}

func (p *PetNameGeneratorServer) GenerateMany(req *petnamepb.PetnameStreamRequest, res grpc.ServerStreamingServer[petnamepb.PetnameResponse]) error {
	wordsCount := req.GetWords()
	namesCount := req.GetNames()
	sep := req.GetSeparator()
	if wordsCount <= 0 {
		return status.Error(codes.InvalidArgument, "words count must be positive")
	}
	if namesCount <= 0 {
		return status.Error(codes.InvalidArgument, "names count must be positive")
	}

	var i int64 = 0
	for ; i < namesCount; i++ {
		name := p.generationService.Generate(wordsCount, sep)
		err := res.Send(&petnamepb.PetnameResponse{Name: name})
		if err != nil {
			p.logger.Error("writing response", "error", err)
			return err
		}
	}
	return nil
}
