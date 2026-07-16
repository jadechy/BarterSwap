package review

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type Repository interface {
	Create(ctx context.Context, r *Review) error
	GetByUserID(ctx context.Context, userID int) ([]Review, error)
	GetByServiceID(ctx context.Context, serviceID int) ([]Review, error)
	HasReviewed(ctx context.Context, exchangeID, authorID int) (bool, error)
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, rev *Review) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire)
		VALUES (?, ?, ?, ?, ?)`,
		rev.ExchangeID, rev.AuthorID, rev.TargetID, rev.Note, rev.Commentaire)
	if err != nil {
		return fmt.Errorf("review.Create: %w", err)
	}
	id, _ := result.LastInsertId()
	rev.ID = int(id)
	return nil
}
func (r *sqlRepository) GetByUserID(ctx context.Context, userID int) ([]Review, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, exchange_id, author_id, target_id, note, commentaire, created_at
		FROM reviews WHERE target_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("review.GetByUserID: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("review.GetByUserID: erreur fermeture rows: %v", cerr)
		}
	}()
	return scanReviews(rows)
}

func (r *sqlRepository) GetByServiceID(ctx context.Context, serviceID int) ([]Review, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, r.commentaire, r.created_at
		FROM reviews r
		JOIN exchanges e ON r.exchange_id = e.id
		WHERE e.service_id = ?`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("review.GetByServiceID: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("review.GetByServiceID: erreur fermeture rows: %v", cerr)
		}
	}()
	return scanReviews(rows)
}

func (r *sqlRepository) HasReviewed(ctx context.Context, exchangeID, authorID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM reviews
		WHERE exchange_id = ? AND author_id = ?`, exchangeID, authorID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("review.HasReviewed: %w", err)
	}
	return count > 0, nil
}

func scanReviews(rows *sql.Rows) ([]Review, error) {
	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID,
			&r.Note, &r.Commentaire, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("review scan: %w", err)
		}
		reviews = append(reviews, r)
	}
	return reviews, nil
}
