package ingest

import (
	"context"
	"ningen/domain"
)

type Source interface {
	Stream(ctx context.Context, out chan<- domain.Item) error
}
