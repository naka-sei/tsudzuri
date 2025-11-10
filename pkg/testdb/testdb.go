package testdb

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kanmu/pgtxdb"

	"github.com/naka-sei/tsudzuri/pkg/uuid"
)

func init() {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		return
	}

	pgtxdb.Register("pgtx", "pgx", dsn)
}

func Open() (*sql.DB, error) {
	uid := uuid.NewV7()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		panic("TEST_DATABASE_DSN is not set")
	}

	dsn = fmt.Sprintf("%s key=%s", dsn, uid.String())

	db, err := sql.Open("pgtx", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
