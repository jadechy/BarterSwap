package offer_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/offer"
)

func newTestRepo(t *testing.T) (offer.Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := offer.NewRepository(db)
	return repo, mock, func() { db.Close() }
}

func TestRepository_GetByID(t *testing.T) {
	cases := []struct {
		name       string
		id         int
		setupMocks func(mock sqlmock.Sqlmock)
		wantErr    error
	}{
		{
			name: "trouve",
			id:   1,
			setupMocks: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "provider_id", "titre", "description", "categorie",
					"duree_minutes", "credits", "ville", "actif", "created_at",
				}).AddRow(1, 2, "Cours de piano", "Description", "Musique", 60, 5, "Paris", true, "2026-01-01")
				mock.ExpectQuery("SELECT (.|\n)*FROM services WHERE id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantErr: nil,
		},
		{
			name: "non trouve",
			id:   99,
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.|\n)*FROM services WHERE id = ?").
					WithArgs(99).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "provider_id", "titre", "description", "categorie",
						"duree_minutes", "credits", "ville", "actif", "created_at",
					}))
			},
			wantErr: apperrors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newTestRepo(t)
			defer cleanup()
			tc.setupMocks(mock)

			o, err := repo.GetByID(context.Background(), tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.id, o.ID)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_Create(t *testing.T) {
	repo, mock, cleanup := newTestRepo(t)
	defer cleanup()

	mock.ExpectExec("INSERT INTO services").
		WithArgs(1, "Cours de piano", "Description", "Musique", 60, 5, "Paris").
		WillReturnResult(sqlmock.NewResult(10, 1))

	o := &offer.Offer{
		ProviderID: 1, Titre: "Cours de piano", Description: "Description",
		Categorie: "Musique", DureeMinutes: 60, Credits: 5, Ville: "Paris",
	}
	err := repo.Create(context.Background(), o)

	require.NoError(t, err)
	require.Equal(t, 10, o.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {
	cases := []struct {
		name         string
		id           int
		rowsAffected int64
		wantErr      error
	}{
		{name: "succes", id: 5, rowsAffected: 1, wantErr: nil},
		{name: "non trouve", id: 99, rowsAffected: 0, wantErr: apperrors.ErrNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newTestRepo(t)
			defer cleanup()

			mock.ExpectExec("DELETE FROM services WHERE id = ?").
				WithArgs(tc.id).
				WillReturnResult(sqlmock.NewResult(0, tc.rowsAffected))

			err := repo.Delete(context.Background(), tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
