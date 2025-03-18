package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Incut struct{}

func (Incut) HandlerIncut(in func()) Request {
	return func(context.Context, *sqlx.Tx) error {
		in()
		return nil
	}
}

func (Incut) CheckupIncut(in func() error) Request {
	return func(context.Context, *sqlx.Tx) error {
		if err := in(); err != nil {
			return errors.Wrap(err, "error on checkup")
		}
		return nil
	}
}
