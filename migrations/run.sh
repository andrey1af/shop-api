#!/bin/sh

set -eu

: "${POSTGRES_HOST:=postgres}"
: "${POSTGRES_PORT:=5432}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"

export PGPASSWORD="$POSTGRES_PASSWORD"

attempt=1
max_attempts=30

until pg_isready \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" >/dev/null 2>&1; do
    if [ "$attempt" -ge "$max_attempts" ]; then
        echo "PostgreSQL is unavailable after $max_attempts attempts" >&2
        exit 1
    fi

    echo "Waiting for PostgreSQL ($attempt/$max_attempts)"
    attempt=$((attempt + 1))
    sleep 2
done

psql_db() {
    psql \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d "$POSTGRES_DB" \
        "$@"
}

psql_db -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
    version varchar(255) PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
);
SQL

set -- /migrations/*.sql

if [ ! -f "$1" ]; then
    echo "No migration files found in /migrations" >&2
    exit 1
fi

for migration do

    version=$(basename "$migration")

    psql_db \
        -1 \
        -v ON_ERROR_STOP=1 \
        -v migration="$migration" \
        -v version="$version" <<'SQL'
SELECT pg_advisory_xact_lock(7483921);

SELECT NOT EXISTS (
    SELECT 1
    FROM schema_migrations
    WHERE version = :'version'
) AS should_apply \gset

\if :should_apply
    \echo Applying migration :version
    \i :migration
    INSERT INTO schema_migrations (version) VALUES (:'version');
\else
    \echo Migration :version is already applied
\endif
SQL
done
