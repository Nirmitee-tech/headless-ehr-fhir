#!/usr/bin/env bash
set -euo pipefail

# Run migrations for a specific tenant or all tenants
# Usage: ./scripts/migrate.sh [tenant_id]

DB_URL="${DATABASE_URL:-postgres://ehr:changeme@localhost:5432/ehr?sslmode=disable}"

if [ -n "${1:-}" ]; then
    echo "Running migrations for tenant: $1"
    psql "$DB_URL" -c "CREATE SCHEMA IF NOT EXISTS tenant_$1;"
    for f in migrations/*.sql; do
        echo "  Applying $(basename "$f")..."
        psql "$DB_URL" -v search_path="tenant_$1" -f "$f"
    done
else
    echo "Running migrations for shared schema..."
    for f in migrations/*.sql; do
        echo "  Applying $(basename "$f")..."
        psql "$DB_URL" -f "$f"
    done
fi

echo "Migrations complete."
