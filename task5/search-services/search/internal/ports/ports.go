package ports

import (
	"context"

	"yadro.com/course/search/internal/models"
)

type DB interface {
	Search(context.Context, []string) (models.SearchResults, error)
}
