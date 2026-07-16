package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/dbx"
)

// Repository expose tout ce dont les autres domaines ont besoin de user
// (ex: offer.Service utilise GetSkills pour valider une création d'offre).
type Repository interface {
	GetByID(ctx context.Context, id int) (User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, id int, u *User) error
	GetSkills(ctx context.Context, userID int) ([]Skill, error)
	SetSkills(ctx context.Context, userID int, skills []Skill) error
	Stats(ctx context.Context, userID int) (Stats, error)
	AddCreditTransaction(ctx context.Context, q dbx.Querier, userID int, exchangeID *int, montant int, typ string) error
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) GetByID(ctx context.Context, id int) (User, error) {
	var u User
	row := r.db.QueryRowContext(ctx, `
		SELECT id, pseudo, bio, ville, credit_balance, created_at
		FROM users WHERE id = ?`, id)

	err := row.Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &u.CreditBalance, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return u, apperrors.ErrNotFound
	}
	if err != nil {
		return u, fmt.Errorf("user.GetByID: %w", err)
	}

	skills, err := r.GetSkills(ctx, id)
	if err != nil {
		return u, err
	}
	u.Skills = skills

	return u, nil
}

func (r *sqlRepository) Create(ctx context.Context, u *User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("user.Create begin: %w", err)
	}
	defer func() {
		if rerr := tx.Rollback(); rerr != nil && !errors.Is(rerr, sql.ErrTxDone) {
			log.Printf("user.Create: erreur rollback: %v", rerr)
		}
	}()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (pseudo, bio, ville, credit_balance)
		VALUES (?, ?, ?, 10)`, u.Pseudo, u.Bio, u.Ville)
	if err != nil {
		return fmt.Errorf("user.Create insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("user.Create lastInsertId: %w", err)
	}
	u.ID = int(id)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
		VALUES (?, NULL, 10, 'welcome')`, u.ID)
	if err != nil {
		return fmt.Errorf("user.Create welcome credits: %w", err)
	}

	return tx.Commit()
}

func (r *sqlRepository) Update(ctx context.Context, id int, u *User) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users SET pseudo = ?, bio = ?, ville = ?
		WHERE id = ?`, u.Pseudo, u.Bio, u.Ville, id)
	if err != nil {
		return fmt.Errorf("user.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *sqlRepository) GetSkills(ctx context.Context, userID int) ([]Skill, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT nom, niveau FROM skills WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("user.GetSkills: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("user.GetSkills: erreur fermeture rows: %v", cerr)
		}
	}()

	var skills []Skill
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.Nom, &s.Niveau); err != nil {
			return nil, fmt.Errorf("user.GetSkills scan: %w", err)
		}
		skills = append(skills, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user.GetSkills rows: %w", err)
	}
	return skills, nil
}
func (r *sqlRepository) SetSkills(ctx context.Context, userID int, skills []Skill) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("user.SetSkills begin: %w", err)
	}
	defer func() {
		if rerr := tx.Rollback(); rerr != nil && !errors.Is(rerr, sql.ErrTxDone) {
			log.Printf("user.SetSkills: erreur rollback: %v", rerr)
		}
	}()

	_, err = tx.ExecContext(ctx, `DELETE FROM skills WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("user.SetSkills delete: %w", err)
	}

	for _, s := range skills {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO skills (user_id, nom, niveau)
			VALUES (?, ?, ?)`, userID, s.Nom, s.Niveau)
		if err != nil {
			return fmt.Errorf("user.SetSkills insert: %w", err)
		}
	}

	return tx.Commit()
}

func (r *sqlRepository) Stats(ctx context.Context, userID int) (Stats, error) {
	var s Stats
	s.UserID = userID

	err := r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM services WHERE provider_id = ? AND actif = true),
			(SELECT COUNT(*) FROM exchanges WHERE (requester_id = ? OR owner_id = ?) AND status = 'completed'),
			(SELECT credit_balance FROM users WHERE id = ?),
			(SELECT COALESCE(AVG(note), 0) FROM reviews WHERE target_id = ?),
			(SELECT COUNT(*) FROM reviews WHERE target_id = ?),
			(SELECT COALESCE(SUM(montant), 0) FROM credit_transactions WHERE user_id = ? AND montant > 0),
			(SELECT COALESCE(SUM(ABS(montant)), 0) FROM credit_transactions WHERE user_id = ? AND montant < 0)
		`,
		userID, userID, userID, userID, userID, userID, userID, userID,
	).Scan(
		&s.ServicesActifs,
		&s.EchangesCompletes,
		&s.CreditBalance,
		&s.NoteMoyenne,
		&s.NbAvis,
		&s.TotalGagne,
		&s.TotalDepense,
	)
	if err != nil {
		return s, fmt.Errorf("user.Stats: %w", err)
	}
	return s, nil
}

func (r *sqlRepository) AddCreditTransaction(ctx context.Context, q dbx.Querier, userID int, exchangeID *int, montant int, typ string) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
		VALUES (?, ?, ?, ?)`, userID, exchangeID, montant, typ)
	if err != nil {
		return fmt.Errorf("user.AddCreditTransaction insert: %w", err)
	}

	_, err = q.ExecContext(ctx, `
		UPDATE users SET credit_balance = credit_balance + ? WHERE id = ?`, montant, userID)
	if err != nil {
		return fmt.Errorf("user.AddCreditTransaction update balance: %w", err)
	}
	return nil
}
