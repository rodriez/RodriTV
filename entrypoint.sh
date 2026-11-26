#!/bin/sh
set -eu

ENV_FILE="$HOME/.config/api/.env"

if [ -f "$ENV_FILE" ]; then
    # Export every variable defined in the file
    set -a
    . "$ENV_FILE"
    set +a
else
    echo "env file not found at $ENV_FILE, using the current environment" >&2
fi

# exec makes the server PID 1, so it receives the stop signals
exec "$HOME/bin/startup"
