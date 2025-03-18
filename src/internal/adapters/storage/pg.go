package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"go-chassis/src/internal/adapters/storage/plain"
)

type Storage struct{}

func (Storage) RequestExample(id uuid.UUID, out *[]string) Request {
	return func(ctx context.Context, tx *sqlx.Tx) (err error) {
		*out, err = plain.RequestExample(ctx, tx, id)
		return
	}
}
