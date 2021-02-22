package sqlapi

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type TxContext interface {
	sqlx.ExtContext
	sqlx.PreparerContext
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type TransactionWork func(ctx context.Context, db TxContext) error

func RunInTransaction(ctx context.Context, db *sqlx.DB, work TransactionWork) error {
	return RunInTransactionWithOptions(ctx, db, work, nil)
}

func RunInTransactionWithOptions(ctx context.Context, db *sqlx.DB, work TransactionWork, opts *sql.TxOptions) error {
	tx, err := db.BeginTxx(ctx, opts)
	defer func() {
		tx = nil
	}()

	if err != nil {
		return fmt.Errorf("can't create transaction: %w", err)
	}

	if err := work(ctx, tx); err != nil {
		if rallbackErr := tx.Rollback(); rallbackErr != nil {
			return fmt.Errorf("can't rallback transaction: %w, after inner err: %v", rallbackErr, err)
		}

		return fmt.Errorf("can't do transaction work: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("can't commit transaction: %w", err)
	}
	return nil
}
