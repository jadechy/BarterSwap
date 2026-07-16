# BarterSwap — API d'échange de compétences entre particuliers

BarterSwap est une plateforme d'échange de compétences entre particuliers basée sur un système de crédits-temps : chaque heure de service rendue donne droit à une heure de service reçue. Pas d'argent, pas de troc direct — un vrai système de banque de temps.

## Stack technique

- **Langage** : Go (stdlib `net/http` + `encoding/json`, sans framework externe)
- **Base de données** : MySQL 8, accès via `database/sql` (pas d'ORM)
- **Driver** : `github.com/go-sql-driver/mysql`
- **Documentation API** : Swagger (OpenAPI) via `swaggo`
- **Tests** : `testify` + `mockery` pour les mocks, `sqlmock` pour les tests de repository

## Installation

### Avec Docker (recommandé)

```bash
git clone git@github.com:jadechy/BarterSwap.git
cd barterswap
```

Créez un fichier `.env` à la racine du projet avec le contenu suivant :

```dotenv
DB_DSN=go:password@tcp(db:3306)/barterswap?parseTime=true

MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=barterswap
MYSQL_USER=go
MYSQL_PASSWORD=password

PMA_HOST=db
PMA_USER=go
PMA_PASSWORD=password

PORT=8080
```

Puis démarrez les conteneurs :

```bash
docker compose up -d --build
```

Le schéma SQL et les seeds sont appliqués **automatiquement** au premier démarrage par le script d'entrypoint. Le serveur démarre ensuite sur `http://localhost:8080`.

Suivre les logs pendant le démarrage :

```bash
docker compose logs -f go
```

Vous devez voir :

```
Première initialisation : application du schéma et des seeds...
Base initialisée.
Connexion à la base de données établie
Serveur démarré sur le port 8080
```

#### Réinitialiser la base de données

```bash
docker compose exec db mysql -u go -ppassword -e "DROP DATABASE barterswap; CREATE DATABASE barterswap;"
docker compose exec go rm .db-initialized
docker compose restart go
```

#### phpMyAdmin

Une interface d'administration MySQL est disponible sur `http://localhost:8081` (identifiants dans `.env`).

---

### Sans Docker (installation locale)

Prérequis :

- Go 1.23 ou supérieur
- Une instance MySQL 8 accessible (locale ou distante)

```bash
git clone git@github.com:jadechy/BarterSwap.git
cd barterswap
```

Créez la base de données et appliquez le schéma :

```bash
mysql -u root -p -e "CREATE DATABASE barterswap;"
mysql -u root -p barterswap < schema.sql
mysql -u root -p barterswap < seeds.sql
```

Définissez la variable d'environnement `DB_DSN` avec vos identifiants MySQL locaux (notez que l'hôte devient `localhost`, contrairement au `.env` Docker qui utilise `db` comme nom de service interne) :

```bash
export DB_DSN="go:password@tcp(localhost:3306)/barterswap?parseTime=true"
export PORT=8080
```

Installez les dépendances puis lancez le serveur :

```bash
go mod tidy
go run ./cmd/api
```

Le serveur démarre sur `http://localhost:8080` (ou le port défini dans `PORT`).

## Documentation interactive (Swagger)

Une fois le serveur démarré :

```
http://localhost:8080/swagger/index.html
```

Cliquez sur **Authorize** en haut à droite et renseignez un `X-UserID` valide pour tester les routes protégées directement depuis l'interface.

## Authentification

Toutes les routes (sauf `POST /api/users`) nécessitent le header `X-UserID` correspondant à l'ID de l'utilisateur effectuant la requête :

```bash
curl -H "X-UserID: 1" http://localhost:8080/api/users/1
```

Il s'agit d'une authentification simplifiée (pas de JWT ni de session), conforme au périmètre du projet.

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

**Catégories valides** : Informatique, Jardinage, Bricolage, Cuisine, Musique, Langues, Sport, Tutorat, Déménagement, Photographie, Animalier, Couture, Autre

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

**Cycle de vie d'un échange :**

```
pending → accepted → completed
   ↓          ↓
rejected  cancelled
```

### Évaluations

| Méthode | Path                         | Description                           |
| ------- | ---------------------------- | ------------------------------------- |
| POST    | `/api/exchanges/{id}/review` | Donner un avis sur un échange terminé |
| GET     | `/api/users/{id}/reviews`    | Avis reçus par un utilisateur         |
| GET     | `/api/services/{id}/reviews` | Avis sur un service                   |

Un utilisateur ne peut laisser qu'un seul avis par échange. Note de 1 à 5.

## Exemples d'utilisation

### Créer un compte

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"pseudo":"cecile","ville":"Paris"}'
```

### Déclarer une compétence puis publier une annonce de service

```bash
curl -X PUT http://localhost:8080/api/users/1/skills \
  -H "Content-Type: application/json" \
  -H "X-UserID: 1" \
  -d '[{"nom":"Musique","niveau":"expert"}]'

curl -X POST http://localhost:8080/api/services \
  -H "Content-Type: application/json" \
  -H "X-UserID: 1" \
  -d '{"titre":"Cours de piano débutant","categorie":"Musique","duree_minutes":60,"credits":5,"ville":"Paris"}'
```

### Demander un échange

```bash
curl -X POST http://localhost:8080/api/exchanges \
  -H "Content-Type: application/json" \
  -H "X-UserID: 2" \
  -d '{"service_id":1}'
```

### Cas d'erreur : demander un échange sur son propre service

```bash
curl -X POST http://localhost:8080/api/exchanges \
  -H "Content-Type: application/json" \
  -H "X-UserID: 1" \
  -d '{"service_id":1}'
# => 400 {"error":"impossible de s'échanger son propre service"}
```

### Accepter puis compléter un échange

```bash
curl -X PUT http://localhost:8080/api/exchanges/1/accept -H "X-UserID: 1"
curl -X PUT http://localhost:8080/api/exchanges/1/complete -H "X-UserID: 1"
```

## Architecture

Le projet suit une séparation stricte des responsabilités, organisée par domaine métier :

```
cmd/api/            → point d'entrée, wiring uniquement
internal/
├── apperrors/       → sentinelles d'erreur + type ValidationError
├── httpx/           → réponses JSON + mapping erreur → status HTTP
├── dbx/             → gestion des transactions SQL (interface TxRunner)
├── database/        → connexion MySQL
├── httpserver/      → middlewares (auth, CORS, logging, recovery) + routeur
├── user/            → repository → service → handler
├── service/           → idem
├── exchange/        → idem (+ gestion transactionnelle des crédits)
└── review/          → idem
docs/                → spécification Swagger générée
```

Chaque domaine suit le pattern **Repository → Service → Handler** :

- **Repository** : accès SQL brut, aucune logique métier
- **Service** : règles de gestion (validation, transitions d'état, autorisations)
- **Handler** : décodage JSON, extraction des paramètres HTTP, appel du service

La logique métier ne dépend jamais directement de `net/http` ni de `database/sql` — elle communique via des interfaces (`Repository`), ce qui la rend testable indépendamment de la base de données et du serveur HTTP.

## Gestion d'erreurs

Les erreurs métier sont représentées par des sentinelles (`apperrors.ErrNotFound`, `ErrInsufficientCredits`, etc.) et un type structuré `ValidationError{Champ, Message}` pour les erreurs de validation de champ. Le wrapping (`%w`) est utilisé systématiquement pour conserver le contexte et permettre `errors.Is`/`errors.As`. Le mapping vers les codes HTTP est centralisé dans `httpx.WriteError` :

| Erreur                                                                                   | Code HTTP |
| ---------------------------------------------------------------------------------------- | --------- |
| `ErrNotFound`                                                                            | 404       |
| `ValidationError` / erreurs de règle métier (solde, auto-échange, statut invalide, avis) | 400       |
| `ErrExchangeConflict`                                                                    | 409       |
| `ErrUnauthorized`                                                                        | 403       |
| Header `X-UserID` manquant/invalide                                                      | 401 / 400 |

## Linters

```bash
docker compose exec go go build ./...
docker compose exec go go vet ./...
docker compose exec go errcheck ./...
docker compose exec go gofmt -l .
```

Si les 4 commandes ne retournent absolument rien, le code est valide sur la compilation, `go vet`, les erreurs ignorées et le formatage.

## Tests

```bash
make test        # tests détaillés (-v) avec profil de couverture
make cover        # couverture totale agrégée
make cover-html    # rapport HTML navigable
```

Ou directement :

```bash
docker compose exec go go test ./... -v -cover
```

**Couverture actuelle : ≥ 70%** sur les packages métier (`user`, `service`, `exchange`, `review`, `httpserver`, `httpx`, `dbx`, `apperrors`).

Le calcul exclut volontairement :

- `docs/` — code généré par `swag`, aucune logique à tester
- `**/mocks/` — code généré par `mockery`
- `internal/database/` et `cmd/api/` — wiring d'infrastructure pur (connexion DB, câblage des dépendances), sans règle métier à couvrir unitairement

### Stratégie de tests

- **Tests unitaires (services)** : repository mocké via `mockery` + `testify/mock`, cas table-driven pour chaque règle métier (validation, autorisation, transition d'état)
- **Tests d'API (handlers)** : `net/http/httptest`, vérification des codes HTTP et du format de réponse pour chaque cas nominal et d'erreur
- **Tests de repository** : `sqlmock` pour simuler les réponses MySQL sans dépendance à une vraie base
- **Tests de middleware et de routeur** : `httptest` sur la chaîne complète (`Auth`, `CORS`, `Recovery`, `Logging`)

## Licence

Projet académique — ESGI, module Go.
