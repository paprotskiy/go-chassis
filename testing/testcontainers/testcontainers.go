package testcontainers

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DbConnParams struct {
	PgUser     string
	PgPassword string
	PgDbName   string
	PgSslMode  string
}

type MockParams struct {
	TestDbImage               string
	TestDbPort                string
	ReuseMode                 bool
	SupressCleanupInReuseMode bool
	ReuseContainerName        string
}

func NewInstance(
	ctx context.Context,
	mock *MockParams,
	conn *DbConnParams,
) (*sqlx.DB, func(), error) {

	if mock.ReuseMode && mock.ReuseContainerName == "" {
		return nil, nil, errors.Errorf("reuseDbName must be specified in reuseMode")
	}

	uniqueName := ""
	if mock.ReuseMode {
		uniqueName = mock.ReuseContainerName
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        mock.TestDbImage,
			Name:         uniqueName,
			ExposedPorts: []string{mock.TestDbPort},
			Env: map[string]string{
				"POSTGRES_USER":     conn.PgUser,
				"POSTGRES_PASSWORD": conn.PgPassword,
				"POSTGRES_DB":       conn.PgDbName,
			},
			WaitingFor: wait.
				ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		},
		Started: true,
		Reuse:   mock.ReuseMode,
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to start container")
	}

	mappedHost, err := postgresC.Host(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get container host")
	}
	mappedPort, err := postgresC.MappedPort(ctx, nat.Port(mock.TestDbPort))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get mapped port")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		mappedHost,
		mappedPort.Port(),
		conn.PgUser,
		conn.PgPassword,
		conn.PgDbName,
		conn.PgSslMode,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to connect to database")
	}

	cleanup := func() {
		if mock.ReuseMode && mock.SupressCleanupInReuseMode {
			return
		}

		db.Close()
		_ = postgresC.Terminate(ctx)
	}

	return db, cleanup, nil
}
