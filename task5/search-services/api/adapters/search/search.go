package search

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	client searchpb.SearchClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := searchpb.NewSearchClient(conn)
	return &Client{client: client}, nil
}

func (c Client) Search(ctx context.Context, phrase string, limit int) (core.SearchResult, error) {
	res, err := c.client.Search(ctx, &searchpb.SearchRequest{Phrase: phrase, Limit: int32(limit)})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return core.SearchResult{}, core.ErrBadArguments
		}
		return core.SearchResult{}, fmt.Errorf("search comics: %w", err)
	}
	total := res.GetTotal()
	results := make([]core.SearchResultUnit, 0, total)
	for _, entry := range res.GetResults() {
		results = append(results, core.SearchResultUnit{
			ID:  int(entry.GetId()),
			URL: entry.GetUrl(),
		})
	}

	return core.SearchResult{Results: results, Total: total}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	return nil
}
