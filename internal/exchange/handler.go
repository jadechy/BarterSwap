package exchange

import (
	"encoding/json"
	"errors"
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

func currentUserID(r *http.Request) (int, error) {
	return strconv.Atoi(r.Header.Get("X-UserID"))
}

func parseAction(r *http.Request) (exchangeID, userID int, err error) {
	exchangeID, err = strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, 0, errors.New("id invalide")
	}
	userID, err = currentUserID(r)
	if err != nil {
		return 0, 0, errors.New("X-UserID requis")
	}
	return exchangeID, userID, nil
}

// Create godoc
// @Summary      Créer une demande d'échange
// @Tags         exchanges
// @Security     UserIDAuth
// @Accept       json
// @Produce      json
// @Param        exchange body object true "service_id"
// @Success      201 {object} Exchange
// @Router       /exchanges [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "X-UserID requis"})
		return
	}

	var body struct {
		ServiceID int `json:"service_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	e, err := h.service.Create(r.Context(), userID, body.ServiceID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, e)
}

// List godoc
// @Summary      Lister mes échanges
// @Tags         exchanges
// @Security     UserIDAuth
// @Produce      json
// @Param        status query string false "Filtrer par statut"
// @Success      200 {array} Exchange
// @Router       /exchanges [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "X-UserID requis"})
		return
	}
	status := r.URL.Query().Get("status")

	exchanges, err := h.service.List(r.Context(), userID, status)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, exchanges)
}

// GetByID godoc
// @Summary      Récupérer un échange par ID
// @Tags         exchanges
// @Security     UserIDAuth
// @Produce      json
// @Param        id path int true "ID échange"
// @Success      200 {object} Exchange
// @Failure      404 {object} map[string]string
// @Router       /exchanges/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	e, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, e)
}

// Accept godoc
// @Summary      Accepter un échange
// @Tags         exchanges
// @Security     UserIDAuth
// @Param        id path int true "ID échange"
// @Success      200 {object} map[string]string
// @Router       /exchanges/{id}/accept [put]
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	id, userID, err := parseAction(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.service.Accept(r.Context(), id, userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "échange accepté"})
}

// Reject godoc
// @Summary      Refuser un échange
// @Tags         exchanges
// @Security     UserIDAuth
// @Param        id path int true "ID échange"
// @Success      200 {object} map[string]string
// @Router       /exchanges/{id}/reject [put]
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	id, userID, err := parseAction(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.service.Reject(r.Context(), id, userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "échange refusé"})
}

// Complete godoc
// @Summary      Terminer un échange accepté
// @Tags         exchanges
// @Security     UserIDAuth
// @Param        id path int true "ID échange"
// @Success      200 {object} map[string]string
// @Router       /exchanges/{id}/complete [put]
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	id, userID, err := parseAction(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.service.Complete(r.Context(), id, userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "échange terminé"})
}

// Cancel godoc
// @Summary      Annuler un échange
// @Tags         exchanges
// @Security     UserIDAuth
// @Param        id path int true "ID échange"
// @Success      200 {object} map[string]string
// @Router       /exchanges/{id}/cancel [put]
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, userID, err := parseAction(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.service.Cancel(r.Context(), id, userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "échange annulé"})
}
