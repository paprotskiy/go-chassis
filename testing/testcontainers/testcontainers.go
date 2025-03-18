package testcontainers

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dockerImage = "postgres:17-alpine"
	exposedPort = "5432"
)

func NewInstance(
	ctx context.Context,
	pgUser, pgPassword string,
	pgDbName, pgSslMode string,
) (*sqlx.DB, func(), error) {
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        dockerImage,
			ExposedPorts: []string{exposedPort},
			Env: map[string]string{
				"POSTGRES_USER":     pgUser,
				"POSTGRES_PASSWORD": pgPassword,
				"POSTGRES_DB":       pgDbName,
			},
			WaitingFor: wait.
				ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		},
		Started: true,
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to start container")
	}

	mappedHost, err := postgresC.Host(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get container host")
	}
	mappedPort, err := postgresC.MappedPort(ctx, exposedPort)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get mapped port")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		mappedHost,
		mappedPort.Port(),
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
		db.Close()
		_ = postgresC.Terminate(ctx)
	}

	return db, cleanup, nil
}
