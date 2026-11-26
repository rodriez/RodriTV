#!/usr/bin/env bash
# Removes everything setup.sh created, including the database. Cannot be undone.
set -euo pipefail

INSTALL_DIR="${RODRITV_HOME:-$HOME/rodritv}"
PROJECT="${COMPOSE_PROJECT_NAME:-rodritv}"
IMAGES=(rodritv-api postgres:18-alpine)

die() {
    echo "error: $*" >&2
    exit 1
}

# Without docker the containers would survive the deletion of their data
docker info >/dev/null 2>&1 || die "cannot connect to docker, is it running?"

cat <<MSG
This will permanently delete:
  - the containers and network of the "$PROJECT" project
  - the images: ${IMAGES[*]}
  - the database in $INSTALL_DIR/db
  - the configuration in $INSTALL_DIR/api
MSG
read -rp 'Type "yes" to continue: ' answer || answer=""
[ "$answer" = "yes" ] || die "cancelled, nothing was deleted"

# Found by label, so this works even if the .env file is already gone
filter="label=com.docker.compose.project=$PROJECT"

for id in $(docker ps -aq --filter "$filter"); do
    docker rm -f -v "$id" >/dev/null
done

for id in $(docker network ls -q --filter "$filter"); do
    docker network rm "$id" >/dev/null
done

for image in "${IMAGES[@]}"; do
    if docker image inspect "$image" >/dev/null 2>&1; then
        # Not forced: an image used by another container is kept
        docker image rm "$image" >/dev/null || echo "warning: image $image was not removed" >&2
    fi
done

# Only the folders setup.sh created, never the whole INSTALL_DIR recursively
for dir in "$INSTALL_DIR/db" "$INSTALL_DIR/api"; do
    [ -e "$dir" ] || continue
    # On Linux the database files belong to the postgres user
    rm -rf "$dir" 2>/dev/null || sudo rm -rf "$dir"
done

if [ -d "$INSTALL_DIR" ]; then
    rmdir "$INSTALL_DIR" 2>/dev/null || echo "warning: $INSTALL_DIR has other files, it was kept" >&2
fi

echo "RodriTV was removed."
