package apiservice

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"go-chassis/src/internal/adapters/cfg"
	"go-chassis/src/internal/core/ports"
	"go-chassis/src/internal/core/services/apiservice/checkups"
)

func New(
	ctx context.Context,
	cfg *cfg.Config,
	logger ports.Logger,
	commandProvider ports.Transactor,
	queryProvider ports.Transactor,
	repo ports.Repo,
	trepo ports.ToutboxRepo,
) (*ApiService, error) {
	return &ApiService{
		repo:            repo,
		trepo:           trepo,
		commandProvider: commandProvider,
		queryProvider:   queryProvider,
	}, nil
}

type ApiService struct {
	repo            ports.Repo
	trepo           ports.ToutboxRepo
	commandProvider ports.Transactor
	queryProvider   ports.Transactor
}

func (d ApiService) SystemOperation(ctx context.Context, in any) (*struct{}, error) {
	now := time.Now()
	opID := uuid.New()
	data, _ := json.Marshal(in)

	err := d.commandProvider(ctx, sql.LevelSerializable,
		d.trepo.Push(&opID, &now, &data),
		// d.repo.CheckupIncut(checkups.AlwaysErr()),
	)
	if err != nil {
		if checkups.SameType[*checkups.CheckupErr](err) {
			return nil, errors.Errorf("caught !!!!!")
		}
		return nil, errors.Errorf("who knows what was that...")
	}

	return nil, nil
}
