package core

import (
	"context"
)

type Searcher interface {
	Search(ctx context.Context, phrase string, limit int) ([]Comics, error)
}

type DB interface {
	Search(ctx context.Context, keyword string) ([]int, error)
	Get(ctx context.Context, ID int) (Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}
