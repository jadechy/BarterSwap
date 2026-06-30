package main

import (
	"log"
	"os"
	"net/http"
)

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", createUserHandler(db))
	mux.HandleFunc("GET /api/users/{id}", getUserHandler(db))
	mux.HandleFunc("PUT /api/users/{id}", updateUserHandler(db))
	mux.HandleFunc("GET /api/users/{id}/skills", getUserSkillsHandler(db))
	mux.HandleFunc("PUT /api/users/{id}/skills", updateUserSkillsHandler(db))
	mux.HandleFunc("GET /api/users/{id}/reviews", getUserReviewsHandler(db))
	mux.HandleFunc("GET /api/users/{id}/stats", getUserStatsHandler(db))

	mux.HandleFunc("GET /api/services", listServicesHandler(db))
	mux.HandleFunc("POST /api/services", createServiceHandler(db))
	mux.HandleFunc("GET /api/services/{id}", getServiceHandler(db))
	mux.HandleFunc("PUT /api/services/{id}", updateServiceHandler(db))
	mux.HandleFunc("DELETE /api/services/{id}", deleteServiceHandler(db))
	// mux.HandleFunc("GET /api/services?categorie={cat}", listServicesHandler(db))
	// mux.HandleFunc("GET /api/services?ville={ville}", listServicesHandler(db))
	// mux.HandleFunc("GET /api/services?search={keyword}", listServicesHandler(db))
	mux.HandleFunc("GET /api/services/{id}/reviews", getServiceReviewsHandler(db))

	mux.HandleFunc("GET /api/exchanges", listExchangesHandler(db))
	mux.HandleFunc("POST /api/exchanges", createExchangeHandler(db))
	mux.HandleFunc("GET /api/exchanges/{id}", getExchangeHandler(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", acceptExchangeHandler(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", rejectExchangeHandler(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", completeExchangeHandler(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", cancelExchangeHandler(db))
	// mux.HandleFunc("GET /api/exchanges?status={status}", listExchangesHandler(db))
	mux.HandleFunc("POST /api/exchanges/{id}/review", createReviewHandler(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur démarré sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, applyMiddlewares(mux)))
}