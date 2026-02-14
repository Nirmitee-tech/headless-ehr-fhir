#!/usr/bin/env bash
set -euo pipefail

# Load seed/reference data into the database
# Usage: ./scripts/seed.sh [tenant_id]

DB_URL="${DATABASE_URL:-postgres://ehr:changeme@localhost:5432/ehr?sslmode=disable}"
SCHEMA="${1:+tenant_$1}"
SCHEMA="${SCHEMA:-public}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEEDS_DIR="${SCRIPT_DIR}/seeds"

echo "Seeding reference data into schema: $SCHEMA"

# Run each seed file in order
for seed_file in "${SEEDS_DIR}"/*.sql; do
    if [ -f "$seed_file" ]; then
        filename=$(basename "$seed_file")
        echo "  Running ${filename}..."
        psql "$DB_URL" -v search_path="$SCHEMA" -f "$seed_file"
    fi
done

echo "Seeding complete."
