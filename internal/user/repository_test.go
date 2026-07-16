package user_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/user"
)

func newTestRepo(t *testing.T) (user.Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := user.NewRepository(db)
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
				userRows := sqlmock.NewRows([]string{
					"id", "pseudo", "bio", "ville", "credit_balance", "created_at",
				}).AddRow(1, "cecile", "bio", "Paris", 10, "2026-01-01")
				mock.ExpectQuery("SELECT (.|\n)*FROM users WHERE id = ?").
					WithArgs(1).
					WillReturnRows(userRows)

				skillRows := sqlmock.NewRows([]string{"nom", "niveau"})
				mock.ExpectQuery("SELECT (.|\n)*FROM skills WHERE user_id = ?").
					WithArgs(1).
					WillReturnRows(skillRows)
			},
			wantErr: nil,
		},
		{
			name: "non trouve",
			id:   99,
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.|\n)*FROM users WHERE id = ?").
					WithArgs(99).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "pseudo", "bio", "ville", "credit_balance", "created_at",
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

			u, err := repo.GetByID(context.Background(), tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.id, u.ID)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_Update(t *testing.T) {
	cases := []struct {
		name         string
		id           int
		rowsAffected int64
		wantErr      error
	}{
		{name: "succes", id: 1, rowsAffected: 1, wantErr: nil},
		{name: "non trouve", id: 99, rowsAffected: 0, wantErr: apperrors.ErrNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newTestRepo(t)
			defer cleanup()

			mock.ExpectExec("UPDATE users SET").
				WithArgs("cecile", "bio", "Paris", tc.id).
				WillReturnResult(sqlmock.NewResult(0, tc.rowsAffected))

			u := &user.User{Pseudo: "cecile", Bio: "bio", Ville: "Paris"}
			err := repo.Update(context.Background(), tc.id, u)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetSkills(t *testing.T) {
	repo, mock, cleanup := newTestRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"nom", "niveau"}).
		AddRow("Musique", "expert").
		AddRow("Cuisine", "débutant")
	mock.ExpectQuery("SELECT (.|\n)*FROM skills WHERE user_id = ?").
		WithArgs(1).
		WillReturnRows(rows)

	skills, err := repo.GetSkills(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, skills, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}
