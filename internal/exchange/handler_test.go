package exchange_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/apperrors"
	"github.com/jadechy/barterswap/internal/dbx"
	dbxmocks "github.com/jadechy/barterswap/internal/dbx/mocks"
	"github.com/jadechy/barterswap/internal/exchange"
	exchangemocks "github.com/jadechy/barterswap/internal/exchange/mocks"
	"github.com/jadechy/barterswap/internal/offer"
	offermocks "github.com/jadechy/barterswap/internal/offer/mocks"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

func newTestHandler(t *testing.T) (*exchange.Handler, *exchangemocks.MockRepository, *offermocks.MockRepository, *usermocks.MockRepository, *dbxmocks.MockTxRunner) {
	repo := exchangemocks.NewMockRepository(t)
	tx := dbxmocks.NewMockTxRunner(t)
	offers := offermocks.NewMockRepository(t)
	users := usermocks.NewMockRepository(t)

	svc := exchange.NewService(repo, tx, offers, users, users)
	h := exchange.NewHandler(svc)

	return h, repo, offers, users, tx
}

func expectWithTx(tx *dbxmocks.MockTxRunner) {
	tx.EXPECT().
		WithTx(mock.Anything, mock.AnythingOfType("func(dbx.Querier) error")).
		RunAndReturn(func(ctx context.Context, fn func(dbx.Querier) error) error {
			return fn(nil)
		})
}

// --- Create ---

func TestHandler_Create_Succes(t *testing.T) {
	h, repo, offers, users, _ := newTestHandler(t)

	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, ProviderID: 2, Credits: 5}, nil)
	repo.EXPECT().
		HasActive(mock.Anything, 10).
		Return(false, nil)
	users.EXPECT().
		GetByID(mock.Anything, 1).
		Return(user.User{ID: 1, CreditBalance: 10}, nil)
	repo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*exchange.Exchange")).
		Return(nil)

	body := strings.NewReader(`{"service_id": 10}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", body)
	req.Header.Set("X-UserID", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var got exchange.Exchange
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, 2, got.OwnerID)
	assert.Equal(t, 1, got.RequesterID)
}

func TestHandler_Create_SansUserID_RetourneUnauthorized(t *testing.T) {
	h, _, _, _, _ := newTestHandler(t)

	body := strings.NewReader(`{"service_id": 10}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", body)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_Create_JSONInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _, _, _ := newTestHandler(t)

	body := strings.NewReader(`{invalide`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", body)
	req.Header.Set("X-UserID", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Create_SoiMeme_RetourneBadRequest(t *testing.T) {
	h, _, offers, _, _ := newTestHandler(t)

	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, ProviderID: 1, Credits: 5}, nil)

	body := strings.NewReader(`{"service_id": 10}`)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", body)
	req.Header.Set("X-UserID", "1")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var got map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Contains(t, got["error"], "propre service")
}

// --- List ---

func TestHandler_List_Succes(t *testing.T) {
	h, repo, _, _, _ := newTestHandler(t)

	repo.EXPECT().
		List(mock.Anything, 1, "pending").
		Return([]exchange.Exchange{
			{ID: 1, OwnerID: 2, RequesterID: 1, Status: "pending"},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges?status=pending", nil)
	req.Header.Set("X-UserID", "1")

	rec := httptest.NewRecorder()
	h.List(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got []exchange.Exchange
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Len(t, got, 1)
}

func TestHandler_List_SansUserID_RetourneUnauthorized(t *testing.T) {
	h, _, _, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges", nil)
	// Pas de header X-UserID

	rec := httptest.NewRecorder()
	h.List(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

// --- GetByID ---

func TestHandler_GetByID_Succes(t *testing.T) {
	h, repo, _, _, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/1", nil)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_GetByID_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/abc", nil)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetByID_NonTrouve_RetourneNotFound(t *testing.T) {
	h, repo, _, _, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 999).
		Return(exchange.Exchange{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/999", nil)
	req.SetPathValue("id", "999")

	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Accept ---

func TestHandler_Accept_Succes(t *testing.T) {
	h, repo, offers, users, tx := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, ServiceID: 10, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)
	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, Credits: 5}, nil)

	expectWithTx(tx)
	users.EXPECT().
		AddCreditTransaction(mock.Anything, mock.Anything, 1, mock.Anything, -5, "spend").
		Return(nil)
	repo.EXPECT().
		UpdateStatus(mock.Anything, mock.Anything, 1, "accepted").
		Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/accept", nil)
	req.Header.Set("X-UserID", "2")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Accept(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Accept_NonProprietaire_RetourneForbidden(t *testing.T) {
	h, repo, _, _, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, Status: "pending"}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/accept", nil)
	req.Header.Set("X-UserID", "99")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Accept(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandler_Accept_IDInvalide_RetourneBadRequest(t *testing.T) {
	h, _, _, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/abc/accept", nil)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.Accept(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_Accept_SansUserID_RetourneBadRequest(t *testing.T) {
	h, _, _, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/accept", nil)
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Accept(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Reject ---

func TestHandler_Reject_Succes(t *testing.T) {
	h, repo, _, _, tx := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)

	expectWithTx(tx)
	repo.EXPECT().
		UpdateStatus(mock.Anything, mock.Anything, 1, "rejected").
		Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/reject", nil)
	req.Header.Set("X-UserID", "2")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Reject(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "échange refusé", got["message"])
}

func TestHandler_Reject_NonProprietaire_RetourneForbidden(t *testing.T) {
	h, repo, _, _, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, Status: "pending"}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/reject", nil)
	req.Header.Set("X-UserID", "99")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Reject(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

// --- Complete ---

func TestHandler_Complete_Succes(t *testing.T) {
	h, repo, offers, users, tx := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, ServiceID: 10, OwnerID: 2, RequesterID: 1, Status: "accepted"}, nil)
	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, Credits: 5}, nil)

	expectWithTx(tx)
	users.EXPECT().
		AddCreditTransaction(mock.Anything, mock.Anything, 2, mock.Anything, 5, "earn").
		Return(nil)
	repo.EXPECT().
		UpdateStatus(mock.Anything, mock.Anything, 1, "completed").
		Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/complete", nil)
	req.Header.Set("X-UserID", "2")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Complete(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Complete_StatutInvalide_RetourneBadRequest(t *testing.T) {
	h, repo, _, _, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/complete", nil)
	req.Header.Set("X-UserID", "2")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Complete(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Cancel ---

func TestHandler_Cancel_Succes(t *testing.T) {
	h, repo, _, _, tx := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)

	expectWithTx(tx)
	repo.EXPECT().
		UpdateStatus(mock.Anything, mock.Anything, 1, "cancelled").
		Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/cancel", nil)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Cancel(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Cancel_AccepteAvecRemboursement_Succes(t *testing.T) {
	h, repo, offers, users, tx := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, ServiceID: 10, OwnerID: 2, RequesterID: 1, Status: "accepted"}, nil)
	offers.EXPECT().
		GetByID(mock.Anything, 10).
		Return(offer.Offer{ID: 10, Credits: 5}, nil)

	expectWithTx(tx)
	users.EXPECT().
		AddCreditTransaction(mock.Anything, mock.Anything, 1, mock.Anything, 5, "refund").
		Return(nil)
	repo.EXPECT().
		UpdateStatus(mock.Anything, mock.Anything, 1, "cancelled").
		Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/cancel", nil)
	req.Header.Set("X-UserID", "1")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Cancel(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_Cancel_NonParticipant_RetourneForbidden(t *testing.T) {
	h, repo, _, _, _ := newTestHandler(t)

	repo.EXPECT().
		GetByID(mock.Anything, 1).
		Return(exchange.Exchange{ID: 1, OwnerID: 2, RequesterID: 1, Status: "pending"}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/cancel", nil)
	req.Header.Set("X-UserID", "99")
	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()
	h.Cancel(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}