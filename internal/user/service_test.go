package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

func TestCreate(t *testing.T) {
	cases := []struct {
		name       string
		user       *user.User
		setupMocks func(repo *usermocks.MockRepository)
		wantErr    error
	}{
		{
			name:       "pseudo vide",
			user:       &user.User{Pseudo: "", Ville: "Paris"},
			setupMocks: func(repo *usermocks.MockRepository) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name: "pseudo blancs uniquement",
			user: &user.User{Pseudo: "   ", Ville: "Paris"},
			setupMocks: func(repo *usermocks.MockRepository) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name: "pseudo valide: succes",
			user: &user.User{Pseudo: "cecile", Ville: "Paris"},
			setupMocks: func(repo *usermocks.MockRepository) {
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := usermocks.NewMockRepository(t)
			tc.setupMocks(repo)
			svc := user.NewService(repo)

			err := svc.Create(context.Background(), tc.user)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdate_PseudoVide(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	u := &user.User{Pseudo: "  "}
	err := svc.Update(context.Background(), 1, u)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestSetSkills(t *testing.T) {
	cases := []struct {
		name       string
		skills     []user.Skill
		setupMocks func(repo *usermocks.MockRepository, skills []user.Skill)
		wantErr    error
	}{
		{
			name:       "niveau invalide",
			skills:     []user.Skill{{Nom: "Musique", Niveau: "legendaire"}},
			setupMocks: func(repo *usermocks.MockRepository, skills []user.Skill) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name:       "nom vide",
			skills:     []user.Skill{{Nom: "  ", Niveau: "expert"}},
			setupMocks: func(repo *usermocks.MockRepository, skills []user.Skill) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name:   "valide: succes",
			skills: []user.Skill{{Nom: "Musique", Niveau: "expert"}},
			setupMocks: func(repo *usermocks.MockRepository, skills []user.Skill) {
				repo.EXPECT().SetSkills(mock.Anything, 1, skills).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := usermocks.NewMockRepository(t)
			tc.setupMocks(repo, tc.skills)
			svc := user.NewService(repo)

			err := svc.SetSkills(context.Background(), 1, tc.skills)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetByID_NonTrouve(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	repo.EXPECT().GetByID(mock.Anything, 42).Return(user.User{}, apperrors.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 42)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}
