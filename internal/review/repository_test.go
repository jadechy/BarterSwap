package review_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/review"
)

func newTestRepo(t *testing.T) (review.Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := review.NewRepository(db)
	return repo, mock, func() { db.Close() }
}

func TestRepository_Create(t *testing.T) {
	repo, mock, cleanup := newTestRepo(t)
	defer cleanup()

	mock.ExpectExec("INSERT INTO reviews").
		WithArgs(1, 2, 3, 5, "Super échange").
		WillReturnResult(sqlmock.NewResult(42, 1))

	r := &review.Review{
		ExchangeID:  1,
		AuthorID:    2,
		TargetID:    3,
		Note:        5,
		Commentaire: "Super échange",
	}
	err := repo.Create(context.Background(), r)

	require.NoError(t, err)
	require.Equal(t, 42, r.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByUserID(t *testing.T) {
	repo, mock, cleanup := newTestRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "exchange_id", "author_id", "target_id", "note", "commentaire", "created_at",
	}).AddRow(1, 10, 2, 3, 5, "Top", "2026-01-01").
		AddRow(2, 11, 4, 3, 4, "Bien", "2026-01-02")

	mock.ExpectQuery("SELECT (.|\n)*FROM reviews WHERE target_id = ?").
		WithArgs(3).
		WillReturnRows(rows)

	reviews, err := repo.GetByUserID(context.Background(), 3)

	require.NoError(t, err)
	require.Len(t, reviews, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByServiceID(t *testing.T) {
	repo, mock, cleanup := newTestRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "exchange_id", "author_id", "target_id", "note", "commentaire", "created_at",
	}).AddRow(1, 10, 2, 3, 5, "Top", "2026-01-01")

	mock.ExpectQuery("SELECT (.|\n)*FROM reviews r(.|\n)*JOIN exchanges(.|\n)*WHERE e.service_id = ?").
		WithArgs(10).
		WillReturnRows(rows)

	reviews, err := repo.GetByServiceID(context.Background(), 10)

	require.NoError(t, err)
	require.Len(t, reviews, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_HasReviewed(t *testing.T) {
	cases := []struct {
		name  string
		count int
		want  bool
	}{
		{name: "deja note", count: 1, want: true},
		{name: "pas encore note", count: 0, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newTestRepo(t)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"count"}).AddRow(tc.count)
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM reviews").
				WithArgs(1, 2).
				WillReturnRows(rows)

			got, err := repo.HasReviewed(context.Background(), 1, 2)

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
