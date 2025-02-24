#!/usr/bin/env sh
set -e

# Optional: Install small 'wait-for' logic or write your own loop
echo "Waiting for Postgres to be ready..."

# Keep checking until Postgres is ready on host=postgres, port=5432
# (We install pg_isready below with 'apk add postgresql-client')
until pg_isready -h postgres -p 5432 -U postgres > /dev/null 2> /dev/null; do
  sleep 1
done

echo "Postgres is up and running. Proceeding..."

# -- OPTIONAL: Create the DB if it doesn't exist.
#    For this step to succeed, the user 'postgres' must have the rights to create DBs.
#    You also need the 'postgresql-client' installed to run psql.
echo "Ensuring database 'thunder' exists..."
echo "CREATE DATABASE thunder;" | psql -h postgres -U postgres 2>/dev/null || true

# -- Now run Prisma migrations (or db push) to create tables
echo "Running Prisma db push..."
prisma-client-go db push

echo "Starting Go application..."
exec go run ./server/main.go
