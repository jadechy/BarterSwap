package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jadechy/barterswap/internal/database"
	"github.com/jadechy/barterswap/internal/dbx"
	"github.com/jadechy/barterswap/internal/exchange"
	"github.com/jadechy/barterswap/internal/httpserver"
	"github.com/jadechy/barterswap/internal/review"
	"github.com/jadechy/barterswap/internal/service"
	"github.com/jadechy/barterswap/internal/user"
)

// @title           Barterswap
// @version         1.0
// @description     API de troc de services entre particuliers, basé sur des crédits-temps.
// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey  UserIDAuth
// @in                          header
// @name                        X-UserID
func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Printf("erreur fermeture DB: %v", cerr)
		}
	}()
	txManager := dbx.NewTxManager(db)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	serviceRepo := service.NewRepository(db)
	serviceService := service.NewService(serviceRepo, userRepo)
	serviceHandler := service.NewHandler(serviceService)

	exchangeRepo := exchange.NewRepository(db)
	exchangeService := exchange.NewService(exchangeRepo, txManager, serviceRepo, userRepo, userRepo)
	exchangeHandler := exchange.NewHandler(exchangeService)

	reviewRepo := review.NewRepository(db)
	reviewService := review.NewService(reviewRepo, exchangeRepo)
	reviewHandler := review.NewHandler(reviewService)

	mux := httpserver.NewRouter(httpserver.Handlers{
		User:     userHandler,
		Service:  serviceHandler,
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
