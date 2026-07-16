package review_test

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
	"github.com/jadechy/barterswap/internal/exchange"
	"github.com/jadechy/barterswap/internal/review"
	reviewmocks "github.com/jadechy/barterswap/internal/review/mocks"
)

func newTestHandler(t *testing.T) (*review.Handler, *reviewmocks.MockRepository, *reviewmocks.MockExchangeGetter) {
	repo := reviewmocks.NewMockRepository(t)
	exchangeRepo := reviewmocks.NewMockExchangeGetter(t)

	svc := review.NewService(repo, exchangeRepo)
	h := review.NewHandler(svc)

	return h, repo, exchangeRepo
}

// --- Create ---

func TestHandler_Create_Succes(t *testing.T) {
	h, repo, exchangeRepo := newTestHandler(t)

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "completed"}, nil)
	repo.EXPECT().
		HasReviewed(mock.Anything, 1, 1).
		Return(false, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*review.Review")).
		Return(nil)

	body := strings.NewReader(`{"note": 5, "commentaire": "Très bon échange"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", body)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var got review.Review
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, 5, got.Note)
	assert.Equal(t, 1, got.AuthorID)
	assert.Equal(t, 2, got.TargetID)
}

func TestHandler_Create_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{"note": 5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/abc/review", body)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_SansUserID_RetourneUnauthorized(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{"note": 5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", body)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_Create_JSONInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{invalide`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", body)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_NoteHorsLimites_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	body := strings.NewReader(`{"note": 0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", body)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_EchangeNonTermine_RetourneBadRequest(t *testing.T) {
	h, _, exchangeRepo := newTestHandler(t)

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)

	body := strings.NewReader(`{"note": 5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", body)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_NonParticipant_RetourneForbidden(t *testing.T) {
	h, _, exchangeRepo := newTestHandler(t)

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 3, Status: "completed"}, nil)

	body := strings.NewReader(`{"note": 5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", body)
	req.Header.Set("X-UserID", "99")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandler_Create_DejaNote_RetourneBadRequest(t *testing.T) {
	h, repo, exchangeRepo := newTestHandler(t)

	exchangeRepo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "completed"}, nil)
	repo.EXPECT().
		HasReviewed(mock.Anything, 1, 1).
		Return(true, nil)

	body := strings.NewReader(`{"note": 5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", body)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var got map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Contains(t, got["error"], "déjà noté")
}

// --- GetByUserID ---

func TestHandler_GetByUserID_Succes(t *testing.T) {
	h, repo, _ := newTestHandler(t)

	repo.EXPECT().
		GetByUserID(mock.Anything, 1).
		Return([]review.Review{
			{ID: 1, ExchangeID: 1, AuthorID: 2, TargetID: 1, Note: 5},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/1/reviews", nil)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.GetByUserID(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got []review.Review
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Len(t, got, 1)
}

func TestHandler_GetByUserID_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/abc/reviews", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetByUserID(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- GetByServiceID ---

func TestHandler_GetByServiceID_Succes(t *testing.T) {
	h, repo, _ := newTestHandler(t)

	repo.EXPECT().
		GetByServiceID(mock.Anything, 10).
		Return([]review.Review{
			{ID: 1, ExchangeID: 1, AuthorID: 2, TargetID: 1, Note: 4},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/services/10/reviews", nil)
	req.SetPathValue("id", "10")

	rec := httptest.NewRecorder()
	h.GetByServiceID(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got []review.Review
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Len(t, got, 1)
}

func TestHandler_GetByServiceID_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/abc/reviews", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetByServiceID(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetByServiceID_ErreurRepo_Retourne500(t *testing.T) {
	h, repo, _ := newTestHandler(t)

	repo.EXPECT().
		GetByServiceID(mock.Anything, 10).
		Return(nil, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/services/10/reviews", nil)
	req.SetPathValue("id", "10")

	rec := httptest.NewRecorder()
	h.GetByServiceID(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}