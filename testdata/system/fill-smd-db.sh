#! /usr/bin/env bash

set -eu

# Stop HSM
docker compose stop cray-smd

# Destroy existing HSM database data 
docker compose stop cray-smd-postgres
docker compose rm cray-smd-postgres -f

# Bring up fresh database
docker compose up -d --no-deps cray-smd-postgres

# Wait for database to comeup
until docker exec -it hms-simulation-environment-cray-smd-postgres-1 pg_isready -h hmsds-postgres -p 5432 -U hmsdsuser -d hmsds
do
    echo "Postgres not up: $?"
    sleep 2
done
echo "Postgres is up: $?"

# Push in database backup
docker exec -i -e PGPASSWORD=hmsdsuser hms-simulation-environment-cray-smd-postgres-1 psql -h hmsds-postgres -p 5432 -U hmsdsuser hmsds < "$1"

# Perform any schema migrations
docker compose up --no-deps cray-smd-init  

# Bring backup HSM
docker compose up -d --no-deps cray-smd

# TODO for some reason the API gateway needs to be restarted
docker compose restart api-gateway

