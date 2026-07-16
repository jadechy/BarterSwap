package service_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/service"
	servicemocks "github.com/jadechy/barterswap/internal/service/mocks"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

// newTestHandler construit un Handler avec un vrai Service, mais des repositories mockés.
func newTestHandler(t *testing.T) (*service.Handler, *servicemocks.MockRepository, *usermocks.MockRepository) {
	repo := servicemocks.NewMockRepository(t)
	userRepo := usermocks.NewMockRepository(t)

	svc := service.NewService(repo, userRepo)
	h := service.NewHandler(svc)

	return h, repo, userRepo
}

// --- Create ---

func TestHandler_Create_Succes(t *testing.T) {
	h, repo, userRepo := newTestHandler(t)

	userRepo.EXPECT().
		GetSkills(mock.Anything, 1).
		Return([]user.Skill{{Nom: "Informatique"}}, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*service.Service")).
		Return(nil)

	body := strings.NewReader(`{
		"provider_id": 1,
		"titre": "Cours de programmation",
		"categorie": "Informatique",
		"duree_minutes": 60,
		"credits": 5
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/services", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var got service.Service
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "Cours de programmation", got.Titre)
	assert.Equal(t, "Informatique", got.Categorie)
}

func TestHandler_Create_JSONInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{invalide`)
	req := httptest.NewRequest(http.MethodPost, "/api/services", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_TitreVide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{
		"provider_id": 1,
		"titre": "",
		"categorie": "Informatique",
		"duree_minutes": 60,
		"credits": 5
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/services", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var got map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Contains(t, got["error"], "titre")
}

func TestHandler_Create_CategorieInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{
		"provider_id": 1,
		"titre": "Cours de programmation",
		"categorie": "Astrologie",
		"duree_minutes": 60,
		"credits": 5
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/services", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_CreditsNegatifs_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{
		"provider_id": 1,
		"titre": "Cours de programmation",
		"categorie": "Informatique",
		"duree_minutes": 60,
		"credits": 0
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/services", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_CompetenceManquante_RetourneBadRequest(t *testing.T) {
	h, _, userRepo := newTestHandler(t)

	userRepo.EXPECT().
		GetSkills(mock.Anything, 1).
		Return([]user.Skill{{Nom: "Jardinage"}}, nil)

	body := strings.NewReader(`{
		"provider_id": 1,
		"titre": "Cours de programmation",
		"categorie": "Informatique",
		"duree_minutes": 60,
		"credits": 5
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/services", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var got map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Contains(t, got["error"], "compétence")
}

// --- GetByID ---

func TestHandler_GetByID_Succes(t *testing.T) {
	h, repo, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(service.Service{ID: 1, Titre: "Cours de guitare", Categorie: "Musique"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/services/1", nil)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got service.Service
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "Cours de guitare", got.Titre)
}

func TestHandler_GetByID_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/abc", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetByID_NonTrouve_RetourneNotFound(t *testing.T) {
	h, repo, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 999).
		Return(service.Service{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/services/999", nil)
	req.SetPathValue("id", "999")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
