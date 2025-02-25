#!/bin/sh
set -e

echo "Running Prisma migrations..."
prisma-client-go db push

echo "Starting application..."
exec "$@"
