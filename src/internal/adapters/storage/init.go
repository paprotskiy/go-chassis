package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"go-chassis/src/internal/adapters/cfg"
)

const (
	folderName              = "src/internal/adapters/persistent/migrations"
	defaultDb               = "postgres"
	pgCodeDuplacateDatabase = "42P04"
)

func NewDbProvider(c *cfg.Config) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PG.Host, c.PG.Port, c.PG.User, c.PG.Password, c.PG.DbName, c.PG.SslMode)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open sql conn to database with error")
	}

	db.SetMaxOpenConns(c.PG.ConnPool.MaxOpenConns)
	db.SetMaxIdleConns(c.PG.ConnPool.IdleConns)

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping db conn")
	}

	return db, nil
}

func pgDbAlreadyExistsErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), pgCodeDuplacateDatabase)
}

func ApplyMigrations(cfg *cfg.Config, dbProvider *sqlx.DB) error {
	driver, err := postgres.WithInstance(dbProvider.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres abstraction :: %v", err)
	}

	relativePathKey := "file://"
	srcUrl := relativePathKey + folderName
	migrator, err := migrate.NewWithDatabaseInstance(srcUrl, cfg.PG.DbName, driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator instance :: %v", err)
	}

	err = migrator.Up()
	if err != nil && err.Error() != "no change" {
		return fmt.Errorf("failed to apply migrations :: %v", err)
	}

	return nil
}

func CreateDatabaseIfNotExists(c *cfg.Config) error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PG.Host,
		c.PG.Port,
		c.PG.User,
		c.PG.Password,
		c.PG.DbNameDefault,
		c.PG.SslMode)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		panic(fmt.Errorf("failed to open sql conn to database with error :: %v", err))
	}
	defer db.Close()

	query := fmt.Sprintf(`create database "%s"`, c.PG.DbName)

	_, err = db.Exec(query)
	if err == nil {
		return nil
	}
	if pgDbAlreadyExistsErr(err) {
		return nil
	}
	if strings.HasPrefix(err.Error(), "pq: database") &&
		strings.HasSuffix(err.Error(), "already exists") {
		return nil
	}

	return errors.Wrap(err, "failed to create database if not exists")
}

type Request func(context.Context, *sqlx.Tx) error

type Transactor func(ctx context.Context, isolationLevel sql.IsolationLevel, requests ...Request) error

func transaction(ctx context.Context, db *sqlx.DB, isolationLevel sql.IsolationLevel, readOnly bool, requests ...Request) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: isolationLevel,
		ReadOnly:  readOnly,
	})

	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for idx, req := range requests {
		if err = req(ctx, tx); err != nil {
			return errors.Wrapf(err, "failed to execute request #%v", idx)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// TODO: test
func NewCommandProvider(db *sqlx.DB) func(ctx context.Context, isolationLevel sql.IsolationLevel, requests ...Request) error {
	return func(ctx context.Context, isolationLevel sql.IsolationLevel, requests ...Request) error {
		return transaction(ctx, db, isolationLevel, false, requests...)
	}
}

// TODO: test
func NewQueryProvider(db *sqlx.DB) func(ctx context.Context, isolationLevel sql.IsolationLevel, requests ...Request) error {
	return func(ctx context.Context, isolationLevel sql.IsolationLevel, requests ...Request) error {
		return transaction(ctx, db, isolationLevel, true, requests...)
	}
}

type manualTransactor interface {
	Commit() error
	Rollback() error
}

func StartManCommand(ctx context.Context, db *sqlx.DB, isolationLevel sql.IsolationLevel) (manualTransactor, error) {
	return startManTransaction(ctx, db, isolationLevel, false)
}

func StartManQuery(ctx context.Context, db *sqlx.DB, isolationLevel sql.IsolationLevel) (manualTransactor, error) {
	return startManTransaction(ctx, db, isolationLevel, true)
}

func startManTransaction(ctx context.Context, db *sqlx.DB, isolationLevel sql.IsolationLevel, readOnly bool) (manualTransactor, error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: isolationLevel,
		ReadOnly:  readOnly,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin txx with provider")
	}

	return tx, nil
}
