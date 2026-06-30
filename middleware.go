package main

import (
	"log"
	"net/http"
	"strconv"
)

// loggingMiddleware log chaque requête HTTP
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware récupère les panics et renvoie un 500
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware ajoute les headers CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-UserID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware vérifie la présence du header X-UserID
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-UserID")
		if userIDStr == "" {
			http.Error(w, "header X-UserID manquant", http.StatusUnauthorized)
			return
		}
		_, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "X-UserID invalide", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// applyMiddlewares chaîne tous les middlewares
func applyMiddlewares(h http.Handler) http.Handler {
	h = authMiddleware(h)
	h = corsMiddleware(h)
	h = recoveryMiddleware(h)
	h = loggingMiddleware(h)
	return h
}