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
	pgUser     = "user-test"
	pgPassword = "password-test"
	pgDbName   = "dbname-test"
	pgSslMode  = "disable"
)

var (
	freePort uint32 = 5555
)

func Test_TestContainers_Snippet(t *testing.T) {
	for idx := range 10 {
		start := time.Now()
		db, cleanup, err := testcontainers.NewInstance(
			context.Background(),
			pgUser,
			pgPassword,
			pgDbName,
			pgSslMode,
		)
		if err != nil {
			t.Log(errors.Wrap(err, "failed to set up postgres"))
			return
		}

		var now time.Time
		err = db.Get(&now, "SELECT NOW()")
		if err != nil {
			t.Fatalf("failed to query database: %v", err)
		}

		t.Log(idx, ": ", time.Since(start).Milliseconds())
		defer cleanup()
	}

	t.Fatal("knock out test for seeing logs")
}

func Test_EmbeddedSnippet(t *testing.T) {
	for idx := range 10 {
		start := time.Now()

		db, cleanup, err := embedded.NewInstance(
			"localhost",
			freePort,
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
