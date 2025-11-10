package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/naka-sei/tsudzuri/infrastructure/db/ent"
)

type TxnProvider interface {
	BeginTx(ctx context.Context, source *conn) error
	Client() *ent.Client
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error
	BeginSavepoint(ctx context.Context) (string, error)
	RollbackToSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
}

type defaultTxnProvider struct {
	tx           *sql.Tx
	client       *ent.Client
	savepointSeq int
}

// NewDefaultTxnProvider creates a new defaultTxnProvider.
func NewDefaultTxnProvider() TxnProvider {
	return &defaultTxnProvider{}
}

// BeginTx begins a new transaction and prepares a transactional ent.Client.
func (p *defaultTxnProvider) BeginTx(ctx context.Context, source *conn) error {
	if p.tx != nil {
		return fmt.Errorf("transaction already started")
	}
	if source == nil || source.db == nil {
		return errors.New("write connection is not configured")
	}
	if source.dialect == "" {
		source.dialect = dialect.Postgres
	}
	sqlTx, err := source.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	drv := entsql.NewDriver(source.dialect, entsql.Conn{ExecQuerier: sqlTx})
	client := ent.NewClient(ent.Driver(drv))
	if source.debug {
		client = client.Debug()
	}
	p.tx = sqlTx
	p.client = client
	return nil
}

// Client returns the transactional ent.Client, if any.
func (p *defaultTxnProvider) Client() *ent.Client {
	return p.client
}

// CommitTx commits the current transaction.
func (p *defaultTxnProvider) CommitTx(ctx context.Context) error {
	if p.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}
	err := p.tx.Commit()
	p.tx = nil
	p.client = nil
	p.savepointSeq = 0
	return err
}

// RollbackTx rolls back the current transaction.
func (p *defaultTxnProvider) RollbackTx(ctx context.Context) error {
	if p.tx == nil {
		return fmt.Errorf("no transaction to rollback")
	}
	err := p.tx.Rollback()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	p.tx = nil
	p.client = nil
	p.savepointSeq = 0
	return nil
}

// BeginSavepoint creates and returns a savepoint identifier within the current transaction.
func (p *defaultTxnProvider) BeginSavepoint(ctx context.Context) (string, error) {
	if p.tx == nil {
		return "", fmt.Errorf("no active transaction")
	}
	p.savepointSeq++
	name := fmt.Sprintf("sp_%04d", p.savepointSeq)
	if err := p.exec(ctx, fmt.Sprintf("SAVEPOINT %s", name)); err != nil {
		return "", err
	}
	return name, nil
}

// RollbackToSavepoint rolls back to the given savepoint within the current transaction.
func (p *defaultTxnProvider) RollbackToSavepoint(ctx context.Context, name string) error {
	return p.exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
}

// ReleaseSavepoint releases a previously created savepoint.
func (p *defaultTxnProvider) ReleaseSavepoint(ctx context.Context, name string) error {
	return p.exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name))
}

func (p *defaultTxnProvider) exec(ctx context.Context, stmt string) error {
	if p.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	if _, err := p.tx.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}
