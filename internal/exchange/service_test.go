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
	"github.com/jadechy/barterswap/internal/service"
	servicemocks "github.com/jadechy/barterswap/internal/service/mocks"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

type exchangeMocks struct {
	repo     *exchangemocks.MockRepository
	tx       *dbxmocks.MockTxRunner
	services *servicemocks.MockRepository
	users    *usermocks.MockRepository
}

func newExchangeService(t *testing.T) (*exchange.Manager, exchangeMocks) {
	m := exchangeMocks{
		repo:     exchangemocks.NewMockRepository(t),
		tx:       dbxmocks.NewMockTxRunner(t),
		services: servicemocks.NewMockRepository(t),
		users:    usermocks.NewMockRepository(t),
	}
	svc := exchange.NewService(m.repo, m.tx, m.services, m.users, m.users)
	return svc, m
}

func TestCreate(t *testing.T) {
	cases := []struct {
		name        string
		requesterID int
		serviceID   int
		setupMocks  func(m exchangeMocks)
		wantErr     error
	}{
		{
			name:        "soi-meme",
			requesterID: 1,
			serviceID:   10,
			setupMocks: func(m exchangeMocks) {
				m.services.EXPECT().GetByID(mock.Anything, 10).Return(service.Service{ID: 10, ProviderID: 1, Credits: 5}, nil)
			},
			wantErr: apperrors.ErrSelfExchange,
		},
		{
			name:        "echange actif existant",
			requesterID: 1,
			serviceID:   10,
			setupMocks: func(m exchangeMocks) {
				m.services.EXPECT().GetByID(mock.Anything, 10).Return(service.Service{ID: 10, ProviderID: 2, Credits: 5}, nil)
				m.repo.EXPECT().HasActive(mock.Anything, 10).Return(true, nil)
			},
			wantErr: apperrors.ErrExchangeConflict,
		},
		{
			name:        "solde insuffisant",
			requesterID: 1,
			serviceID:   10,
			setupMocks: func(m exchangeMocks) {
				m.services.EXPECT().GetByID(mock.Anything, 10).Return(service.Service{ID: 10, ProviderID: 2, Credits: 20}, nil)
				m.repo.EXPECT().HasActive(mock.Anything, 10).Return(false, nil)
				m.users.EXPECT().GetByID(mock.Anything, 1).Return(user.User{ID: 1, CreditBalance: 5}, nil)
			},
			wantErr: apperrors.ErrInsufficientCredits,
		},
		{
			name:        "valide: succes",
			requesterID: 1,
			serviceID:   10,
			setupMocks: func(m exchangeMocks) {
				m.services.EXPECT().GetByID(mock.Anything, 10).Return(service.Service{ID: 10, ProviderID: 2, Credits: 5}, nil)
				m.repo.EXPECT().HasActive(mock.Anything, 10).Return(false, nil)
				m.users.EXPECT().GetByID(mock.Anything, 1).Return(user.User{ID: 1, CreditBalance: 10}, nil)
				m.repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*exchange.Exchange")).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newExchangeService(t)
			tc.setupMocks(m)

			_, err := svc.Create(context.Background(), tc.requesterID, tc.serviceID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccept(t *testing.T) {
	cases := []struct {
		name       string
		exchangeID int
		userID     int
		setupMocks func(m exchangeMocks)
		wantErr    error
	}{
		{
			name:       "non proprietaire",
			exchangeID: 1,
			userID:     99,
			setupMocks: func(m exchangeMocks) {
				m.repo.EXPECT().GetByID(mock.Anything, 1).Return(exchange.Exchange{ID: 1, OwnerID: 2, Status: "pending"}, nil)
			},
			wantErr: apperrors.ErrUnauthorized,
		},
		{
			name:       "statut invalide",
			exchangeID: 1,
			userID:     2,
			setupMocks: func(m exchangeMocks) {
				m.repo.EXPECT().GetByID(mock.Anything, 1).Return(exchange.Exchange{ID: 1, OwnerID: 2, Status: "completed"}, nil)
			},
			wantErr: apperrors.ErrInvalidStatus,
		},
		{
			name:       "valide: succes",
			exchangeID: 1,
			userID:     2,
			setupMocks: func(m exchangeMocks) {
				m.repo.EXPECT().GetByID(mock.Anything, 1).
					Return(exchange.Exchange{ID: 1, ServiceID: 10, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)
				m.services.EXPECT().GetByID(mock.Anything, 10).Return(service.Service{ID: 10, Credits: 5}, nil)
				m.tx.EXPECT().
					WithTx(mock.Anything, mock.AnythingOfType("func(dbx.Querier) error")).
					RunAndReturn(func(ctx context.Context, fn func(dbx.Querier) error) error {
						return fn(nil)
					})
				m.users.EXPECT().
					AddCreditTransaction(mock.Anything, mock.Anything, 1, mock.Anything, -5, "spend").
					Return(nil)
				m.repo.EXPECT().UpdateStatus(mock.Anything, mock.Anything, 1, "accepted").Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newExchangeService(t)
			tc.setupMocks(m)

			err := svc.Accept(context.Background(), tc.exchangeID, tc.userID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
