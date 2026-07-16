package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/jadechy/barterswap/internal/apperrors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, u *User) error {
	if strings.TrimSpace(u.Pseudo) == "" {
		return fmt.Errorf("le pseudo est requis: %w", apperrors.ErrValidation)
	}
	return s.repo.Create(ctx, u)
}

func (s *Service) GetByID(ctx context.Context, id int) (User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, u *User) error {
	if strings.TrimSpace(u.Pseudo) == "" {
		return fmt.Errorf("le pseudo est requis: %w", apperrors.ErrValidation)
	}
	return s.repo.Update(ctx, id, u)
}

func (s *Service) GetSkills(ctx context.Context, userID int) ([]Skill, error) {
	return s.repo.GetSkills(ctx, userID)
}

func (s *Service) SetSkills(ctx context.Context, userID int, skills []Skill) error {
	for _, sk := range skills {
		if !contains(NiveauxValides, sk.Niveau) {
			return fmt.Errorf("niveau invalide %q: %w", sk.Niveau, apperrors.ErrValidation)
		}
		if strings.TrimSpace(sk.Nom) == "" {
			return fmt.Errorf("le nom de la compétence est requis: %w", apperrors.ErrValidation)
		}
	}
	return s.repo.SetSkills(ctx, userID, skills)
}

func (s *Service) Stats(ctx context.Context, userID int) (Stats, error) {
	return s.repo.Stats(ctx, userID)
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, value) {
			return true
		}
	}
	return false
}
