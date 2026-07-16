package exchange_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/dbx"
	dbxmocks "github.com/jadechy/barterswap/internal/dbx/mocks"
	"github.com/jadechy/barterswap/internal/exchange"
	exchangemocks "github.com/jadechy/barterswap/internal/exchange/mocks"
	"github.com/jadechy/barterswap/internal/offer"
	offermocks "github.com/jadechy/barterswap/internal/offer/mocks"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

func TestCreate_SoiMeme_RetourneErreurSelfExchange(t *testing.T) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)
	svc := exchange.NewService(repo, tx, offers, users, users)

	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, ProviderID: 1, Credits: 5}, nil)

	_, err := svc.Create(context.Background(), 1, 10)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrSelfExchange)
}

func TestCreate_EchangeActifExistant_RetourneErreurConflict(t *testing.T) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)
	svc := exchange.NewService(repo, tx, offers, users, users)

	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, ProviderID: 2, Credits: 5}, nil)
	repo.EXPECT().
		HasActive(mock.Anything, 10).
		Return(true, nil)

	_, err := svc.Create(context.Background(), 1, 10)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrExchangeConflict)
}

func TestCreate_SoldeInsuffisant_RetourneErreur(t *testing.T) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)
	svc := exchange.NewService(repo, tx, offers, users, users)

	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, ProviderID: 2, Credits: 20}, nil)
	repo.EXPECT().
		HasActive(mock.Anything, 10).
		Return(false, nil)
	users.EXPECT().
		GetByID(mock.Anything, 1).
		Return(user.User{ID: 1, CreditBalance: 5}, nil)

	_, err := svc.Create(context.Background(), 1, 10)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrInsufficientCredits)
}

func TestCreate_Valide_Succes(t *testing.T) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)
	svc := exchange.NewService(repo, tx, offers, users, users)

	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, ProviderID: 2, Credits: 5}, nil)
	repo.EXPECT().
		HasActive(mock.Anything, 10).
		Return(false, nil)
	users.EXPECT().
		GetByID(mock.Anything, 1).
		Return(user.User{ID: 1, CreditBalance: 10}, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*exchange.Exchange")).
		Return(nil)

	e, err := svc.Create(context.Background(), 1, 10)

	require.NoError(t, err)
	assert.Equal(t, 2, e.OwnerID)
	assert.Equal(t, 1, e.RequesterID)
}

func TestAccept_NonProprietaire_RetourneErreurUnauthorized(t *testing.T) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)
	svc := exchange.NewService(repo, tx, offers, users, users)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, Status: "pending"}, nil)

	err := svc.Accept(context.Background(), 1, 99)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestAccept_StatutInvalide_RetourneErreur(t *testing.T) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)
	svc := exchange.NewService(repo, tx, offers, users, users)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, Status: "completed"}, nil)

	err := svc.Accept(context.Background(), 1, 2)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrInvalidStatus)
}

func TestAccept_Valide_Succes(t *testing.T) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)
	svc := exchange.NewService(repo, tx, offers, users, users)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, ServiceID: 10, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)
	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, Credits: 5}, nil)

	tx.EXPECT().
		WithTx(mock.Anything, mock.AnythingOfType("func(dbx.Querier) error")).
		RunAndReturn(func(ctx context.Context, fn func(dbx.Querier) error) error {
			return fn(nil)
		})
	users.EXPECT().
		AddCreditTransaction(mock.Anything, mock.Anything, 1, mock.Anything, -5, "spend").
		Return(nil)
	repo.EXPECT().
		UpdateStatus(mock.Anything, mock.Anything, 1, "accepted").
		Return(nil)

	err := svc.Accept(context.Background(), 1, 2)

	require.NoError(t, err)
}
