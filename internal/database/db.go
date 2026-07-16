package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Open initialise la connexion à la base de données MySQL à partir de DB_DSN.
func Open() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN non défini")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur ouverture DB: %w", err)
	}

	// Limites de pool explicites : évite d'épuiser les connexions MySQL
	// sous charge, et borne la durée de vie des connexions pour absorber
	// les redémarrages/rotations côté serveur DB sans erreurs en cascade.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur connexion DB: %w", err)
	}

	log.Println("Connexion à la base de données établie")
	return db, nil
}
