package offer

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jadechy/barterswap/internal/apperrors"
)

type Repository interface {
	GetByID(ctx context.Context, id int) (Offer, error)
	List(ctx context.Context, f ListFilter) ([]Offer, error)
	Create(ctx context.Context, o *Offer) error
	Update(ctx context.Context, id int, o *Offer) error
	Delete(ctx context.Context, id int) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) GetByID(ctx context.Context, id int) (Offer, error) {
	var o Offer
	row := r.db.QueryRowContext(ctx, `
		SELECT id, provider_id, titre, description, categorie,
		       duree_minutes, credits, ville, actif, created_at
		FROM services WHERE id = ?`, id)

	err := row.Scan(&o.ID, &o.ProviderID, &o.Titre, &o.Description,
		&o.Categorie, &o.DureeMinutes, &o.Credits, &o.Ville, &o.Actif, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return o, fmt.Errorf("offre %d introuvable: %w", id, apperrors.ErrNotFound)
	}
	if err != nil {
		return o, fmt.Errorf("offer.GetByID: %w", err)
	}
	return o, nil
}

func (r *sqlRepository) List(ctx context.Context, f ListFilter) ([]Offer, error) {
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

	var offers []Offer
	for rows.Next() {
		var o Offer
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

func (r *sqlRepository) Create(ctx context.Context, o *Offer) error {
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

func (r *sqlRepository) Update(ctx context.Context, id int, o *Offer) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE services SET titre = ?, description = ?, categorie = ?,
		duree_minutes = ?, credits = ?, ville = ?, actif = ?
		WHERE id = ?`,
		o.Titre, o.Description, o.Categorie, o.DureeMinutes, o.Credits, o.Ville, o.Actif, id)
	if err != nil {
		return fmt.Errorf("offer.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *sqlRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM services WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("offer.Delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
