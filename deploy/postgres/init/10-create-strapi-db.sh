#!/bin/bash
# Creates a SEPARATE database + role for Strapi, isolated from the main app DB.
# Runs once, on first initialization of the Postgres data volume.
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER ${STRAPI_DB_USER:-strapi} WITH PASSWORD '${STRAPI_DB_PASSWORD:-strapi_dev_pw}';
    CREATE DATABASE ${STRAPI_DB_NAME:-strapi} OWNER ${STRAPI_DB_USER:-strapi};
EOSQL

echo "Created Strapi database '${STRAPI_DB_NAME:-strapi}' owned by '${STRAPI_DB_USER:-strapi}'."
