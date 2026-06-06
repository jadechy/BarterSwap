# Projet BarterSwap - API d'échange de compétences entre particuliers

## Installation
```bash
git clone git@github.com:jadechy/BarterSwap.git
cd barterswap
docker compose up -d --build
```
### Base de données
Appliquer le schéma SQL :

```bash
docker compose cp schema.sql db:/schema.sql
docker compose exec db mysql -u go -ppassword barterswap -e "source /schema.sql"
```
Pour repartir de zéro après une modification du schéma :

```bash
docker compose exec db mysql -u go -ppassword -e "DROP DATABASE barterswap; CREATE DATABASE barterswap;"
docker compose cp schema.sql db:/schema.sql
docker compose exec db mysql -u go -ppassword barterswap -e "source /schema.sql"
```

### Seeds

Insérer les données de test :

```bash
docker compose cp seeds.sql db:/seeds.sql
docker compose exec db mysql -u go -ppassword barterswap -e "source /seeds.sql"
```

## Démarrage du serveur
```bash
docker compose exec go go run .
```