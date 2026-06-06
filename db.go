package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// openDB initialise la connexion à la base de données
func openDB() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN non défini")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur ouverture DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur connexion DB: %w", err)
	}

	log.Println("Connexion à la base de données établie")
	return db, nil
}