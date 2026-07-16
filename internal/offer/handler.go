package offer

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jadechy/barterswap/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary      Créer une offre de service
// @Tags         offers
// @Security     UserIDAuth
// @Accept       json
// @Produce      json
// @Param        offer body Offer true "Offre à créer"
// @Success      201 {object} Offer
// @Failure      400 {object} map[string]string
// @Router       /services [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var o Offer
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	if err := h.service.Create(r.Context(), &o); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, o)
}

// GetByID godoc
// @Summary      Récupérer une offre par ID
// @Tags         offers
// @Security     UserIDAuth
// @Produce      json
// @Param        id path int true "ID de l'offre"
// @Success      200 {object} Offer
// @Failure      404 {object} map[string]string
// @Router       /services/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}

	o, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, o)
}
