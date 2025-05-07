package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"go-chassis/src/internal/adapters/cfg"
	"go-chassis/src/internal/adapters/logger"
	"go-chassis/src/internal/adapters/msgbroker"
	"go-chassis/src/internal/adapters/storage"
	"go-chassis/src/internal/core/services/apiservice"
	"go-chassis/src/internal/core/services/toutbox"
	"go-chassis/src/usecases/api"
)

var version = `will be filled after compilation thnx to "-X main.version" flag`

const (
	apiPort               = "3000"
	kafkaAddress          = ""              // TODO: must be configured
	toutboxChunkSize      = 50              // TODO: must be configured
	toutboxScrapingPeriod = 3 * time.Second // TODO: must be configured
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := cfg.NewConfig()
	if err != nil {
		panic(errors.Wrap(err, "critical: failed to read configuration"))
	}

	logLevel := logger.GetLogLevel(cfg.Debug)
	logger := logger.NewJsonLogger(ctx, os.Stdout, logLevel, version)
	logger.Info("application have started", "loglevel", logLevel)

	appShutdown := make(chan struct{})
	fatalErrs := make(chan error, 12)

	sigterm := make(chan os.Signal, 1)
	go gracefulShutdown(
		cancel,
		sigterm,
		fatalErrs,
		appShutdown,
		cfg.CoolingDown,
		logger,
	)

	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)

	createDbErr := storage.CreateDatabaseIfNotExists(cfg)
	fatalErrs <- errors.Wrap(createDbErr, "failed to create db")

	dbProvider, err := storage.NewDbProvider(cfg)
	fatalErrs <- errors.Wrap(err, "failed to get db provider")

	err = storage.ApplyMigrations(cfg, dbProvider)
	fatalErrs <- errors.Wrap(err, "failed to apply migrations")

	commandProvider := storage.NewCommandProvider(dbProvider)
	queryProvider := storage.NewQueryProvider(dbProvider)
	repos := struct {
		main struct {
			storage.Incut
			storage.Storage
		}
		toutbox storage.ToutboxExample
	}{}

	apiService, err := apiservice.New(
		ctx,
		cfg,
		logger,
		commandProvider,
		queryProvider,
		repos.main,
		repos.toutbox,
	)
	fatalErrs <- errors.Wrap(err, "failed to initialize domain server")

	msgbroker, err := msgbroker.New(kafkaAddress)
	fatalErrs <- errors.Wrap(err, "failed to create message broker")

	go toutbox.NewToutboxMsgRelay(
		logger,
		toutboxChunkSize,
		toutboxScrapingPeriod,
		commandProvider,
		queryProvider,
		repos.toutbox,
		msgbroker,
	).Run(ctx)

	go func() {
		err := api.ListenAndServe(fmt.Sprintf(":%s", apiPort), apiService, logLevel, version) // TODO: port as env
		fatalErrs <- errors.Wrap(err, "http server error")
	}()

	<-appShutdown
}

func gracefulShutdown(
	ctxCancel func(),
	sigterm chan os.Signal,
	fatalErrs chan error,
	appShutdown chan struct{},
	coolingDownInterval time.Duration,
	logger *logger.Logger,
) {
	run := true
	for run {
		select {
		case <-sigterm:
			logger.Info("sigterm signal received, starting graceful shutdown")
			run = false
		case err := <-fatalErrs:
			if err == nil {
				continue
			}

			logger.Error(errors.Wrap(err, "fatal error was received, starting graceful shutdown"))
			run = false
		}
	}

	ctxCancel()
	time.Sleep(coolingDownInterval)

	logger.Info("graceful shutdown completed")
	close(appShutdown)
}
