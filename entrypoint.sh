#!/bin/sh
set -e

echo "Running Prisma migrations..."
cd pkg && prisma-client-go db push

echo "Starting application..."
cd ..
exec "$@"
