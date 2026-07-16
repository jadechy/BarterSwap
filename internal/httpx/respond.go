package httpx

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jadechy/barterswap/internal/apperrors"
)

// WriteJSON écrit une réponse JSON avec le status donné.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Le corps est déjà parti (header + status écrits), on ne peut plus
		// changer le status ici. On journalise pour ne pas perdre l'info silencieusement.
		log.Printf("httpx: échec encodage JSON: %v", err)
	}
}

// WriteError mappe une erreur de domaine (apperrors) vers le status HTTP approprié
// et écrit une réponse JSON standardisée {"error": "..."}.
func WriteError(w http.ResponseWriter, err error) {
	if valErr, ok := errors.AsType[apperrors.ValidationError](err); ok {
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": valErr.Error(),
			"champ": valErr.Champ,
		})
		return
	}

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, apperrors.ErrValidation),
		errors.Is(err, apperrors.ErrSelfExchange),
		errors.Is(err, apperrors.ErrInsufficientCredits),
		errors.Is(err, apperrors.ErrExchangeNotDone),
		errors.Is(err, apperrors.ErrAlreadyReviewed),
		errors.Is(err, apperrors.ErrInvalidStatus):
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, apperrors.ErrExchangeConflict):
		WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, apperrors.ErrUnauthorized):
		WriteJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	default:
		log.Printf("httpx: erreur interne non mappée: %v", err)
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}
