package offer

import (
	"context"
	"fmt"
	"strings"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/user"
)

type Service struct {
	repo     Repository
	userRepo user.Repository
}

func NewService(repo Repository, userRepo user.Repository) *Service {
	return &Service{repo: repo, userRepo: userRepo}
}

func (s *Service) Create(ctx context.Context, o *Offer) error {
	if strings.TrimSpace(o.Titre) == "" {
		return fmt.Errorf("le titre est requis: %w", apperrors.ErrValidation)
	}
	if !contains(CategoriesValides, o.Categorie) {
		return fmt.Errorf("catégorie invalide %q: %w", o.Categorie, apperrors.ErrValidation)
	}
	if o.Credits <= 0 {
		return fmt.Errorf("le coût en crédits doit être positif: %w", apperrors.ErrValidation)
	}

	skills, err := s.userRepo.GetSkills(ctx, o.ProviderID)
	if err != nil {
		return err
	}
	hasSkill := false
	for _, sk := range skills {
		if strings.EqualFold(sk.Nom, o.Categorie) {
			hasSkill = true
			break
		}
	}
	if !hasSkill {
		return fmt.Errorf("vous n'avez pas la compétence %q: %w", o.Categorie, apperrors.ErrValidation)
	}

	return s.repo.Create(ctx, o)
}

func (s *Service) GetByID(ctx context.Context, id int) (Offer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Offer, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) Update(ctx context.Context, id int, o *Offer) error {
	if !contains(CategoriesValides, o.Categorie) {
		return fmt.Errorf("catégorie invalide %q: %w", o.Categorie, apperrors.ErrValidation)
	}
	return s.repo.Update(ctx, id, o)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, value) {
			return true
		}
	}
	return false
}
