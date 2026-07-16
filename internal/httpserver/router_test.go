package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jadechy/barterswap/internal/exchange"
	exchangemocks "github.com/jadechy/barterswap/internal/exchange/mocks"
	"github.com/jadechy/barterswap/internal/httpserver"
	"github.com/jadechy/barterswap/internal/offer"
	offermocks "github.com/jadechy/barterswap/internal/offer/mocks"
	"github.com/jadechy/barterswap/internal/review"
	reviewmocks "github.com/jadechy/barterswap/internal/review/mocks"
	"github.com/jadechy/barterswap/internal/user"
	usermocks "github.com/jadechy/barterswap/internal/user/mocks"
)

func TestNewRouter_RouteConnue_Repond(t *testing.T) {
	userRepo := usermocks.NewMockRepository(t)
	userHandler := user.NewHandler(user.NewService(userRepo))

	offerRepo := offermocks.NewMockRepository(t)
	offerHandler := offer.NewHandler(offer.NewService(offerRepo, userRepo))

	exchangeRepo := exchangemocks.NewMockRepository(t)
	txMock := exchangeRepo // placeholder, remplace par un vrai mock TxRunner si besoin

	_ = txMock
	exchangeHandler := exchange.NewHandler(nil) // voir remarque ci-dessous

	reviewRepo := reviewmocks.NewMockRepository(t)
	reviewHandler := review.NewHandler(review.NewService(reviewRepo, exchangeRepo))

	mux := httpserver.NewRouter(httpserver.Handlers{
		User:     userHandler,
		Offer:    offerHandler,
		Exchange: exchangeHandler,
		Review:   reviewHandler,
	})

	require.NotNil(t, mux)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// On ne vérifie pas un code précis (dépend des mocks), juste que la route existe
	// et ne renvoie pas 404 (ce qui indiquerait une route non enregistrée).
	require.NotEqual(t, http.StatusNotFound, rec.Code)
}

func TestApplyMiddlewares_ChaineBienAppliquee(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	h := httpserver.ApplyMiddlewares(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	req.Header.Set("X-UserID", "1")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTeapot, rec.Code)
}
