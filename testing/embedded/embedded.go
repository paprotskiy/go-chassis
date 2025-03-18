package embedded

import (
	"bytes"
	"fmt"
	"io"
	"time"

	. "github.com/fergusstrange/embedded-postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	timeout     = 15 * time.Second
	maxConn     = "200"
	runtimePath = "./embedded-db-vendor"
	binRepoURL  = "https://repo.local/central.proxy"
)

func LoggerStub() io.Writer {
	return &bytes.Buffer{}
}

func NewInstance(
	pgHost string,
	pgPort uint32,
	pgUser, pgPassword string,
	pgDbName, pgSslMode string,
	logger io.Writer,
) (*sqlx.DB, func(), error) {
	// parsedPort, err := strconv.ParseUint(pgPort, 10, 32)
	// if err != nil {
	// 	return nil, nil, errors.Wrap(err, "failed to parse port")
	// }

	postgres := NewDatabase(DefaultConfig().
		Username(pgUser).
		Password(pgPassword).
		Database(pgDbName).
		Version(V14).
		RuntimePath(runtimePath).
		// BinaryRepositoryURL(binRepoURL).
		Port(pgPort).
		StartTimeout(timeout).
		StartParameters(map[string]string{"max_connections": maxConn}).
		Logger(logger))
	if err := postgres.Start(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to start embedded postgres")
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pgHost,
		pgPort,
		pgUser,
		pgPassword,
		pgDbName,
		pgSslMode,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to connect to database")
	}

	cleanup := func() {
		err = postgres.Stop()
		if err != nil {
			err = errors.Wrap(err, "failed to stop embedded postgres")
			panic(err.Error())
		}
	}

	return db, cleanup, nil
}
