# EPIC-0004: Infraestructura de imágenes, compose, scripts y Makefile (`rodritv-infra`)

- SPECS de origen: SPECS-004 (`.sdlc/specify/SPECS-004-rodritv-infra.md`)
- RFC de origen: RFC-001 (`.sdlc/define/RFC-001-streaming-engine-failover-iptv.md`)
- Roles: Tech Lead + Software Engineer
- Fecha: 2026-11-10
- Proyectos involucrados: `rodritv-infra` (artefactos de `build/`, `deployments/`, `scripts/` y raíz del repo)

## Objetivo técnico
Al completar todas las US queda construida la infraestructura de despliegue de RodriTV: Dockerfiles multi-stage de las 3 imágenes (`rodritv-api`, `rodritv-playlist`, `rodritv-web`) con runtime `alpine:3.21`, usuario no-root y healthchecks propios (la de playlist incluye ffmpeg, Chromium y yt-dlp pinneado con sha256); `deployments/docker-compose.yml` con los 4 servicios (db, api, playlist, web), healthchecks, orden de arranque y límites de memoria; `deployments/.env.example` sin secretos; `scripts/install.sh` idempotente que crea `~/.rodritv/` y levanta el stack; `scripts/release.sh` que versiona las 3 imágenes con el mismo tag semántico; y el `Makefile` raíz como punto de entrada único de desarrollo y release.

## Alcance
- Entra: `build/api/Dockerfile`, `build/web/Dockerfile`, `build/playlist/Dockerfile`, `deployments/docker-compose.yml`, `deployments/.env.example`, `scripts/install.sh`, `scripts/release.sh`, `Makefile`.
- NO entra: las migraciones `migrations/000001_init.*.sql` ni `migrations.go` (ya planificadas en EPIC-0001 / T-0003), `sqlc.yaml` ni las fuentes `scripts/sqlc/*.sql` (EPIC-0001 / T-0008 y ss.), los binarios Go que dockerizan estas imágenes (EPIC-0001/0002/0003), registry de imágenes, CI/CD remoto, backups, TLS, monitoreo externo (fuera de alcance SPECS-004.11).

## Dependencias externas (fuera del Plan)
Ninguna gestión no-codeable. Advertencias técnicas:
- **Los binarios `cmd/api`, `cmd/playlist` y `cmd/web` deben existir** (EPIC-0001/0002/0003) para que los `docker build` completos y el smoke test de compose funcionen end-to-end; los Dockerfiles pueden escribirse y mergearse antes sin romper nada (no forman parte del build de Go).
- **`migrations/` y `sqlc.yaml` deben existir** (T-0003, T-0008) para que los targets `make up` (migración en arranque de la API) y `make generate` funcionen; el Makefile se mergea igualmente antes.
- La verificación del sha256 del release pinneado de yt-dlp requiere **una consulta puntual a internet** durante T-0068 para fijar `YTDLP_VERSION`/`YTDLP_SHA256`; quedan pinneados en el Dockerfile y los builds siguientes no dependen de nada más.
- El ensayo manual del rollback del RFC §8 y el test de idempotencia de `install.sh` en el servidor hogareño (SPECS-004.7) no son tareas de código; se realizan antes del primer release con migraciones.

## Nota de verificación
El repositorio es greenfield: `build/` y `deployments/` existen vacíos; `scripts/` y `Makefile` no existen. Las migraciones y fuentes sqlc ya están planificadas en EPIC-0001 (T-0003, T-0008..T-0010) y no se duplican aquí. Todos los archivos citados son `[NUEVO]`; no hubo nombres existentes que verificar. La numeración continúa la de EPIC-0003 (US-0022..US-0025, T-0066..T-0073).

## Historias de usuario (ordenadas por dependencias)
- [US-0022: Dockerfiles multi-stage de las 3 imágenes](./US-0022/description.md)
- [US-0023: `docker-compose.yml` y `.env.example` de 4 servicios](./US-0023/description.md)
- [US-0024: Scripts `install.sh` y `release.sh`](./US-0024/description.md)
- [US-0025: Makefile raíz del repo](./US-0025/description.md)
