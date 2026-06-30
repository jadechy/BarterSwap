package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrValidation),
		errors.Is(err, ErrSelfExchange),
		errors.Is(err, ErrInsufficientCredits),
		errors.Is(err, ErrExchangeNotDone),
		errors.Is(err, ErrAlreadyReviewed),
		errors.Is(err, ErrInvalidStatus):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrExchangeConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrUnauthorized):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}

func getCurrentUserID(r *http.Request) (int, error) {
	idStr := r.Header.Get("X-UserID")
	return strconv.Atoi(idStr)
}

func createUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
			return
		}

		if err := createUserService(db, &u); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, u)
	}
}

func getUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		u, err := getUserById(db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, u)
	}
}

func updateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
			return
		}

		if err := updateUserService(db, id, &u); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, u)
	}
}

func getUserSkillsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		skills, err := getSkillsByUserID(db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, skills)
	}
}

func updateUserSkillsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		var skills []Skill
		if err := json.NewDecoder(r.Body).Decode(&skills); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
			return
		}

		if err := updateUserSkillsService(db, id, skills); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, skills)
	}
}

func getUserReviewsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		reviews, err := getReviewsByUserID(db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, reviews)
	}
}

func getUserStatsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		stats, err := getUserStats(db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, stats)
	}
}

func listServicesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categorie := r.URL.Query().Get("categorie")
		ville := r.URL.Query().Get("ville")
		search := r.URL.Query().Get("search")

		services, err := listServices(db, categorie, ville, search)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, services)
	}
}

func createServiceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s Service
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
			return
		}

		if err := createServiceService(db, &s); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, s)
	}
}

func getServiceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		s, err := getServiceByID(db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, s)
	}
}

func updateServiceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		var s Service
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
			return
		}

		if err := updateServiceService(db, id, &s); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, s)
	}
}

func deleteServiceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		if err := deleteService(db, id); err != nil {
			writeError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getServiceReviewsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		reviews, err := getReviewsByServiceID(db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, reviews)
	}
}

func createExchangeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getCurrentUserID(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "X-UserID requis"})
			return
		}

		var body struct {
			ServiceID int `json:"service_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
			return
		}

		e, err := createExchangeService(db, userID, body.ServiceID)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, e)
	}
}

func listExchangesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getCurrentUserID(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "X-UserID requis"})
			return
		}

		status := r.URL.Query().Get("status")

		exchanges, err := listExchanges(db, userID, status)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, exchanges)
	}
}

func getExchangeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		e, err := getExchangeByID(db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, e)
	}
}

func acceptExchangeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, userID, err := parseExchangeAction(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := acceptExchangeService(db, id, userID); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "échange accepté"})
	}
}

func rejectExchangeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, userID, err := parseExchangeAction(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := rejectExchangeService(db, id, userID); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "échange refusé"})
	}
}

func completeExchangeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, userID, err := parseExchangeAction(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := completeExchangeService(db, id, userID); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "échange terminé"})
	}
}

func cancelExchangeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, userID, err := parseExchangeAction(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if err := cancelExchangeService(db, id, userID); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "échange annulé"})
	}
}

func parseExchangeAction(r *http.Request) (exchangeID, userID int, err error) {
	exchangeID, err = strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, 0, errors.New("id invalide")
	}
	userID, err = getCurrentUserID(r)
	if err != nil {
		return 0, 0, errors.New("X-UserID requis")
	}
	return exchangeID, userID, nil
}

func createReviewHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exchangeID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id invalide"})
			return
		}

		userID, err := getCurrentUserID(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "X-UserID requis"})
			return
		}

		var body struct {
			Note        int    `json:"note"`
			Commentaire string `json:"commentaire"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
			return
		}

		review := Review{
			ExchangeID:  exchangeID,
			AuthorID:    userID,
			Note:        body.Note,
			Commentaire: body.Commentaire,
		}

		if err := createReviewService(db, &review); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, review)
	}
}