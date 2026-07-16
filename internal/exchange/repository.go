package exchange

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/dbx"
)

type Repository interface {
	GetByID(ctx context.Context, id int) (Exchange, error)
	List(ctx context.Context, userID int, status string) ([]Exchange, error)
	Create(ctx context.Context, e *Exchange) error
	HasActive(ctx context.Context, serviceID int) (bool, error)
	// UpdateStatus prend un Querier explicite : appelé depuis Service dans une tx
	// partagée avec l'écriture des credit_transactions.
	UpdateStatus(ctx context.Context, q dbx.Querier, id int, status string) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) GetByID(ctx context.Context, id int) (Exchange, error) {
	var e Exchange
	row := r.db.QueryRowContext(ctx, `
		SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
		FROM exchanges WHERE id = ?`, id)

	err := row.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID,
		&e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return e, fmt.Errorf("exchange %d introuvable: %w", id, apperrors.ErrNotFound)

	}
	if err != nil {
		return e, fmt.Errorf("exchange.GetByID: %w", err)
	}
	return e, nil
}

func (r *sqlRepository) List(ctx context.Context, userID int, status string) ([]Exchange, error) {
	query := `SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
	          FROM exchanges WHERE (requester_id = ? OR owner_id = ?)`
	args := []any{userID, userID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exchange.List: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("exchange.List: erreur fermeture rows: %v", cerr)
		}
	}()

	var exchanges []Exchange
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID,
			&e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("exchange.List scan: %w", err)
		}
		exchanges = append(exchanges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("exchange.List rows: %w", err)
	}
	return exchanges, nil
}

func (r *sqlRepository) Create(ctx context.Context, e *Exchange) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO exchanges (service_id, requester_id, owner_id, status)
		VALUES (?, ?, ?, 'pending')`,
		e.ServiceID, e.RequesterID, e.OwnerID)
	if err != nil {
		return fmt.Errorf("exchange.Create: %w", err)
	}
	id, _ := result.LastInsertId()
	e.ID = int(id)
	return nil
}

func (r *sqlRepository) HasActive(ctx context.Context, serviceID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM exchanges
		WHERE service_id = ? AND status IN ('pending', 'accepted')`, serviceID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("exchange.HasActive: %w", err)
	}
	return count > 0, nil
}

func (r *sqlRepository) UpdateStatus(ctx context.Context, q dbx.Querier, id int, status string) error {
	result, err := q.ExecContext(ctx, `UPDATE exchanges SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("exchange.UpdateStatus: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
