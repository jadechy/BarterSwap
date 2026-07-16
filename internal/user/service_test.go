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

func TestCreate_PseudoVide_RetourneErreurValidation(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	u := &user.User{Pseudo: "", Ville: "Paris"}
	err := svc.Create(context.Background(), u)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestCreate_PseudoValide_Succes(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	u := &user.User{Pseudo: "cecile", Ville: "Paris"}
	repo.EXPECT().Create(mock.Anything, u).Return(nil)

	err := svc.Create(context.Background(), u)

	require.NoError(t, err)
}

func TestUpdate_PseudoVide_RetourneErreurValidation(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	u := &user.User{Pseudo: "  "}
	err := svc.Update(context.Background(), 1, u)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestSetSkills_NiveauInvalide_RetourneErreurValidation(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	skills := []user.Skill{{Nom: "Musique", Niveau: "legendaire"}}
	err := svc.SetSkills(context.Background(), 1, skills)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestSetSkills_NomVide_RetourneErreurValidation(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	skills := []user.Skill{{Nom: "  ", Niveau: "expert"}}
	err := svc.SetSkills(context.Background(), 1, skills)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestSetSkills_Valide_Succes(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	skills := []user.Skill{{Nom: "Musique", Niveau: "expert"}}
	repo.EXPECT().SetSkills(mock.Anything, 1, skills).Return(nil)

	err := svc.SetSkills(context.Background(), 1, skills)

	require.NoError(t, err)
}

func TestGetByID_NonTrouve_RetourneErrNotFound(t *testing.T) {
	repo := usermocks.NewMockRepository(t)
	svc := user.NewService(repo)

	repo.EXPECT().
		GetByID(mock.Anything, 42).
		Return(user.User{}, apperrors.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 42)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}
