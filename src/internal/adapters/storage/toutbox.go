package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	pg "github.com/lib/pq"
	"github.com/pkg/errors"
)

type tableName string

const (
	toutboxExampleTable = "toutbox.toutbox_template"
)

type ToutboxPayload struct {
	IdempotenceKey uuid.UUID       `db:"idempotence_key"`
	Data           json.RawMessage `db:"data"`
	Created        time.Time       `db:"created"`
	Transferred    *time.Time      `db:"transferred"`
}

type ToutboxExample struct {
	transactionalOutBoxStorageTemplate
}

func (t ToutboxExample) tableName() tableName { return toutboxExampleTable }

func (t ToutboxExample) Push(id *uuid.UUID, created *time.Time, data *[]byte) Request {
	return t.push(t.tableName, &ToutboxPayload{
		IdempotenceKey: *id,
		Data:           *data,
		Created:        *created,
		Transferred:    nil,
	})
}

func (t ToutboxExample) Pull(chunkSize *int, out *[]ToutboxPayload) Request {
	return t.pull(
		t.tableName,
		chunkSize,
		out,
	)
}

func (t ToutboxExample) MarkAsTransferred(ids *[]uuid.UUID, transferred *time.Time) Request {
	return t.markAsTransferred(
		t.tableName,
		ids,
		transferred,
	)
}

type transactionalOutBoxStorageTemplate struct{}

func (t transactionalOutBoxStorageTemplate) push(tableName func() tableName, payload *ToutboxPayload) Request {
	return func(ctx context.Context, tx *sqlx.Tx) error {
		// TODO: or copy-paste is better?
		query := fmt.Sprintf(`
			INSERT INTO %s 
			(
				"idempotence_key",
				"data",
				"created",
				"transferred"
			) VALUES ($1,$2,$3,$4)`,
			tableName(),
		)

		args := []any{
			payload.IdempotenceKey,
			payload.Data,
			payload.Created,
			payload.Transferred,
		}

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return errors.Wrap(err, "failed to get execute request")
		}

		return nil
	}
}

func (t transactionalOutBoxStorageTemplate) pull(tableName func() tableName, chunkSize *int, out *[]ToutboxPayload) Request {
	return func(ctx context.Context, tx *sqlx.Tx) error {
		// TODO: or copy-paste is better?
		// TODO: some logic leaking with parameter IS NULL
		query := fmt.Sprintf(`
			SELECT 
				* 
			FROM %s 
			WHERE transferred IS NULL 
			ORDER BY created ASC
			LIMIT $1`,
			tableName(),
		)
		args := []any{
			chunkSize,
		}

		err := tx.SelectContext(ctx, out, query, args...)
		if err != nil {
			return errors.Wrap(err, "failed to get execute request")
		}
		return nil
	}
}

func (transactionalOutBoxStorageTemplate) markAsTransferred(tableName func() tableName, idempotenceKeys *[]uuid.UUID, transferred *time.Time) Request {
	return func(ctx context.Context, tx *sqlx.Tx) error {
		query := fmt.Sprintf(`
			UPDATE 
				%s 
			SET 
				"transferred" = $1 
			WHERE "idempotence_key" = ANY($2)`,
			tableName(),
		)
		args := []any{
			transferred,
			pg.Array(*idempotenceKeys),
		}

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return errors.Wrap(err, "failed to get execute request")
		}
		return nil
	}
}
