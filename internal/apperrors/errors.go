package apperrors

import "errors"

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
