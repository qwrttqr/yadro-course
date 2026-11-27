package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

func NewServer(service core.Updater) *Server {
	return &Server{service: service}
}

type Server struct {
	updatepb.UnimplementedUpdateServer
	service core.Updater
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Status(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatusReply, error) {
	reply := &updatepb.StatusReply{
		Status: updatepb.Status_STATUS_UNSPECIFIED,
	}
	switch s.service.Status(ctx) {
	case core.StatusIdle:
		reply.Status = updatepb.Status_STATUS_IDLE
	case core.StatusRunning:
		reply.Status = updatepb.Status_STATUS_RUNNING
	}
	return reply, nil
}

func (s *Server) Update(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.service.Update(ctx)
	switch {
	case err == nil:
		return &emptypb.Empty{}, nil
	case errors.Is(err, core.ErrAlreadyExists):
		return nil, status.Error(codes.AlreadyExists, "update already running")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return nil, status.FromContextError(err).Err()
	default:
		return nil, status.Error(codes.Internal, "update failed")
	}
}

func (s *Server) Stats(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatsReply, error) {
	stats, err := s.service.Stats(ctx)
	switch {
	case err == nil:
		return &updatepb.StatsReply{
			ComicsTotal:   int64(stats.ComicsTotal),
			ComicsFetched: int64(stats.ComicsFetched),
			WordsTotal:    int64(stats.WordsTotal),
			WordsUnique:   int64(stats.WordsUnique),
		}, nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return nil, status.FromContextError(err).Err()
	default:
		return nil, status.Error(codes.Internal, "getting stats failed")
	}
}

func (s *Server) Drop(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.service.Drop(ctx)
	switch {
	case err == nil:
		return &emptypb.Empty{}, err
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return nil, status.FromContextError(err).Err()
	default:
		return nil, status.Error(codes.Internal, "drop failed")
	}
}
