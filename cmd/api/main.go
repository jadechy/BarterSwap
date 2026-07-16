package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jadechy/barterswap/internal/database"
	"github.com/jadechy/barterswap/internal/dbx"
	"github.com/jadechy/barterswap/internal/exchange"
	"github.com/jadechy/barterswap/internal/httpserver"
	"github.com/jadechy/barterswap/internal/offer"
	"github.com/jadechy/barterswap/internal/review"
	"github.com/jadechy/barterswap/internal/user"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	txManager := dbx.NewTxManager(db)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	offerRepo := offer.NewRepository(db)
	offerService := offer.NewService(offerRepo, userRepo)
	offerHandler := offer.NewHandler(offerService)

	exchangeRepo := exchange.NewRepository(db)
	exchangeService := exchange.NewService(exchangeRepo, txManager, offerRepo, userRepo)
	exchangeHandler := exchange.NewHandler(exchangeService)

	reviewRepo := review.NewRepository(db)
	reviewService := review.NewService(reviewRepo, exchangeRepo)
	reviewHandler := review.NewHandler(reviewService)

	mux := httpserver.NewRouter(httpserver.Handlers{
		User:     userHandler,
		Offer:    offerHandler,
		Exchange: exchangeHandler,
		Review:   reviewHandler,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur démarré sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, httpserver.ApplyMiddlewares(mux)))
}
