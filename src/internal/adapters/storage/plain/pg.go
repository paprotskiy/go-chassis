package plain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func RequestExample(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (res []string, err error) {
	query := `select * from table where "id" = $1`
	values := []any{
		id,
	}

	err = tx.SelectContext(ctx, res, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to get execute request :: %v", err)
	}

	return res, nil
}
