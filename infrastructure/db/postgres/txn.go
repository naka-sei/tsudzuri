package postgres

import (
	"context"
	"fmt"

	"github.com/naka-sei/tsudzuri/infrastructure/db/ent"
)

type TxnProvider interface {
	BeginTx(ctx context.Context, client *ent.Client) (*ent.Client, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error
}

type defaultTxnProvider struct {
	txClient *ent.Tx
}

// NewDefaultTxnProvider creates a new defaultTxnProvider.
func NewDefaultTxnProvider() TxnProvider {
	return &defaultTxnProvider{}
}

// BeginTx begins a new transaction and returns the sql.Tx.
func (p *defaultTxnProvider) BeginTx(ctx context.Context, client *ent.Client) (*ent.Client, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	p.txClient = tx
	return tx.Client(), nil
}

// CommitTx commits the current transaction.
func (p *defaultTxnProvider) CommitTx(ctx context.Context) error {
	if p.txClient == nil {
		return fmt.Errorf("no transaction to commit")
	}
	if err := p.txClient.Commit(); err != nil {
		return err
	}
	return nil
}

// RollbackTx rolls back the current transaction.
func (p *defaultTxnProvider) RollbackTx(ctx context.Context) error {
	if p.txClient == nil {
		return fmt.Errorf("no transaction to rollback")
	}
	if err := p.txClient.Rollback(); err != nil {
		return err
	}
	return nil
}
