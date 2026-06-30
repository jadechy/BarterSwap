package main

import (
	"database/sql"
	"fmt"
)

func getUserById(db *sql.DB, id int) (User, error) {
	var u User
	row := db.QueryRow(`SELECT id, pseudo, bio, ville, credit_balance, created_at FROM users WHERE id = ?`, id)

	err := row.Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &u.CreditBalance, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return u, ErrNotFound
	}
	if err != nil {
		return u, fmt.Errorf("getUserByID: %w", err)
	}

	skills, err := getSkillsByUserID(db, id)
	if err != nil {
		return u, err
	}
	u.Skills = skills

	return u, nil
}

func createUser(db *sql.DB, u *User) error {
	tx, err := db.Begin()
	if err != nil{
		return fmt.Errorf("createUser begin: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO users (pseudo, bio, ville, credit_balance)
		VALUES (?, ?, ?, 10)`, u.Pseudo, u.Bio, u.Ville)

	if err != nil {
		return fmt.Errorf("createUser insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("createUser lastInsertId: %w", err)
	}

	u.ID = int(id)

	_, err = tx.Exec(`
		INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
		VALUES (?, NULL, 10, 'welcome')`, u.ID)
	if err != nil {
		return fmt.Errorf("createUser welcome credits: %w", err)
	}

	return tx.Commit()
}

func updateUser(db *sql.DB, id int, u *User) error {
	result, err:= db.Exec(`
		UPDATE users SET pseudo = ?, bio = ?, ville = ?
		WHERE id = ?`, u.Pseudo, u.Bio, u.Ville, id)
	
	if err != nil {
		return fmt.Errorf("updateUser: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func getSkillsByUserID(db *sql.DB, userID int) ([]Skill, error) {
	rows, err := db.Query(`
		SELECT nom, niveau FROM skills WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("getSkillsByUserID: %w", err)
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.Nom, &s.Niveau); err != nil {
			return nil, fmt.Errorf("getSkillsByUserID scan: %w", err)
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func setUserSkills(db *sql.DB, userID int, skills []Skill) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("setUserSkills begin: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM skills WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("setUserSkills delete: %w", err)
	}

	for _, s := range skills {
		_, err = tx.Exec(`
			INSERT INTO skills (user_id, nom, niveau)
			VALUES (?, ?, ?)`, userID, s.Nom, s.Niveau)
		if err != nil {
			return fmt.Errorf("setUserSkills insert: %w", err)
		}
	}

	return tx.Commit()
}

func getServiceByID(db *sql.DB, id int) (Service, error) {
	var s Service
	row := db.QueryRow(`
		SELECT id, provider_id, titre, description, categorie,
		       duree_minutes, credits, ville, actif, created_at
		FROM services WHERE id = ?`, id)

	err := row.Scan(&s.ID, &s.ProviderID, &s.Titre, &s.Description,
		&s.Categorie, &s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return s, ErrNotFound
	}
	if err != nil {
		return s, fmt.Errorf("getServiceByID: %w", err)
	}
	return s, nil
}

func listServices(db *sql.DB, categorie, ville, search string) ([]Service, error) {
	query := `SELECT id, provider_id, titre, description, categorie,
	                 duree_minutes, credits, ville, actif, created_at
	          FROM services WHERE actif = true`
	args := []any{}

	if categorie != "" {
		query += " AND categorie = ?"
		args = append(args, categorie)
	}
	if ville != "" {
		query += " AND ville = ?"
		args = append(args, ville)
	}
	if search != "" {
		query += " AND (titre LIKE ? OR description LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listServices: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.ID, &s.ProviderID, &s.Titre, &s.Description,
			&s.Categorie, &s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("listServices scan: %w", err)
		}
		services = append(services, s)
	}
	return services, nil
}

func createService(db *sql.DB, s *Service) error {
	result, err := db.Exec(`
		INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif)
		VALUES (?, ?, ?, ?, ?, ?, ?, true)`,
		s.ProviderID, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville)
	if err != nil {
		return fmt.Errorf("createService: %w", err)
	}
	id, _ := result.LastInsertId()
	s.ID = int(id)
	return nil
}

func updateService(db *sql.DB, id int, s *Service) error {
	result, err := db.Exec(`
		UPDATE services SET titre = ?, description = ?, categorie = ?,
		duree_minutes = ?, credits = ?, ville = ?, actif = ?
		WHERE id = ?`,
		s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif, id)
	if err != nil {
		return fmt.Errorf("updateService: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func deleteService(db *sql.DB, id int) error {
	result, err := db.Exec(`DELETE FROM services WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleteService: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func getExchangeByID(db *sql.DB, id int) (Exchange, error) {
	var e Exchange
	row := db.QueryRow(`
		SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
		FROM exchanges WHERE id = ?`, id)

	err := row.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID,
		&e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	if err != nil {
		return e, fmt.Errorf("getExchangeByID: %w", err)
	}
	return e, nil
}

func listExchanges(db *sql.DB, userID int, status string) ([]Exchange, error) {
	query := `SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
	          FROM exchanges WHERE (requester_id = ? OR owner_id = ?)`
	args := []any{userID, userID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listExchanges: %w", err)
	}
	defer rows.Close()

	var exchanges []Exchange
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID,
			&e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("listExchanges scan: %w", err)
		}
		exchanges = append(exchanges, e)
	}
	return exchanges, nil
}

func createExchange(db *sql.DB, e *Exchange) error {
	result, err := db.Exec(`
		INSERT INTO exchanges (service_id, requester_id, owner_id, status)
		VALUES (?, ?, ?, 'pending')`,
		e.ServiceID, e.RequesterID, e.OwnerID)
	if err != nil {
		return fmt.Errorf("createExchange: %w", err)
	}
	id, _ := result.LastInsertId()
	e.ID = int(id)
	return nil
}

func updateExchangeStatus(db *sql.DB, id int, status string) error {
	result, err := db.Exec(`
		UPDATE exchanges SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("updateExchangeStatus: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func hasActiveExchange(db *sql.DB, serviceID int) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM exchanges
		WHERE service_id = ? AND status IN ('pending', 'accepted')`, serviceID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("hasActiveExchange: %w", err)
	}
	return count > 0, nil
}

func addCreditTransaction(db *sql.DB, tx *sql.Tx, userID int, exchangeID *int, montant int, typeTransaction string) error {
	_, err := tx.Exec(`
		INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
		VALUES (?, ?, ?, ?)`, userID, exchangeID, montant, typeTransaction)
	if err != nil {
		return fmt.Errorf("addCreditTransaction insert: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE users SET credit_balance = credit_balance + ? WHERE id = ?`, montant, userID)
	if err != nil {
		return fmt.Errorf("addCreditTransaction update balance: %w", err)
	}
	return nil
}

func createReview(db *sql.DB, r *Review) error {
	result, err := db.Exec(`
		INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire)
		VALUES (?, ?, ?, ?, ?)`,
		r.ExchangeID, r.AuthorID, r.TargetID, r.Note, r.Commentaire)
	if err != nil {
		return fmt.Errorf("createReview: %w", err)
	}
	id, _ := result.LastInsertId()
	r.ID = int(id)
	return nil
}

func getReviewsByUserID(db *sql.DB, userID int) ([]Review, error) {
	rows, err := db.Query(`
		SELECT id, exchange_id, author_id, target_id, note, commentaire, created_at
		FROM reviews WHERE target_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("getReviewsByUserID: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID,
			&r.Note, &r.Commentaire, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("getReviewsByUserID scan: %w", err)
		}
		reviews = append(reviews, r)
	}
	return reviews, nil
}

func getReviewsByServiceID(db *sql.DB, serviceID int) ([]Review, error) {
	rows, err := db.Query(`
		SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, r.commentaire, r.created_at
		FROM reviews r
		JOIN exchanges e ON r.exchange_id = e.id
		WHERE e.service_id = ?`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("getReviewsByServiceID: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID,
			&r.Note, &r.Commentaire, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("getReviewsByServiceID scan: %w", err)
		}
		reviews = append(reviews, r)
	}
	return reviews, nil
}

func hasReviewed(db *sql.DB, exchangeID, authorID int) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM reviews
		WHERE exchange_id = ? AND author_id = ?`, exchangeID, authorID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("hasReviewed: %w", err)
	}
	return count > 0, nil
}

func getUserStats(db *sql.DB, userID int) (UserStats, error) {
	var stats UserStats
	stats.UserID = userID

	err := db.QueryRow(`
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
		&stats.ServicesActifs,
		&stats.EchangesCompletes,
		&stats.CreditBalance,
		&stats.NoteMoyenne,
		&stats.NbAvis,
		&stats.TotalGagne,
		&stats.TotalDepense,
	)
	if err != nil {
		return stats, fmt.Errorf("getUserStats: %w", err)
	}
	return stats, nil
}