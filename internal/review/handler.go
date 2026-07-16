package review

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jadechy/barterswap/internal/httpx"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) *Handler {
	return &Handler{manager: manager}
}

func currentUserID(r *http.Request) (int, error) {
	return strconv.Atoi(r.Header.Get("X-UserID"))
}

// Create godoc
// @Summary      Laisser un avis sur un échange terminé
// @Tags         reviews
// @Security     UserIDAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "ID de l'échange"
// @Param        review body object true "Note et commentaire"
// @Success      201 {object} Review
// @Failure      400 {object} map[string]string
// @Router       /exchanges/{id}/review [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	exchangeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}

	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "X-UserID requis"})
		return
	}

	var body struct {
		Note        int    `json:"note"`
		Commentaire string `json:"commentaire"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	rev := Review{
		ExchangeID:  exchangeID,
		AuthorID:    userID,
		Note:        body.Note,
		Commentaire: body.Commentaire,
	}

	if err := h.manager.Create(r.Context(), &rev); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, rev)
}

// GetByUserID godoc
// @Summary      Avis reçus par un utilisateur
// @Tags         reviews
// @Security     UserIDAuth
// @Produce      json
// @Param        id path int true "ID utilisateur"
// @Success      200 {array} Review
// @Router       /users/{id}/reviews [get]
func (h *Handler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	reviews, err := h.manager.GetByUserID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, reviews)
}

// GetByServiceID godoc
// @Summary      Avis reçus sur une offre
// @Tags         reviews
// @Security     UserIDAuth
// @Produce      json
// @Param        id path int true "ID de l'offre"
// @Success      200 {array} Review
// @Router       /services/{id}/reviews [get]
func (h *Handler) GetByServiceID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	reviews, err := h.manager.GetByServiceID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, reviews)
}
