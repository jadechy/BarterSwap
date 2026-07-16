package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/jadechy/barterswap/internal/apperrors"
)

type Manager struct {
	repo Repository
}

func NewService(repo Repository) *Manager {
	return &Manager{repo: repo}
}

func (s *Manager) Create(ctx context.Context, u *User) error {
	if strings.TrimSpace(u.Pseudo) == "" {
		return apperrors.ValidationError{Champ: "pseudo", Message: "le pseudo est requis"}
	}
	return s.repo.Create(ctx, u)
}

func (s *Manager) GetByID(ctx context.Context, id int) (User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Manager) Update(ctx context.Context, id int, u *User) error {
	if strings.TrimSpace(u.Pseudo) == "" {
		return apperrors.ValidationError{Champ: "pseudo", Message: "le pseudo est requis"}
	}
	return s.repo.Update(ctx, id, u)
}

func (s *Manager) GetSkills(ctx context.Context, userID int) ([]Skill, error) {
	return s.repo.GetSkills(ctx, userID)
}

func (s *Manager) SetSkills(ctx context.Context, userID int, skills []Skill) error {
	for _, sk := range skills {
		if !contains(NiveauxValides, sk.Niveau) {
			return apperrors.ValidationError{
				Champ:   "niveau",
				Message: fmt.Sprintf("niveau invalide %q", sk.Niveau),
			}
		}
		if strings.TrimSpace(sk.Nom) == "" {
			return apperrors.ValidationError{Champ: "nom", Message: "le nom de la compétence est requis"}
		}
	}
	return s.repo.SetSkills(ctx, userID, skills)
}

func (s *Manager) Stats(ctx context.Context, userID int) (Stats, error) {
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
