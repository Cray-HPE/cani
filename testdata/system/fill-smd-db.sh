#! /usr/bin/env bash
#
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
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

