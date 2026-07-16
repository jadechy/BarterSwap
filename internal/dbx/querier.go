package dbx

import (
	"context"
	"database/sql"
	"fmt"
)

// Querier est satisfaite à la fois par *sql.DB et *sql.Tx —
// ça permet aux repositories d'écrire du code SQL identique,
// qu'ils soient appelés en dehors ou à l'intérieur d'une transaction.
type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
type TxRunner interface {
	WithTx(ctx context.Context, fn func(q Querier) error) error
}
type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTx ouvre une transaction, exécute fn avec un Querier lié à cette transaction,
// commit si fn réussit, rollback sinon (ou en cas de panic).
func (tm *TxManager) WithTx(ctx context.Context, fn func(q Querier) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("dbx.WithTx begin: %w", err)
	}
	defer tx.Rollback() // no-op si déjà commit

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
