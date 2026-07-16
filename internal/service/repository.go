package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jadechy/barterswap/internal/apperrors"
)

type Repository interface {
	GetByID(ctx context.Context, id int) (Service, error)
	List(ctx context.Context, f ListFilter) ([]Service, error)
	Create(ctx context.Context, o *Service) error
	Update(ctx context.Context, id int, o *Service) error
	Delete(ctx context.Context, id int) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) GetByID(ctx context.Context, id int) (Service, error) {
	var o Service
	row := r.db.QueryRowContext(ctx, `
		SELECT id, provider_id, titre, description, categorie,
		       duree_minutes, credits, ville, actif, created_at
		FROM services WHERE id = ?`, id)

	err := row.Scan(&o.ID, &o.ProviderID, &o.Titre, &o.Description,
		&o.Categorie, &o.DureeMinutes, &o.Credits, &o.Ville, &o.Actif, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return o, fmt.Errorf("offre %d: %w", id, apperrors.ErrNotFound)
	}
	if err != nil {
		return o, fmt.Errorf("offer.GetByID: %w", err)
	}
	return o, nil
}

func (r *sqlRepository) List(ctx context.Context, f ListFilter) ([]Service, error) {
	query := `SELECT id, provider_id, titre, description, categorie,
	                 duree_minutes, credits, ville, actif, created_at
	          FROM services WHERE actif = true`
	args := []any{}

	if f.Categorie != "" {
		query += " AND categorie = ?"
		args = append(args, f.Categorie)
	}
	if f.Ville != "" {
		query += " AND ville = ?"
		args = append(args, f.Ville)
	}
	if f.Search != "" {
		query += " AND (titre LIKE ? OR description LIKE ?)"
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%")
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("offer.List: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("offer.List: erreur fermeture rows: %v", cerr)
		}
	}()

	var offers []Service
	for rows.Next() {
		var o Service
		if err := rows.Scan(&o.ID, &o.ProviderID, &o.Titre, &o.Description,
			&o.Categorie, &o.DureeMinutes, &o.Credits, &o.Ville, &o.Actif, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("offer.List scan: %w", err)
		}
		offers = append(offers, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("offer.List rows: %w", err)
	}
	return offers, nil
}

func (r *sqlRepository) Create(ctx context.Context, o *Service) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif)
		VALUES (?, ?, ?, ?, ?, ?, ?, true)`,
		o.ProviderID, o.Titre, o.Description, o.Categorie, o.DureeMinutes, o.Credits, o.Ville)
	if err != nil {
		return fmt.Errorf("offer.Create: %w", err)
	}
	id, _ := result.LastInsertId()
	o.ID = int(id)
	return nil
}

func (r *sqlRepository) Update(ctx context.Context, id int, o *Service) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE services SET titre = ?, description = ?, categorie = ?,
		duree_minutes = ?, credits = ?, ville = ?, actif = ?
		WHERE id = ?`,
		o.Titre, o.Description, o.Categorie, o.DureeMinutes, o.Credits, o.Ville, o.Actif, id)
	if err != nil {
		return fmt.Errorf("offer.Update: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("offer.Update rowsAffected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("service %d: %w", id, apperrors.ErrNotFound)
	}
	return nil
}

func (r *sqlRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM services WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("offer.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("offer.Delete rowsAffected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("service %d: %w", id, apperrors.ErrNotFound)
	}
	return nil
}
