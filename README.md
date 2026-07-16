# Projet BarterSwap - API d'échange de compétences entre particuliers

## Installation

```bash
git clone git@github.com:jadechy/BarterSwap.git
cd barterswap
docker compose up -d --build
```

### Base de données

Pour repartir de zéro après une modification du schéma :

```bash
docker compose exec db mysql -u go -ppassword -e "DROP DATABASE barterswap; CREATE DATABASE barterswap;"
docker compose cp schema.sql db:/schema.sql
docker compose exec db mysql -u go -ppassword barterswap -e "source /schema.sql"
```

### Seeds

## Démarrage du serveur

```bash
docker compose exec go go mod tidy
docker compose exec go go run .
```

## Endpoints

### Utilisateurs

| Méthode | Path                      | Description                                                         |
| ------- | ------------------------- | ------------------------------------------------------------------- |
| POST    | `/api/users`              | Créer un compte (10 crédits de bienvenue attribués automatiquement) |
| GET     | `/api/users/{id}`         | Profil public d'un utilisateur                                      |
| PUT     | `/api/users/{id}`         | Modifier son profil                                                 |
| GET     | `/api/users/{id}/skills`  | Compétences d'un utilisateur                                        |
| PUT     | `/api/users/{id}/skills`  | Définir ses compétences (écrase les précédentes)                    |
| GET     | `/api/users/{id}/reviews` | Avis reçus par un utilisateur                                       |
| GET     | `/api/users/{id}/stats`   | Statistiques d'un utilisateur                                       |

### Services

| Méthode | Path                         | Description                             |
| ------- | ---------------------------- | --------------------------------------- |
| GET     | `/api/services`              | Liste des services (filtres optionnels) |
| POST    | `/api/services`              | Créer une annonce de service            |
| GET     | `/api/services/{id}`         | Détail d'un service                     |
| PUT     | `/api/services/{id}`         | Modifier son annonce                    |
| DELETE  | `/api/services/{id}`         | Supprimer son annonce                   |
| GET     | `/api/services/{id}/reviews` | Avis sur un service                     |

**Filtres disponibles sur `GET /api/services`** (query parameters) :

- `?categorie={cat}` — filtrer par catégorie
- `?ville={ville}` — filtrer par ville
- `?search={mot-clé}` — recherche textuelle (titre/description)

### Échanges

| Méthode | Path                           | Description                           |
| ------- | ------------------------------ | ------------------------------------- |
| POST    | `/api/exchanges`               | Créer une demande d'échange           |
| GET     | `/api/exchanges`               | Liste des échanges (requêtes + reçus) |
| GET     | `/api/exchanges/{id}`          | Détail d'un échange                   |
| PUT     | `/api/exchanges/{id}/accept`   | Accepter une demande                  |
| PUT     | `/api/exchanges/{id}/reject`   | Refuser une demande                   |
| PUT     | `/api/exchanges/{id}/complete` | Marquer comme terminé                 |
| PUT     | `/api/exchanges/{id}/cancel`   | Annuler (demandeur ou offreur)        |
| POST    | `/api/exchanges/{id}/review`   | Donner un avis sur un échange terminé |

**Filtre disponible sur `GET /api/exchanges`** :

- `?status={status}` — filtrer par statut (`pending`, `accepted`, `rejected`, `cancelled`, `completed`)

### Authentification

Toutes les routes (sauf `POST /api/users`) nécessitent le header `X-UserID` correspondant à l'ID de l'utilisateur effectuant la requête.

### Linter

- go vet
- go fmt
- go errcheck

```
docker compose exec go go build ./...
docker compose exec go go vet ./...
docker compose exec go errcheck ./...
docker compose exec go gofmt -l .
```

Si les 4 commandes ne retournent absolument rien, alors le repo est valide sur la compilation, vet, les erreurs ignorées et le formatage.
