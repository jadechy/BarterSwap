package httpserver

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/jadechy/barterswap/docs"
	exchangehttp "github.com/jadechy/barterswap/internal/exchange"
	offerhttp "github.com/jadechy/barterswap/internal/offer"
	reviewhttp "github.com/jadechy/barterswap/internal/review"
	userhttp "github.com/jadechy/barterswap/internal/user"
)

// Handlers regroupe tous les handlers de domaine nécessaires au routeur.
// Ça évite une signature NewRouter(a, b, c, d) où l'ordre des paramètres
// devient source d'erreur à mesure que le nombre de domaines grandit.
type Handlers struct {
	User     *userhttp.Handler
	Offer    *offerhttp.Handler
	Exchange *exchangehttp.Handler
	Review   *reviewhttp.Handler
}

func NewRouter(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// --- Users ---
	mux.HandleFunc("POST /api/users", h.User.Create)
	mux.HandleFunc("GET /api/users/{id}", h.User.GetByID)
	mux.HandleFunc("PUT /api/users/{id}", h.User.Update)
	mux.HandleFunc("GET /api/users/{id}/skills", h.User.GetSkills)
	mux.HandleFunc("PUT /api/users/{id}/skills", h.User.SetSkills)
	mux.HandleFunc("GET /api/users/{id}/stats", h.User.GetStats)
	// Route montée sous /users/ mais gérée par le domaine review (cf. discussion précédente)
	mux.HandleFunc("GET /api/users/{id}/reviews", h.Review.GetByUserID)

	// --- Offers (anciennement "services") ---
	// mux.HandleFunc("GET /api/services", h.Offer.List)
	mux.HandleFunc("POST /api/services", h.Offer.Create)
	mux.HandleFunc("GET /api/services/{id}", h.Offer.GetByID)
	// mux.HandleFunc("PUT /api/services/{id}", h.Offer.Update)
	// mux.HandleFunc("DELETE /api/services/{id}", h.Offer.Delete)
	mux.HandleFunc("GET /api/services/{id}/reviews", h.Review.GetByServiceID)

	// --- Exchanges ---
	mux.HandleFunc("GET /api/exchanges", h.Exchange.List)
	mux.HandleFunc("POST /api/exchanges", h.Exchange.Create)
	mux.HandleFunc("GET /api/exchanges/{id}", h.Exchange.GetByID)
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", h.Exchange.Accept)
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", h.Exchange.Reject)
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", h.Exchange.Complete)
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", h.Exchange.Cancel)
	mux.HandleFunc("POST /api/exchanges/{id}/review", h.Review.Create)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return mux
}
