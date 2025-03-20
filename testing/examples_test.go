// code snippets for integrational testing
package testing

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"go-chassis/testing/embedded"
	"go-chassis/testing/testcontainers"
)

const (
	freePortForEmbeddedDb uint32 = 5555

	pgUser     = "user-test"
	pgPassword = "password-test"
	pgDbName   = "dbname-test"
	pgSslMode  = "disable"

	// dbImage = "postgres:17-alpine"
	dbImage           = "my-postgres-unprivileged:latest"
	dbExposedPort     = "5432"
	dbReuseMode       = true
	dbReuseModeDbName = "back-scratch-testdb"
)

func Test_TestContainers_Snippet(t *testing.T) {
	for idx := range 10 {
		start := time.Now()
		db, cleanup, err := testcontainers.NewInstance(
			context.Background(),
			&testcontainers.MockParams{
				TestDbImage:               dbImage,
				TestDbPort:                "",
				ReuseMode:                 dbReuseMode,
				SupressCleanupInReuseMode: false,
				ReuseDbName:               dbReuseModeDbName,
			},
			&testcontainers.DbConnParams{
				PgUser:     pgUser,
				PgPassword: pgPassword,
				PgDbName:   pgDbName,
				PgSslMode:  pgSslMode,
			},
		)
		if err != nil {
			t.Log(errors.Wrap(err, "failed to set up postgres"))
			return
		}
		if false {
			defer cleanup()
		}

		var now time.Time

		err = db.Get(&now, "SELECT NOW()")
		if err != nil {
			t.Fatalf("failed to query database: %v", err)
		}

		t.Log(idx, ": ", time.Since(start).Milliseconds())
	}

	t.Error("knock out test for seeing logs")
}

func Test_EmbeddedSnippet(t *testing.T) {
	for idx := range 10 {
		start := time.Now()

		db, cleanup, err := embedded.NewInstance(
			"localhost",
			freePortForEmbeddedDb,
			pgUser,
			pgPassword,
			pgDbName,
			pgSslMode,
			embedded.LoggerStub(),
		)
		if err != nil {
			t.Fatalf("failed to create connection to embedded db")
		}

		var now time.Time
		err = db.Get(&now, "SELECT NOW()")
		if err != nil {
			t.Fatalf("failed to query database: %v", err)
		}

		t.Log(idx, ": ", time.Since(start).Milliseconds())
		cleanup()
	}

	t.Fatal("knock out test for seeing logs")
}
