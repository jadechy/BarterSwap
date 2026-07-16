package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jadechy/barterswap/internal/httpserver"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuth_SansHeader_RetourneUnauthorized(t *testing.T) {
	h := httpserver.Auth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_HeaderInvalide_RetourneBadRequest(t *testing.T) {
	h := httpserver.Auth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	req.Header.Set("X-UserID", "abc")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuth_HeaderValide_PasseAuHandlerSuivant(t *testing.T) {
	h := httpserver.Auth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	req.Header.Set("X-UserID", "1")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCORS_RequeteOptions_RetourneNoContent(t *testing.T) {
	h := httpserver.CORS(okHandler())
	req := httptest.NewRequest(http.MethodOptions, "/api/users/1", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestCORS_AjouteLesHeaders(t *testing.T) {
	h := httpserver.CORS(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("header CORS manquant ou incorrect")
	}
}

func TestRecovery_RecupereUnePanic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := httpserver.Recovery(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestLogging_PasseAuHandlerSuivant(t *testing.T) {
	h := httpserver.Logging(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}
}
