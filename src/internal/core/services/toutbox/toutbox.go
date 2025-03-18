package toutbox

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"go-chassis/src/internal/adapters/storage"
	"go-chassis/src/internal/core/ports"
)

func NewToutboxMsgRelay(
	logger ports.Logger,
	scrapeCapacity int,
	scrapingPeriod time.Duration,
	commandProvider ports.Transactor,
	queryProvider ports.Transactor,
	trepo ports.ToutboxRepo,
	msgBroker ports.MsgBroker,
) *toutBoxMsgRelay {
	return &toutBoxMsgRelay{
		logger:          logger,
		scrapeCapacity:  scrapeCapacity,
		scrapingPeriod:  scrapingPeriod,
		commandProvider: commandProvider,
		queryProvider:   queryProvider,
		trepo:           trepo,
		msgBroker:       msgBroker,
	}
}

type toutBoxMsgRelay struct {
	logger          ports.Logger
	scrapeCapacity  int
	scrapingPeriod  time.Duration
	commandProvider ports.Transactor
	queryProvider   ports.Transactor
	trepo           ports.ToutboxRepo
	msgBroker       ports.MsgBroker
}

func (t toutBoxMsgRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(t.scrapingPeriod).C // TODO extract for testability

	skipTicker := make(chan struct{}, 1)

	for {
		var (
			err     error
			handled *int
		)

		select {
		case <-ctx.Done():
			return
		case <-skipTicker:
			t.logger.Debug("skipped waiting")
			handled, err = t.transfer(ctx)
		case <-ticker:
			t.logger.Debug("by tick")
			handled, err = t.transfer(ctx)
		}

		if err != nil {
			t.logger.Error(errors.Wrap(err, "failed to transfer data from toutbox to msg broker"))
			continue
		}

		if *handled == t.scrapeCapacity {
			t.logger.Debug(`scraped`,
				slog.Int("handled", *handled),
				slog.Int("capacity", t.scrapeCapacity),
			)

			select {
			case skipTicker <- struct{}{}:
			default:
			}
		}

	}
}

func (t toutBoxMsgRelay) transfer(ctx context.Context) (*int, error) {
	// warn: service knows something about repo implementation, but it is the best among compromises
	pulled := new([]storage.ToutboxPayload)
	err := t.queryProvider(ctx, sql.LevelSerializable,
		t.trepo.Pull(&t.scrapeCapacity, pulled),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ")
	}

	for idx, record := range *pulled {
		err := t.msgBroker.PushToBroker(record.Data)
		if err != nil {
			return nil, errors.Wrapf(err, `failed to push record #%d to message-broker`, idx)
		}
	}

	ids := Map(*pulled, func(in storage.ToutboxPayload) uuid.UUID {
		return in.IdempotenceKey
	})

	now := time.Now()
	err = t.commandProvider(ctx, sql.LevelReadCommitted,
		t.trepo.MarkAsTransferred(ref(ids), ref(now)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to mark records as transferred")
	}

	return ref(len(ids)), nil
}
