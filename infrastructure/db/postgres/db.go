package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/naka-sei/tsudzuri/infrastructure/db/ent"
)

// conn is an interface that represents a database connection that can begin transactions.
type conn struct {
	db        *sql.DB
	entClient *ent.Client
	dialect   string
	debug     bool
}

type txnCtxKey struct{}

// Connection returns a new conn instance.
type Connection struct {
	read  *conn
	write *conn
}

// ReadOnlyDB returns the read-only ent.Client.
func (c *Connection) ReadOnlyDB(ctx context.Context) *ent.Client {
	if provider, ok := getTransactionFromContext(ctx); ok {
		if client := provider.Client(); client != nil {
			return client
		}
	}

	return c.read.entClient
}

// WriteDB returns the write ent.Client.
func (c *Connection) WriteDB(ctx context.Context) *ent.Client {
	if provider, ok := getTransactionFromContext(ctx); ok {
		if client := provider.Client(); client != nil {
			return client
		}
	}

	return c.write.entClient
}

type ConnectionOption func(*Connection)

// WithDebug enables debug mode for the connection.
func WithDebug() ConnectionOption {
	return func(c *Connection) {
		// Enable debug mode for both read and write connections.
		c.read.debug = true
		c.write.debug = true
		c.read.entClient = c.read.entClient.Debug()
		c.write.entClient = c.write.entClient.Debug()
	}
}

// WithReadDB sets the read connection to be used for read-only operations.
func WithReadDB(db *sql.DB) ConnectionOption {
	return func(c *Connection) {
		c.read.db = db
		c.read.entClient = ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
		c.read.dialect = dialect.Postgres
		if c.read.debug {
			c.read.entClient = c.read.entClient.Debug()
		}
	}
}

// WithWriteDB sets the write connection to be used for write operations.
func WithWriteDB(db *sql.DB) ConnectionOption {
	return func(c *Connection) {
		c.write.db = db
		c.write.entClient = ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
		c.write.dialect = dialect.Postgres
		if c.write.debug {
			c.write.entClient = c.write.entClient.Debug()
		}
	}
}

func addSearchPath(dsn, schema string) string {
	if strings.Contains(dsn, "search_path") {
		return dsn
	}
	return fmt.Sprintf("%s search_path=%s", dsn, schema)
}

func NewConnection(readDB, writeDB, schema string, opts ...ConnectionOption) (*Connection, error) {
	readDSN := addSearchPath(readDB, schema)
	writeDSN := addSearchPath(writeDB, schema)

	read, err := sql.Open("pgx", readDSN)
	if err != nil {
		return nil, err
	}

	write, err := sql.Open("pgx", writeDSN)
	if err != nil {
		return nil, err
	}

	read.SetMaxOpenConns(20)
	read.SetMaxIdleConns(20)
	read.SetConnMaxIdleTime(time.Second * 5)
	read.SetConnMaxLifetime(time.Second * 10)

	write.SetMaxOpenConns(20)
	write.SetMaxIdleConns(20)
	write.SetConnMaxIdleTime(time.Second * 5)
	write.SetConnMaxLifetime(time.Second * 10)

	err = read.Ping()
	if err != nil {
		return nil, err
	}

	err = write.Ping()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		read:  &conn{db: read, entClient: ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, read))), dialect: dialect.Postgres},
		write: &conn{db: write, entClient: ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, write))), dialect: dialect.Postgres},
	}

	for _, opt := range opts {
		opt(conn)
	}

	return conn, nil
}

// getTransactionFromContext retrieves the current transaction provider from the context.
func getTransactionFromContext(ctx context.Context) (TxnProvider, bool) {
	provider, ok := ctx.Value(txnCtxKey{}).(TxnProvider)
	return provider, ok
}

// setTransactionInContext sets the current transaction provider in the context.
func setTransactionInContext(ctx context.Context, provider TxnProvider) context.Context {
	return context.WithValue(ctx, txnCtxKey{}, provider)
}

// RunInTransaction runs the given function within a database transaction.
func (c *Connection) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if there's already a transaction in the context.
	if provider, ok := getTransactionFromContext(ctx); ok {
		savepoint, err := provider.BeginSavepoint(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if r := recover(); r != nil {
				_ = provider.RollbackToSavepoint(ctx, savepoint)
				panic(r)
			}
		}()
		if err := fn(ctx); err != nil {
			if rbErr := provider.RollbackToSavepoint(ctx, savepoint); rbErr != nil {
				return fmt.Errorf("savepoint rollback error: %v, original error: %w", rbErr, err)
			}
			return err
		}
		if err := provider.ReleaseSavepoint(ctx, savepoint); err != nil {
			return err
		}
		return nil
	}

	provider := NewDefaultTxnProvider()
	if err := provider.BeginTx(ctx, c.write); err != nil {
		return err
	}
	ctxWithTx := setTransactionInContext(ctx, provider)

	defer func() {
		if r := recover(); r != nil {
			_ = provider.RollbackTx(ctx)
			panic(r)
		}
	}()

	if err := fn(ctxWithTx); err != nil {
		rerr := provider.RollbackTx(ctx)
		if rerr != nil {
			return fmt.Errorf("transaction rollback error: %v, original error: %w", rerr, err)
		}
		return err
	}

	if cerr := provider.CommitTx(ctx); cerr != nil {
		return fmt.Errorf("transaction commit error: %w", cerr)
	}

	return nil
}

// Close closes the database connections.
func (c *Connection) Close() error {
	if cerr := c.read.entClient.Close(); cerr != nil {
		return cerr
	}
	if err := c.read.db.Close(); err != nil {
		return err
	}

	if cerr := c.write.entClient.Close(); cerr != nil {
		return cerr
	}
	if err := c.write.db.Close(); err != nil {
		return err
	}
	return nil
}

func PtrInt32ToInt(p *int32) int {
	if p == nil {
		return 0
	}
	return int(*p)
}
