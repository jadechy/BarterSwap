package apperrors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("ressource introuvable")
	ErrInsufficientCredits = errors.New("crédits insuffisants")
	ErrExchangeConflict    = errors.New("un échange est déjà en cours pour ce service")
	ErrSelfExchange        = errors.New("impossible de s'échanger un service à soi-même")
	ErrExchangeNotDone     = errors.New("l'échange n'est pas terminé")
	ErrAlreadyReviewed     = errors.New("vous avez déjà laissé un avis pour cet échange")
	ErrInvalidStatus       = errors.New("transition de statut invalide")
	ErrUnauthorized        = errors.New("action non autorisée")
	ErrValidation          = errors.New("données invalides")
)

type ValidationError struct {
	Champ   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation échouée sur %s : %s", e.Champ, e.Message)
}

func (e ValidationError) Is(target error) bool {
	return target == ErrValidation
}
