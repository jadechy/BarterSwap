package review_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/exchange"
	exchangemocks "github.com/jadechy/barterswap/internal/exchange/mocks"
	"github.com/jadechy/barterswap/internal/review"
	reviewmocks "github.com/jadechy/barterswap/internal/review/mocks"
)

func TestCreate(t *testing.T) {
	cases := []struct {
		name       string
		review     *review.Review
		setupMocks func(repo *reviewmocks.MockRepository, exchangeRepo *exchangemocks.MockRepository)
		wantErr    error
	}{
		{
			name:       "note hors bornes",
			review:     &review.Review{ExchangeID: 1, AuthorID: 1, Note: 6},
			setupMocks: func(repo *reviewmocks.MockRepository, exchangeRepo *exchangemocks.MockRepository) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name:   "echange non termine",
			review: &review.Review{ExchangeID: 1, AuthorID: 1, Note: 5},
			setupMocks: func(repo *reviewmocks.MockRepository, exchangeRepo *exchangemocks.MockRepository) {
				exchangeRepo.EXPECT().
					GetByID(mock.Anything, 1).
					Return(exchange.Exchange{ID: 1, Status: "pending", OwnerID: 2, RequesterID: 1}, nil)
			},
			wantErr: apperrors.ErrExchangeNotDone,
		},
		{
			name:   "auteur hors echange",
			review: &review.Review{ExchangeID: 1, AuthorID: 99, Note: 5},
			setupMocks: func(repo *reviewmocks.MockRepository, exchangeRepo *exchangemocks.MockRepository) {
				exchangeRepo.EXPECT().
					GetByID(mock.Anything, 1).
					Return(exchange.Exchange{ID: 1, Status: "completed", OwnerID: 2, RequesterID: 1}, nil)
			},
			wantErr: apperrors.ErrUnauthorized,
		},
		{
			name:   "deja note",
			review: &review.Review{ExchangeID: 1, AuthorID: 1, Note: 5},
			setupMocks: func(repo *reviewmocks.MockRepository, exchangeRepo *exchangemocks.MockRepository) {
				exchangeRepo.EXPECT().
					GetByID(mock.Anything, 1).
					Return(exchange.Exchange{ID: 1, Status: "completed", OwnerID: 2, RequesterID: 1}, nil)
				repo.EXPECT().HasReviewed(mock.Anything, 1, 1).Return(true, nil)
			},
			wantErr: apperrors.ErrAlreadyReviewed,
		},
		{
			name:   "valide: succes",
			review: &review.Review{ExchangeID: 1, AuthorID: 1, Note: 5, Commentaire: "Top"},
			setupMocks: func(repo *reviewmocks.MockRepository, exchangeRepo *exchangemocks.MockRepository) {
				exchangeRepo.EXPECT().
					GetByID(mock.Anything, 1).
					Return(exchange.Exchange{ID: 1, Status: "completed", OwnerID: 2, RequesterID: 1}, nil)
				repo.EXPECT().HasReviewed(mock.Anything, 1, 1).Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*review.Review")).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := reviewmocks.NewMockRepository(t)
			exchangeRepo := exchangemocks.NewMockRepository(t)
			tc.setupMocks(repo, exchangeRepo)
			svc := review.NewService(repo, exchangeRepo)

			err := svc.Create(context.Background(), tc.review)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 2, tc.review.TargetID)
			}
		})
	}
}
