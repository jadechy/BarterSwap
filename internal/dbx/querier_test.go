package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/jadechy/barterswap/internal/dbx"
)

func TestWithTx(t *testing.T) {
	cases := []struct {
		name      string
		fn        func(q dbx.Querier) error
		expectFn  func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "succes: commit",
			fn: func(q dbx.Querier) error {
				return nil
			},
			expectFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "erreur: rollback",
			fn: func(q dbx.Querier) error {
				return errors.New("boom")
			},
			expectFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("erreur ouverture sqlmock: %v", err)
			}
			defer db.Close()

			tc.expectFn(mock)

			tm := dbx.NewTxManager(db)
			err = tm.WithTx(context.Background(), tc.fn)

			if tc.wantErr && err == nil {
				t.Error("erreur attendue, aucune reçue")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("aucune erreur attendue, reçu: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("attentes sqlmock non satisfaites: %v", err)
			}
		})
	}
}
