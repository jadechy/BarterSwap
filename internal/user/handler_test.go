package user_test

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
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

func newTestHandler(t *testing.T) (*user.Handler, *usermocks.MockRepository) {
	repo := usermocks.NewMockRepository(t)

	svc := user.NewService(repo)
	h := user.NewHandler(svc)

	return h, repo
}

// --- Create ---

func TestHandler_Create_Succes(t *testing.T) {
	h, repo := newTestHandler(t)

	repo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*user.User")).
		Return(nil)

	body := strings.NewReader(`{"pseudo": "jade"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var got user.User
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "jade", got.Pseudo)
}

func TestHandler_Create_JSONInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`{invalide`)
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_PseudoVide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`{"pseudo": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var got map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Contains(t, got["error"], "pseudo")
}

// --- GetByID ---

func TestHandler_GetByID_Succes(t *testing.T) {
	h, repo := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(user.User{ID: 1, Pseudo: "jade"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got user.User
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "jade", got.Pseudo)
}

func TestHandler_GetByID_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/abc", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetByID_NonTrouve_RetourneNotFound(t *testing.T) {
	h, repo := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 999).
		Return(user.User{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	req.SetPathValue("id", "999")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Update ---

func TestHandler_Update_Succes(t *testing.T) {
	h, repo := newTestHandler(t)

	repo.EXPECT().
		Update(mock.Anything, 1, mock.AnythingOfType("*user.User")).
		Return(nil)

	body := strings.NewReader(`{"pseudo": "jade-updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Update(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got user.User
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "jade-updated", got.Pseudo)
}

func TestHandler_Update_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`{"pseudo": "jade"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/abc", body)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.Update(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Update_JSONInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`{invalide`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Update(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Update_PseudoVide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`{"pseudo": ""}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Update(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- GetSkills ---

func TestHandler_GetSkills_Succes(t *testing.T) {
	h, repo := newTestHandler(t)

	repo.EXPECT().
		GetSkills(mock.Anything, 1).
		Return([]user.Skill{{Nom: "Informatique", Niveau: "expert"}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/1/skills", nil)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.GetSkills(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got []user.Skill
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Len(t, got, 1)
	assert.Equal(t, "Informatique", got[0].Nom)
}

func TestHandler_GetSkills_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/abc/skills", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetSkills(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- SetSkills ---

func TestHandler_SetSkills_Succes(t *testing.T) {
	h, repo := newTestHandler(t)

	repo.EXPECT().
		SetSkills(mock.Anything, 1, mock.AnythingOfType("[]user.Skill")).
		Return(nil)

	body := strings.NewReader(`[{"nom": "Jardinage", "niveau": "débutant"}]`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.SetSkills(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got []user.Skill
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Len(t, got, 1)
}

func TestHandler_SetSkills_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`[{"nom": "Jardinage", "niveau": "débutant"}]`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/abc/skills", body)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.SetSkills(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_SetSkills_JSONInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`{invalide`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.SetSkills(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_SetSkills_NiveauInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`[{"nom": "Jardinage", "niveau": "maitre-jedi"}]`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.SetSkills(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_SetSkills_NomVide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	body := strings.NewReader(`[{"nom": "", "niveau": "expert"}]`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.SetSkills(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- GetStats ---

func TestHandler_GetStats_Succes(t *testing.T) {
	h, repo := newTestHandler(t)

	repo.EXPECT().
		Stats(mock.Anything, 1).
		Return(user.Stats{UserID: 1, EchangesCompletes: 3, NoteMoyenne: 4.5}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/1/stats", nil)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.GetStats(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got user.Stats
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, 3, got.EchangesCompletes)
}

func TestHandler_GetStats_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/abc/stats", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetStats(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}