package exchange

import (
	"context"
	"fmt"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/dbx"
	"github.com/jadechy/barterswap/internal/offer"
	"github.com/jadechy/barterswap/internal/user"
)

type OfferGetter interface {
	GetByID(ctx context.Context, id int) (offer.Offer, error)
}

type UserGetter interface {
	GetByID(ctx context.Context, id int) (user.User, error)
}

type CreditLedger interface {
	AddCreditTransaction(ctx context.Context, q dbx.Querier, userID int, exchangeID *int, montant int, typ string) error
}

type Service struct {
	repo      Repository
	txManager dbx.TxRunner
	offers    OfferGetter
	users     UserGetter
	credits   CreditLedger
}

func NewService(repo Repository, txManager dbx.TxRunner, offers OfferGetter, users UserGetter, credits CreditLedger) *Service {
	return &Service{repo: repo, txManager: txManager, offers: offers, users: users, credits: credits}
}

func (s *Service) Create(ctx context.Context, requesterID, offerID int) (Exchange, error) {
	var e Exchange

	o, err := s.offers.GetByID(ctx, offerID)
	if err != nil {
		return e, err
	}

	if o.ProviderID == requesterID {
		return e, fmt.Errorf("impossible de s'échanger son propre service: %w", apperrors.ErrSelfExchange)
	}

	active, err := s.repo.HasActive(ctx, offerID)
	if err != nil {
		return e, err
	}
	if active {
		return e, fmt.Errorf("ce service a déjà un échange en cours: %w", apperrors.ErrExchangeConflict)
	}

	requester, err := s.users.GetByID(ctx, requesterID)
	if err != nil {
		return e, err
	}
	if requester.CreditBalance < o.Credits {
		return e, fmt.Errorf("solde insuffisant: %w", apperrors.ErrInsufficientCredits)
	}

	e = Exchange{
		ServiceID:   offerID,
		RequesterID: requesterID,
		OwnerID:     o.ProviderID,
	}
	err = s.repo.Create(ctx, &e)
	return e, err
}

func (s *Service) GetByID(ctx context.Context, id int) (Exchange, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, userID int, status string) ([]Exchange, error) {
	return s.repo.List(ctx, userID, status)
}

func (s *Service) Accept(ctx context.Context, exchangeID, userID int) error {
	e, err := s.repo.GetByID(ctx, exchangeID)
	if err != nil {
		return err
	}
	if e.OwnerID != userID {
		return fmt.Errorf("seul le propriétaire du service peut accepter: %w", apperrors.ErrUnauthorized)
	}
	if e.Status != "pending" {
		return fmt.Errorf("l'échange n'est pas en attente: %w", apperrors.ErrInvalidStatus)
	}

	o, err := s.offers.GetByID(ctx, e.ServiceID)
	if err != nil {
		return err
	}

	return s.txManager.WithTx(ctx, func(q dbx.Querier) error {
		if err := s.credits.AddCreditTransaction(ctx, q, e.RequesterID, &exchangeID, -o.Credits, "spend"); err != nil {
			return err
		}
		return s.repo.UpdateStatus(ctx, q, exchangeID, "accepted")
	})
}

func (s *Service) Reject(ctx context.Context, exchangeID, userID int) error {
	e, err := s.repo.GetByID(ctx, exchangeID)
	if err != nil {
		return err
	}
	if e.OwnerID != userID {
		return fmt.Errorf("seul le propriétaire du service peut refuser: %w", apperrors.ErrUnauthorized)
	}
	if e.Status != "pending" {
		return fmt.Errorf("l'échange n'est pas en attente: %w", apperrors.ErrInvalidStatus)
	}

	return s.txManager.WithTx(ctx, func(q dbx.Querier) error {
		return s.repo.UpdateStatus(ctx, q, exchangeID, "rejected")
	})
}

func (s *Service) Complete(ctx context.Context, exchangeID, userID int) error {
	e, err := s.repo.GetByID(ctx, exchangeID)
	if err != nil {
		return err
	}
	if e.OwnerID != userID && e.RequesterID != userID {
		return fmt.Errorf("vous ne faites pas partie de cet échange: %w", apperrors.ErrUnauthorized)
	}
	if e.Status != "accepted" {
		return fmt.Errorf("l'échange doit être accepté pour être terminé: %w", apperrors.ErrInvalidStatus)
	}

	o, err := s.offers.GetByID(ctx, e.ServiceID)
	if err != nil {
		return err
	}

	return s.txManager.WithTx(ctx, func(q dbx.Querier) error {
		if err := s.credits.AddCreditTransaction(ctx, q, e.OwnerID, &exchangeID, o.Credits, "earn"); err != nil {
			return err
		}
		return s.repo.UpdateStatus(ctx, q, exchangeID, "completed")
	})
}

func (s *Service) Cancel(ctx context.Context, exchangeID, userID int) error {
	e, err := s.repo.GetByID(ctx, exchangeID)
	if err != nil {
		return err
	}
	if e.OwnerID != userID && e.RequesterID != userID {
		return fmt.Errorf("vous ne faites pas partie de cet échange: %w", apperrors.ErrUnauthorized)
	}
	if e.Status != "pending" && e.Status != "accepted" {
		return fmt.Errorf("cet échange ne peut plus être annulé: %w", apperrors.ErrInvalidStatus)
	}

	return s.txManager.WithTx(ctx, func(q dbx.Querier) error {
		if e.Status == "accepted" {
			o, err := s.offers.GetByID(ctx, e.ServiceID)
			if err != nil {
				return err
			}
			if err := s.credits.AddCreditTransaction(ctx, q, e.RequesterID, &exchangeID, o.Credits, "refund"); err != nil {
				return err
			}
		}
		return s.repo.UpdateStatus(ctx, q, exchangeID, "cancelled")
	})
}
