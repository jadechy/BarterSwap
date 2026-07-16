package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/exchange"
	exchangemocks "github.com/jadechy/barterswap/internal/exchange/mocks"
	reviewmocks "github.com/jadechy/barterswap/internal/review/mocks"
)

func TestCreate_NoteHorsBornes_RetourneErreurValidation(t *testing.T) {
	repo := reviewmocks.NewMockRepository(t)
	exchangeRepo := exchangemocks.NewMockRepository(t)
	svc := NewService(repo, exchangeRepo)

	r := &Review{ExchangeID: 1, AuthorID: 1, Note: 6}
	err := svc.Create(context.Background(), r)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestCreate_EchangeNonTermine_RetourneErreur(t *testing.T) {
	repo := reviewmocks.NewMockRepository(t)
	exchangeRepo := exchangemocks.NewMockRepository(t)
	svc := NewService(repo, exchangeRepo)

	r := &Review{ExchangeID: 1, AuthorID: 1, Note: 5}

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, Status: "pending", OwnerID: 2, RequesterID: 1}, nil)

	err := svc.Create(context.Background(), r)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrExchangeNotDone)
}

func TestCreate_AuteurHorsEchange_RetourneErreurUnauthorized(t *testing.T) {
	repo := reviewmocks.NewMockRepository(t)
	exchangeRepo := exchangemocks.NewMockRepository(t)
	svc := NewService(repo, exchangeRepo)

	r := &Review{ExchangeID: 1, AuthorID: 99, Note: 5}

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, Status: "completed", OwnerID: 2, RequesterID: 1}, nil)

	err := svc.Create(context.Background(), r)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestCreate_DejaNote_RetourneErreur(t *testing.T) {
	repo := reviewmocks.NewMockRepository(t)
	exchangeRepo := exchangemocks.NewMockRepository(t)
	svc := NewService(repo, exchangeRepo)

	r := &Review{ExchangeID: 1, AuthorID: 1, Note: 5}

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, Status: "completed", OwnerID: 2, RequesterID: 1}, nil)
	repo.EXPECT().
		HasReviewed(mock.Anything, 1, 1).
		Return(true, nil)

	err := svc.Create(context.Background(), r)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrAlreadyReviewed)
}

func TestCreate_Valide_Succes(t *testing.T) {
	repo := reviewmocks.NewMockRepository(t)
	exchangeRepo := exchangemocks.NewMockRepository(t)
	svc := NewService(repo, exchangeRepo)

	r := &Review{ExchangeID: 1, AuthorID: 1, Note: 5, Commentaire: "Top"}

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, Status: "completed", OwnerID: 2, RequesterID: 1}, nil)
	repo.EXPECT().
		HasReviewed(mock.Anything, 1, 1).
		Return(false, nil)
	repo.EXPECT().
		Create(mock.Anything, r).
		Return(nil)

	err := svc.Create(context.Background(), r)

	require.NoError(t, err)
	assert.Equal(t, 2, r.TargetID)
}
