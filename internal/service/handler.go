package service

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

// Create godoc
// @Summary      Créer une offre de service
// @Tags         services
// @Security     UserIDAuth
// @Accept       json
// @Produce      json
// @Param        service body Service true "Offre à créer"
// @Success      201 {object} Service
// @Failure      400 {object} map[string]string
// @Router       /services [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var o Service
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	if err := h.manager.Create(r.Context(), &o); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, o)
}

// GetByID godoc
// @Summary      Récupérer une offre par ID
// @Tags         services
// @Security     UserIDAuth
// @Produce      json
// @Param        id path int true "ID du service"
// @Success      200 {object} Service
// @Failure      404 {object} map[string]string
// @Router       /services/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}

	o, err := h.manager.GetByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, o)
}
