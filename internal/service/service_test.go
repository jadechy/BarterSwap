package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/service"
	servicemocks "github.com/jadechy/barterswap/internal/service/mocks"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

func TestCreate(t *testing.T) {
	cases := []struct {
		name       string
		service    *service.Service
		setupMocks func(repo *servicemocks.MockRepository, userRepo *usermocks.MockRepository)
		wantErr    error // nil si succès attendu
	}{
		{
			name:       "titre vide",
			service:    &service.Service{Titre: "", Categorie: "Informatique", Credits: 5, ProviderID: 1},
			setupMocks: func(repo *servicemocks.MockRepository, userRepo *usermocks.MockRepository) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name:       "categorie invalide",
			service:    &service.Service{Titre: "Cours de piano", Categorie: "Sorcellerie", Credits: 5, ProviderID: 1},
			setupMocks: func(repo *servicemocks.MockRepository, userRepo *usermocks.MockRepository) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name:       "credits nuls ou negatifs",
			service:    &service.Service{Titre: "Cours de piano", Categorie: "Musique", Credits: 0, ProviderID: 1},
			setupMocks: func(repo *servicemocks.MockRepository, userRepo *usermocks.MockRepository) {},
			wantErr:    apperrors.ErrValidation,
		},
		{
			name:    "provider sans la competence",
			service: &service.Service{Titre: "Cours de piano", Categorie: "Musique", Credits: 5, ProviderID: 1},
			setupMocks: func(repo *servicemocks.MockRepository, userRepo *usermocks.MockRepository) {
				userRepo.EXPECT().
					GetSkills(mock.Anything, 1).
					Return([]user.Skill{{Nom: "Jardinage", Niveau: "expert"}}, nil)
			},
			wantErr: apperrors.ErrValidation,
		},
		{
			name:    "provider avec la competence: succes",
			service: &service.Service{Titre: "Cours de piano", Categorie: "Musique", Credits: 5, ProviderID: 1},
			setupMocks: func(repo *servicemocks.MockRepository, userRepo *usermocks.MockRepository) {
				userRepo.EXPECT().
					GetSkills(mock.Anything, 1).
					Return([]user.Skill{{Nom: "Musique", Niveau: "expert"}}, nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*service.Service")).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := servicemocks.NewMockRepository(t)
			userRepo := usermocks.NewMockRepository(t)
			tc.setupMocks(repo, userRepo)
			svc := service.NewService(repo, userRepo)

			err := svc.Create(context.Background(), tc.service)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	cases := []struct {
		name       string
		id         int
		setupMocks func(repo *servicemocks.MockRepository)
		wantErr    error
	}{
		{
			name: "non trouve",
			id:   99,
			setupMocks: func(repo *servicemocks.MockRepository) {
				repo.EXPECT().GetByID(mock.Anything, 99).Return(service.Service{}, apperrors.ErrNotFound)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "trouve",
			id:   1,
			setupMocks: func(repo *servicemocks.MockRepository) {
				repo.EXPECT().GetByID(mock.Anything, 1).Return(service.Service{ID: 1, Titre: "Cours"}, nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := servicemocks.NewMockRepository(t)
			userRepo := usermocks.NewMockRepository(t)
			tc.setupMocks(repo)
			svc := service.NewService(repo, userRepo)

			_, err := svc.GetByID(context.Background(), tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdate_CategorieInvalide(t *testing.T) {
	repo := servicemocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := service.NewService(repo, userRepo)

	o := &service.Service{Titre: "X", Categorie: "Inexistant"}
	err := svc.Update(context.Background(), 1, o)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrValidation)
	repo.AssertNotCalled(t, "Update")
}

func TestDelete_ErreurRepository_Propagee(t *testing.T) {
	repo := servicemocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := service.NewService(repo, userRepo)

	repo.EXPECT().Delete(mock.Anything, 5).Return(assert.AnError)

	err := svc.Delete(context.Background(), 5)

	require.Error(t, err)
}

func TestList(t *testing.T) {
	repo := servicemocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := service.NewService(repo, userRepo)

	f := service.ListFilter{Categorie: "Musique"}
	repo.EXPECT().List(mock.Anything, f).Return([]service.Service{{ID: 1, Titre: "Cours"}}, nil)

	result, err := svc.List(context.Background(), f)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestList_ErreurRepository_Propagee(t *testing.T) {
	repo := servicemocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := service.NewService(repo, userRepo)

	f := service.ListFilter{}
	repo.EXPECT().List(mock.Anything, f).Return(nil, assert.AnError)

	_, err := svc.List(context.Background(), f)

	require.Error(t, err)
}

func TestCreate_ErreurGetSkills_Propagee(t *testing.T) {
	repo := servicemocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)
	svc := service.NewService(repo, userRepo)

	o := &service.Service{Titre: "Cours", Categorie: "Musique", Credits: 5, ProviderID: 1}
	userRepo.EXPECT().GetSkills(mock.Anything, 1).Return(nil, assert.AnError)

	err := svc.Create(context.Background(), o)

	require.Error(t, err)
	repo.AssertNotCalled(t, "Create")
}
