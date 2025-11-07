package service

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_transaction/transaction.go -source=./transaction.go -package=mocktransaction
type TransactionService interface {
	// RunInTransaction runs the given function within a database transaction.
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
