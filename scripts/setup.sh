#!/usr/bin/env bash
# Installs and starts RodriTV. Safe to run again: existing data and .env are kept.
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${RODRITV_HOME:-$HOME/rodritv}"
ENV_FILE="$INSTALL_DIR/api/.env"

die() {
    echo "error: $*" >&2
    exit 1
}

if docker compose version >/dev/null 2>&1; then
    COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
    COMPOSE=(docker-compose)
else
    die "docker compose is not installed"
fi
docker info >/dev/null 2>&1 || die "cannot connect to docker, is it running?"

mkdir -p "$INSTALL_DIR/db" "$INSTALL_DIR/api"

if [ -f "$ENV_FILE" ]; then
    echo "Keeping the existing $ENV_FILE"
else
    # A new password would not match a database that is already initialized.
    # The db folder may be unreadable here because postgres takes ownership of it.
    if [ -n "$(ls -A "$INSTALL_DIR/db" 2>/dev/null)" ] || [ ! -r "$INSTALL_DIR/db" ]; then
        die "$INSTALL_DIR/db has data but $ENV_FILE is missing, restore that file first"
    fi

    # Limited characters, so it needs no escaping in the .env file or the URL
    while true; do
        read -rsp "PostgreSQL password: " password || die "no password provided"
        echo
        if [[ ! "$password" =~ ^[A-Za-z0-9._~-]{8,}$ ]]; then
            echo "Use at least 8 characters: letters, numbers and . _ ~ -" >&2
            continue
        fi
        read -rsp "Repeat the password: " confirmation || die "no password provided"
        echo
        [ "$password" = "$confirmation" ] && break
        echo "The passwords do not match" >&2
    done

    (
        umask 077
        cat >"$ENV_FILE" <<ENV
# RodriTV configuration

# Installation folder
RODRITV_HOME="$INSTALL_DIR"

# Address and port where the API is published on this machine
API_BIND=127.0.0.1
API_PORT=9290

# PostgreSQL. These are applied only when the database is first created
POSTGRES_USER=rodritv
POSTGRES_PASSWORD=$password
POSTGRES_DB=rodritvdb

# API connection. Must match the PostgreSQL values above
DATABASE_URL=postgres://rodritv:$password@db:5432/rodritvdb?sslmode=disable
ENV
    )
    echo "Created $ENV_FILE"
fi

compose=("${COMPOSE[@]}" --env-file "$ENV_FILE" -f "$REPO_DIR/docker-compose.yaml")

RODRITV_HOME="$INSTALL_DIR" "${compose[@]}" up -d --build

echo
echo "RodriTV started. To manage it:"
echo "  ${compose[*]} ps|logs|stop|up -d|down"
