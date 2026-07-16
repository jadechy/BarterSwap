package httpserver

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Logging journalise chaque requête HTTP.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Recovery récupère les panics et renvoie un 500 plutôt que de faire crasher le serveur.
func Recovery(next http.Handler) http.Handler {
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

// CORS ajoute les headers nécessaires aux requêtes cross-origin.
func CORS(next http.Handler) http.Handler {
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

// Auth vérifie la présence et la validité du header X-UserID.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/swagger/") {
			next.ServeHTTP(w, r)
			return
		}

		userIDStr := r.Header.Get("X-UserID")
		if userIDStr == "" {
			http.Error(w, "header X-UserID manquant", http.StatusUnauthorized)
			return
		}
		if _, err := strconv.Atoi(userIDStr); err != nil {
			http.Error(w, "X-UserID invalide", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ApplyMiddlewares chaîne tous les middlewares dans l'ordre d'exécution souhaité.
func ApplyMiddlewares(h http.Handler) http.Handler {
	h = Auth(h)
	h = CORS(h)
	h = Recovery(h)
	h = Logging(h)
	return h
}
