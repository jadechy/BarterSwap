package httpx_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/httpx"
)

func TestWriteError_StatusMapping(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", apperrors.ErrNotFound, 404},
		{"validation", apperrors.ErrValidation, 400},
		{"validation error struct", apperrors.ValidationError{Champ: "x", Message: "y"}, 400},
		{"wrapped validation", fmt.Errorf("le pseudo est requis: %w", apperrors.ErrValidation), 400},
		{"conflict", apperrors.ErrExchangeConflict, 409},
		{"unauthorized", apperrors.ErrUnauthorized, 403},
		{"unknown", fmt.Errorf("boom"), 500},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			httpx.WriteError(w, tc.err)
			if w.Code != tc.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestWriteJSON_Succes(t *testing.T) {
	w := httptest.NewRecorder()
	httpx.WriteJSON(w, 200, map[string]string{"ok": "true"})

	if w.Code != 200 {
		t.Errorf("got status %d, want 200", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("got Content-Type %q, want application/json", w.Header().Get("Content-Type"))
	}
}
