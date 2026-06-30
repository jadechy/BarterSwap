package main

import (
	"database/sql"
	"fmt"
	"strings"
)

func createUserService(db *sql.DB, u *User) error {
	if strings.TrimSpace(u.Pseudo) == "" {
		return fmt.Errorf("le pseudo est requis: %w", ErrValidation)
	}
	return createUser(db, u)
}

func updateUserService(db *sql.DB, id int, u *User) error {
	if strings.TrimSpace(u.Pseudo) == "" {
		return fmt.Errorf("le pseudo est requis: %w", ErrValidation)
	}
	return updateUser(db, id, u)
}

func updateUserSkillsService(db *sql.DB, userID int, skills []Skill) error {
	for _, s := range skills {
		if !contains(NiveauxValides, s.Niveau) {
			return fmt.Errorf("niveau invalide %q: %w", s.Niveau, ErrValidation)
		}
		if strings.TrimSpace(s.Nom) == "" {
			return fmt.Errorf("le nom de la compétence est requis: %w", ErrValidation)
		}
	}
	return setUserSkills(db, userID, skills)
}

func createServiceService(db *sql.DB, s *Service) error {
	if strings.TrimSpace(s.Titre) == "" {
		return fmt.Errorf("le titre est requis: %w", ErrValidation)
	}
	if !contains(CategoriesValides, s.Categorie) {
		return fmt.Errorf("catégorie invalide %q: %w", s.Categorie, ErrValidation)
	}
	if s.Credits <= 0 {
		return fmt.Errorf("le coût en crédits doit être positif: %w", ErrValidation)
	}

	// Vérifie que le provider possède bien cette compétence (cas de test n°3 du sujet)
	skills, err := getSkillsByUserID(db, s.ProviderID)
	if err != nil {
		return err
	}
	hasSkill := false
	for _, sk := range skills {
		if strings.EqualFold(sk.Nom, s.Categorie) {
			hasSkill = true
			break
		}
	}
	if !hasSkill {
		return fmt.Errorf("vous n'avez pas la compétence %q: %w", s.Categorie, ErrValidation)
	}

	return createService(db, s)
}

func updateServiceService(db *sql.DB, id int, s *Service) error {
	if !contains(CategoriesValides, s.Categorie) {
		return fmt.Errorf("catégorie invalide %q: %w", s.Categorie, ErrValidation)
	}
	return updateService(db, id, s)
}

func createExchangeService(db *sql.DB, requesterID, serviceID int) (Exchange, error) {
	var e Exchange

	service, err := getServiceByID(db, serviceID)
	if err != nil {
		return e, err
	}

	if service.ProviderID == requesterID {
		return e, fmt.Errorf("impossible de s'échanger son propre service: %w", ErrSelfExchange)
	}

	active, err := hasActiveExchange(db, serviceID)
	if err != nil {
		return e, err
	}
	if active {
		return e, fmt.Errorf("ce service a déjà un échange en cours: %w", ErrExchangeConflict)
	}

	requester, err := getUserById(db, requesterID)
	if err != nil {
		return e, err
	}
	if requester.CreditBalance < service.Credits {
		return e, fmt.Errorf("solde insuffisant: %w", ErrInsufficientCredits)
	}

	e = Exchange{
		ServiceID:   serviceID,
		RequesterID: requesterID,
		OwnerID:     service.ProviderID,
	}
	err = createExchange(db, &e)
	return e, err
}

func acceptExchangeService(db *sql.DB, exchangeID, userID int) error {
	e, err := getExchangeByID(db, exchangeID)
	if err != nil {
		return err
	}

	if e.OwnerID != userID {
		return fmt.Errorf("seul le propriétaire du service peut accepter: %w", ErrUnauthorized)
	}
	if e.Status != "pending" {
		return fmt.Errorf("l'échange n'est pas en attente: %w", ErrInvalidStatus)
	}

	service, err := getServiceByID(db, e.ServiceID)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("acceptExchangeService begin: %w", err)
	}
	defer tx.Rollback()

	if err := addCreditTransaction(db, tx, e.RequesterID, &exchangeID, -service.Credits, "spend"); err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE exchanges SET status = 'accepted' WHERE id = ?`, exchangeID); err != nil {
		return fmt.Errorf("acceptExchangeService update: %w", err)
	}

	return tx.Commit()
}

func rejectExchangeService(db *sql.DB, exchangeID, userID int) error {
	e, err := getExchangeByID(db, exchangeID)
	if err != nil {
		return err
	}
	if e.OwnerID != userID {
		return fmt.Errorf("seul le propriétaire du service peut refuser: %w", ErrUnauthorized)
	}
	if e.Status != "pending" {
		return fmt.Errorf("l'échange n'est pas en attente: %w", ErrInvalidStatus)
	}
	return updateExchangeStatus(db, exchangeID, "rejected")
}

func completeExchangeService(db *sql.DB, exchangeID, userID int) error {
	e, err := getExchangeByID(db, exchangeID)
	if err != nil {
		return err
	}
	if e.OwnerID != userID && e.RequesterID != userID {
		return fmt.Errorf("vous ne faites pas partie de cet échange: %w", ErrUnauthorized)
	}
	if e.Status != "accepted" {
		return fmt.Errorf("l'échange doit être accepté pour être terminé: %w", ErrInvalidStatus)
	}

	service, err := getServiceByID(db, e.ServiceID)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("completeExchangeService begin: %w", err)
	}
	defer tx.Rollback()

	if err := addCreditTransaction(db, tx, e.OwnerID, &exchangeID, service.Credits, "earn"); err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE exchanges SET status = 'completed' WHERE id = ?`, exchangeID); err != nil {
		return fmt.Errorf("completeExchangeService update: %w", err)
	}

	return tx.Commit()
}

func cancelExchangeService(db *sql.DB, exchangeID, userID int) error {
	e, err := getExchangeByID(db, exchangeID)
	if err != nil {
		return err
	}
	if e.OwnerID != userID && e.RequesterID != userID {
		return fmt.Errorf("vous ne faites pas partie de cet échange: %w", ErrUnauthorized)
	}
	if e.Status != "pending" && e.Status != "accepted" {
		return fmt.Errorf("cet échange ne peut plus être annulé: %w", ErrInvalidStatus)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("cancelExchangeService begin: %w", err)
	}
	defer tx.Rollback()

	if e.Status == "accepted" {
		service, err := getServiceByID(db, e.ServiceID)
		if err != nil {
			return err
		}
		if err := addCreditTransaction(db, tx, e.RequesterID, &exchangeID, service.Credits, "refund"); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE exchanges SET status = 'cancelled' WHERE id = ?`, exchangeID); err != nil {
		return fmt.Errorf("cancelExchangeService update: %w", err)
	}

	return tx.Commit()
}

func createReviewService(db *sql.DB, r *Review) error {
	if r.Note < 1 || r.Note > 5 {
		return fmt.Errorf("la note doit être entre 1 et 5: %w", ErrValidation)
	}

	e, err := getExchangeByID(db, r.ExchangeID)
	if err != nil {
		return err
	}

	if e.Status != "completed" {
		return fmt.Errorf("l'échange doit être terminé pour être noté: %w", ErrExchangeNotDone)
	}

	if e.OwnerID != r.AuthorID && e.RequesterID != r.AuthorID {
		return fmt.Errorf("vous ne faites pas partie de cet échange: %w", ErrUnauthorized)
	}

	already, err := hasReviewed(db, r.ExchangeID, r.AuthorID)
	if err != nil {
		return err
	}
	if already {
		return fmt.Errorf("vous avez déjà noté cet échange: %w", ErrAlreadyReviewed)
	}

	if r.AuthorID == e.OwnerID {
		r.TargetID = e.RequesterID
	} else {
		r.TargetID = e.OwnerID
	}

	return createReview(db, r)
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, value) {
			return true
		}
	}
	return false
}