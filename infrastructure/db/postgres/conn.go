package postgres

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/naka-sei/tsudzuri/pkg/testdb"
)

func SetupTestDBConnection(t *testing.T) *Connection {
	t.Helper()

	const schemaName = "tsudzuri"

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	execDB, err := testdb.Open()
	if err != nil {
		t.Fatalf("failed to open testdb: %v", err)
	}

	success := false
	defer func() {
		if !success {
			_ = execDB.Close()
		}
	}()

	conn, err := NewConnection(dsn, dsn, schemaName, WithReadDB(execDB), WithWriteDB(execDB))
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	success = true

	t.Cleanup(func() {
		if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Fatalf("failed to close connection: %v", err)
		}
	})

	return conn
}
