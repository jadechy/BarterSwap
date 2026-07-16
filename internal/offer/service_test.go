package offer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/offer"
	offermocks "github.com/jadechy/barterswap/internal/offer/mocks"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

func TestCreate_TitreVide_RetourneErreurValidation(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	o := &offer.Offer{Titre: "", Categorie: "Informatique", Credits: 5, ProviderID: 1}
	err := svc.Create(context.Background(), o)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestCreate_CategorieInvalide_RetourneErreurValidation(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	o := &offer.Offer{Titre: "Cours de piano", Categorie: "Sorcellerie", Credits: 5, ProviderID: 1}
	err := svc.Create(context.Background(), o)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestCreate_CreditsNegatifsOuNuls_RetourneErreurValidation(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	o := &offer.Offer{Titre: "Cours de piano", Categorie: "Musique", Credits: 0, ProviderID: 1}
	err := svc.Create(context.Background(), o)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestCreate_ProviderSansLaCompetence_RetourneErreurValidation(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	o := &offer.Offer{Titre: "Cours de piano", Categorie: "Musique", Credits: 5, ProviderID: 1}

	userRepo.EXPECT().
		GetSkills(mock.Anything, 1).
		Return([]user.Skill{{Nom: "Jardinage", Niveau: "expert"}}, nil)

	err := svc.Create(context.Background(), o)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestCreate_ProviderAvecLaCompetence_Succes(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	o := &offer.Offer{Titre: "Cours de piano", Categorie: "Musique", Credits: 5, ProviderID: 1}

	userRepo.EXPECT().
		GetSkills(mock.Anything, 1).
		Return([]user.Skill{{Nom: "Musique", Niveau: "expert"}}, nil)
	repo.EXPECT().
		Create(mock.Anything, o).
		Return(nil)

	err := svc.Create(context.Background(), o)

	require.NoError(t, err)
}

func TestGetByID_NonTrouve_RetourneErrNotFound(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	repo.EXPECT().
		GetByID(mock.Anything, 99).
		Return(offer.Offer{}, apperrors.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 99)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestUpdate_CategorieInvalide_RetourneErreurValidation(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	o := &offer.Offer{Titre: "X", Categorie: "Inexistant"}
	err := svc.Update(context.Background(), 1, o)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestDelete_ErreurRepository_Propagee(t *testing.T) {
	repo := offermocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := offer.NewService(repo, userRepo)

	repo.EXPECT().
		Delete(mock.Anything, 5).
		Return(assert.AnError)

	err := svc.Delete(context.Background(), 5)

	require.Error(t, err)
}
