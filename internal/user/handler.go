package user

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
// @Summary      Créer un utilisateur
// @Tags         users
// @Security     UserIDAuth
// @Accept       json
// @Produce      json
// @Param        user body User true "Utilisateur à créer"
// @Success      201 {object} User
// @Failure      400 {object} map[string]string
// @Router       /users [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}
	if err := h.manager.Create(r.Context(), &u); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, u)
}

// GetByID godoc
// @Summary      Récupérer un utilisateur par ID
// @Tags         users
// @Security     UserIDAuth
// @Produce      json
// @Param        id path int true "ID utilisateur"
// @Success      200 {object} User
// @Failure      404 {object} map[string]string
// @Router       /users/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	u, err := h.manager.GetByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

// Update godoc
// @Summary      Mettre à jour un utilisateur
// @Tags         users
// @Security     UserIDAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "ID utilisateur"
// @Param        user body User true "Données à mettre à jour"
// @Success      200 {object} User
// @Failure      400 {object} map[string]string
// @Router       /users/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}
	if err := h.manager.Update(r.Context(), id, &u); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

// GetSkills godoc
// @Summary      Lister les compétences d'un utilisateur
// @Tags         users
// @Security     UserIDAuth
// @Produce      json
// @Param        id path int true "ID utilisateur"
// @Success      200 {array} Skill
// @Router       /users/{id}/skills [get]
func (h *Handler) GetSkills(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	skills, err := h.manager.GetSkills(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, skills)
}

// SetSkills godoc
// @Summary      Remplacer les compétences d'un utilisateur
// @Tags         users
// @Security     UserIDAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "ID utilisateur"
// @Param        skills body []Skill true "Nouvelle liste de compétences"
// @Success      200 {array} Skill
// @Failure      400 {object} map[string]string
// @Router       /users/{id}/skills [put]
func (h *Handler) SetSkills(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	var skills []Skill
	if err := json.NewDecoder(r.Body).Decode(&skills); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}
	if err := h.manager.SetSkills(r.Context(), id, skills); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, skills)
}

// GetStats godoc
// @Summary      Statistiques d'un utilisateur
// @Security     UserIDAuth
// @Tags         users
// @Produce      json
// @Param        id path int true "ID utilisateur"
// @Success      200 {object} Stats
// @Router       /users/{id}/stats [get]
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
		return
	}
	stats, err := h.manager.Stats(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, stats)
}
