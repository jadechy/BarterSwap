package review

import (
	"context"
	"fmt"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/exchange"
)

// ExchangeGetter est le sous-ensemble de exchange.Repository dont review a besoin.
// Définie ici (côté consommateur) plutôt que dans exchange, pour éviter un couplage
// plus large que nécessaire — review n'a besoin que de lire un échange par ID.
type ExchangeGetter interface {
	GetByID(ctx context.Context, id int) (exchange.Exchange, error)
}

type Service struct {
	repo         Repository
	exchangeRepo ExchangeGetter
}

func NewService(repo Repository, exchangeRepo ExchangeGetter) *Service {
	return &Service{repo: repo, exchangeRepo: exchangeRepo}
}

func (s *Service) Create(ctx context.Context, r *Review) error {
	if r.Note < 1 || r.Note > 5 {
		return fmt.Errorf("la note doit être entre 1 et 5: %w", apperrors.ErrValidation)
	}

	e, err := s.exchangeRepo.GetByID(ctx, r.ExchangeID)
	if err != nil {
		return err
	}

	if e.Status != "completed" {
		return fmt.Errorf("l'échange doit être terminé pour être noté: %w", apperrors.ErrExchangeNotDone)
	}

	if e.OwnerID != r.AuthorID && e.RequesterID != r.AuthorID {
		return fmt.Errorf("vous ne faites pas partie de cet échange: %w", apperrors.ErrUnauthorized)
	}

	already, err := s.repo.HasReviewed(ctx, r.ExchangeID, r.AuthorID)
	if err != nil {
		return err
	}
	if already {
		return fmt.Errorf("vous avez déjà noté cet échange: %w", apperrors.ErrAlreadyReviewed)
	}

	if r.AuthorID == e.OwnerID {
		r.TargetID = e.RequesterID
	} else {
		r.TargetID = e.OwnerID
	}

	return s.repo.Create(ctx, r)
}

func (s *Service) GetByUserID(ctx context.Context, userID int) ([]Review, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) GetByServiceID(ctx context.Context, serviceID int) ([]Review, error) {
	return s.repo.GetByServiceID(ctx, serviceID)
}
