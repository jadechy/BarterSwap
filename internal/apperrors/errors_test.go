package apperrors_test

import (
	"errors"
	"testing"

	"github.com/jadechy/barterswap/internal/apperrors"
)

func TestValidationError_Error(t *testing.T) {
	err := apperrors.ValidationError{Champ: "pseudo", Message: "requis"}
	want := "validation échouée sur pseudo : requis"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestValidationError_Is_ErrValidation(t *testing.T) {
	err := apperrors.ValidationError{Champ: "pseudo", Message: "requis"}
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Error("ValidationError devrait être détecté comme ErrValidation via errors.Is")
	}
}
