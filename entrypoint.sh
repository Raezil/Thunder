#!/bin/bash
set -euo pipefail

# Set the host using the PGHOST environment variable, or default to host.docker.internal
HOST=${PGHOST:-localhost}
POSTGRES_PORT=5432
POSTGRES_USER=postgres
echo "Connecting to PostgreSQL at $HOST:$POSTGRES_PORT..."

# Wait for PostgreSQL to become available
while ! nc -z "$HOST" "$POSTGRES_PORT"; do
  echo "Waiting for PostgreSQL to start at $HOST:$POSTGRES_PORT..."
  sleep 1
done

echo "PostgreSQL is up!"

# Optional: Create the database 'postgres' if it does not exist
if ! psql -h "$HOST" -U "$POSTGRES_USER" -lqt | cut -d \| -f 1 | grep -qw postgres; then
  echo "Creating database 'postgres'..."
  psql -h "$HOST" -U "$POSTGRES_USER" -c "CREATE DATABASE thunder;"
else
  echo "Database 'postgres' already exists."
fi

# Run Prisma migrations (or db push) to create tables
echo "Running Prisma db push..."
prisma-client-go db push

# Start the Go application
echo "Starting Go application..."
exec go run ./server/main.go
