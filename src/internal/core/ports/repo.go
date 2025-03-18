package ports

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"go-chassis/src/internal/adapters/storage"
)

type Repo interface {
	CheckupIncut(checkup func() error) storage.Request
	HandlerIncut(handler func()) storage.Request
	RequestExample(id uuid.UUID, out *[]string) storage.Request
}

type ToutboxRepo interface {
	Push(id *uuid.UUID, created *time.Time, date *[]byte) storage.Request
	Pull(chunkSize *int, out *[]storage.ToutboxPayload) storage.Request
	MarkAsTransferred(ids *[]uuid.UUID, transferred *time.Time) storage.Request
}

type Transactor func(ctx context.Context, isolationLevel sql.IsolationLevel, requests ...storage.Request) error
