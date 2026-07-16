#!/bin/sh
set -e

MARKER=".db-initialized"

if [ ! -f "$MARKER" ]; then
    echo "Première initialisation : application du schéma et des seeds..."
    mysql --skip-ssl -h db -u go -ppassword barterswap < schema.sql
    mysql --skip-ssl -h db -u go -ppassword barterswap < seeds.sql
    touch "$MARKER"
    echo "Base initialisée."
else
    echo "Base déjà initialisée (supprime $MARKER pour forcer une réinitialisation)."
fi

go mod tidy
exec go run ./cmd/api