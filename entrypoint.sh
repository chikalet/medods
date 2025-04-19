#!/bin/sh
set -e

echo "Waiting for PostgreSQL to become available..."
while ! pg_isready -h db -p 5432 -U chikalet -d test; do
  sleep 2
done

echo "Applying database migrations..."
migrate -path /app/migrations -database "postgresql://chikalet:root@db:5432/test?sslmode=disable" up

echo "Starting application..."
exec /app/medods