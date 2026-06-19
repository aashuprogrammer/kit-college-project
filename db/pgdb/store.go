package pgdb

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(*Queries) error) error
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

// NewStore creates a new store
func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{
		Queries: New(db),
		db:      db,
	}
}

// ExecTx executes a function within a database transaction
func (store *SQLStore) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}
