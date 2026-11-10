# SPECS-004: rodritv-infra — Imágenes, docker-compose, volúmenes y esquema de base de datos

- **RFC de origen:** RFC-001 — Motor de streaming con failover continuo para lista IPTV unificada (APPROVED, `.sdlc/define/RFC-001-streaming-engine-failover-iptv.md`)
- **Proyecto:** `rodritv-infra` (artefactos de `deployments/` y `build/`)
- **SPECS relacionadas:** SPECS-001 (rodritv-api), SPECS-002 (rodritv-playlist), SPECS-003 (rodritv-web)
- **Rol:** Tech Lead
- **Estado:** APPROVED
- **Fecha / Versión:** 2026-11-09 / v1.0

## 1. Objetivo técnico

Definir completamente: Dockerfiles multi-stage de los 3 binarios Go (con ffmpeg, yt-dlp pinneado y Chromium en la imagen de playlist), el `docker-compose.yml` de 4 servicios con healthchecks y orden de arranque, el script de instalación que crea `~/.rodritv/`, el versionado semántico de imágenes, y el **esquema inicial de PostgreSQL** (migraciones golang-migrate up/down) que comparten la API (SPECS-001) y el motor (SPECS-002).

## 2. Documentación interna consultada

| Documento | Ubicación | Qué aporta |
|---|---|---|
| Arquitectura de código | — (no existe) | Repo greenfield; scaffold con `deployments/` y `build/` vacíos |
| Contratos de APIs actuales | SPECS-001/002/003 | Variables de entorno y healthchecks que los servicios esperan |
| Convenciones y acuerdos | `AGENTS.md`, RFC-001 §7/§8 | Rollout y rollback definidos por el usuario |

## 3. Arquitectura de código

### Estructura de archivos

```
build/
  api/Dockerfile                    # multi-stage: build Go -> runtime alpine
  playlist/Dockerfile               # multi-stage: build Go -> runtime alpine + ffmpeg + yt-dlp + chromium
  web/Dockerfile                    # multi-stage: build Go -> runtime alpine
deployments/
  docker-compose.yml                # 4 servicios (db, api, playlist, web)
  .env.example                      # tags de imagen, credenciales de DB, timezone
scripts/
  install.sh                        # crea ~/.rodritv/{db,data,streaming,logs} + data/logos, copia compose y .env
  release.sh                        # build de las 3 imágenes con el mismo tag semántico
  sqlc/                             # fuentes .sql de las queries que compila sqlc (SPECS-001.3); el código .go se genera en cada paquete de dominio
migrations/                         # paquete raíz (project-layout), embebido en la API
  000001_init.up.sql / 000001_init.down.sql
  migrations.go                     # embed.FS
Makefile                            # comandos del repo: entorno compose, tests, build, generación sqlc
```

### Dockerfiles

- **Base de build común:** `golang:1.27-alpine`, `CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`, caché de capas con `go mod download` primero.
- **Runtime `api` y `web`:** `alpine:3.21` + `ca-certificates` + usuario no-root (`rodritv`, uid 1000); binario en `/app/`.
- **Runtime `playlist`:** `alpine:3.21` + `ca-certificates` + paquetes `ffmpeg` ( Alpine 3.21: ffmpeg 6.x) y `chromium` + **yt-dlp descargado como binario release pinneado** (`ARG YTDLP_VERSION=2025.xx.xx`, `ADD https://github.com/yt-dlp/yt-dlp/releases/download/.../yt-dlp /usr/local/bin/yt-dlp`, sha256 verificado con `ARG YTDLP_SHA256`). El binario queda fijado por imagen: actualizar yt-dlp = bump del `ARG` + rebuild; rollback de imagen = rollback de yt-dlp (RFC §4/§10).
- Cada imagen define `HEALTHCHECK` propio (`wget -qO- http://localhost:<port>/healthz || exit 1`) y `EXPOSE` del puerto.
- Tags: `rodritv-api:<x.y.z>`, `rodritv-playlist:<x.y.z>`, `rodritv-web:<x.y.z>` — las 3 imágenes con **el mismo tag por release** (RFC §7); `release.sh` toma el tag como argumento y lo valida contra `^\d+\.\d+\.\d+$`.

### docker-compose.yml (diseño)

- **Red:** `rodritv` (bridge interna). Solo `api` (9291), `playlist` (9290) y `web` (9292) publican puertos al host; `db` **no publica puerto**.
- **`db`:** `postgres:17-alpine`; env `POSTGRES_USER/PASSWORD/DB` desde `.env`; volumen `~/.rodritv/db:/var/lib/postgresql/data`; healthcheck `pg_isready -U $POSTGRES_USER`; `restart: always`; límite `mem_limit: 512m`.
- **`api`:** imagen `rodritv-api:${TAG}`; env `DATABASE_URL`, `DATA_DIR=/data`, `PORT=9291`; volumen `~/.rodritv/data:/data`; `depends_on: db (condition: service_healthy)`; `restart: always`.
- **`playlist`:** imagen `rodritv-playlist:${TAG}`; env `DATABASE_URL`, `STREAMING_DIR=/streaming`, `DATA_DIR=/data`, `LOGS_DIR=/logs`; volúmenes `~/.rodritv/streaming:/streaming`, `~/.rodritv/logs:/logs`, `~/.rodritv/data:/data:ro`; `depends_on: db (healthy)` **y `api (condition: service_started)`** con healthcheck en `api` — garantiza que el esquema está migrado antes de que el motor lea la DB (RFC §7); `restart: always`; `shm_size: 512m` (Chromium); `mem_limit: 2g`.
- **`web`:** imagen `rodritv-web:${TAG}`; env `API_URL=http://rodritv-api:9291`; `depends_on: api (condition: service_healthy)`; `restart: always`.
- Todos los servicios: `stop_grace_period` acorde a su `SHUTDOWN_TIMEOUT` (15–20 s).

### Script de instalación (`scripts/install.sh`)

1. Verifica dependencias del host (`docker`, `docker compose` v2).
2. Crea `~/.rodritv/{db,data,streaming,logs}` y `~/.rodritv/data/logos` con ownership del usuario actual.
3. Copia `docker-compose.yml` a `~/.rodritv/` y genera `.env` desde `.env.example` (contraseña de DB aleatoria si no existe).
4. `docker compose pull && docker compose up -d`; imprime URLs del panel/playlist.

Idempotente: re-ejecutar no destruye datos ni regenera la contraseña.

### Makefile (raíz del repo)

Punto de entrada único para desarrollo y CI. Los targets de compose usan `docker compose -f deployments/docker-compose.yml` (entorno de desarrollo; la instalación productiva usa la copia en `~/.rodritv/` vía `install.sh`).

| Target | Comando | Descripción |
|---|---|---|
| `make up` | `docker compose -f deployments/docker-compose.yml up -d --build` | Levanta el ambiente completo (db, api, playlist, web) construyendo las imágenes locales |
| `make down` | `... down` | Detiene y elimina los contenedores (conserva volúmenes de `~/.rodritv/`) |
| `make logs` | `... logs -f` | Sigue los logs de todos los servicios (`make logs S=api` filtra uno) |
| `make ps` | `... ps` | Estado y healthchecks de los servicios |
| `make build` | `go build ./...` | Compila los 3 binarios |
| `make test` | `go test ./... -coverprofile=cover.out` | Tests unitarios + integración (testcontainers requiere Docker corriendo); excluye código generado por sqlc del coverage (SPECS-001.7) |
| `make vet` | `go vet ./...` | Análisis estático |
| `make generate` | `sqlc generate` | Regenera el código `.go` desde `scripts/sqlc/` (requiere sqlc instalado; el código generado se commitea) |
| `make release TAG=x.y.z` | `scripts/release.sh $(TAG)` | Build y push de las 3 imágenes con el mismo tag semántico |
| `make clean` | `... down -v` + borra `cover.out` | Reset completo del entorno de desarrollo |

Todos los targets con `.PHONY`; `make help` (target por defecto) lista los comandos con su descripción.

## 4. Decisiones técnicas

| # | Decisión | Alternativas consideradas | Decisión final | Justificación |
|---|---|---|---|---|
| 1 | Base de imágenes runtime | distroless, alpine, debian-slim | **alpine:3.21** (versión pinneada) | Disponibilidad de paquetes `ffmpeg`/`chromium` en apk; imagen chica; versión fijada para builds reproducibles. |
| 2 | yt-dlp | pip en runtime, binario pinneado | **Binario release pinneado con sha256 verificado** | Decisión del RFC: arranque sin internet; rollback de imagen revierte yt-dlp. |
| 3 | Usuario en contenedores | root | **uid 1000 no-root** (`db` usa el usuario postgres de su imagen) | Buenas prácticas; además alinea permisos con `~/.rodritv/` del host (creado por el usuario 1000 típico). |
| 4 | Orden de arranque | solo `depends_on`, healthchecks | **`depends_on` + healthchecks**: db(healthy) → api(migra) → playlist(started)/web(healthy) | Decisión del RFC §7: el motor no lee la DB antes de las migraciones. |
| 5 | Puerto de PostgreSQL | publicado al host | **No publicado** (solo red interna compose) | Superficie mínima en la LAN; debugging con `docker compose exec db psql`. |
| 6 | Credenciales DB | hardcoded en compose, `.env` generado | **`.env` generado por install.sh con contraseña aleatoria**, `.env.example` commiteado sin secretos | No commitear secretos (AGENTS.md §Security). |
| 7 | Chromium sandbox | con sandbox, `--no-sandbox` | **`--no-sandbox`** dentro del contenedor | El sandbox de Chromium requiere privilegios/userns no disponibles por defecto en contenedores; el aislamiento lo da el propio contenedor. |
| 8 | Límites de recursos | sin límites | `mem_limit` por servicio (db 512m, playlist 2g, api/web 256m) | Hardware hogareño (RFC §6): acota fugas de memoria de Chromium/ffmpeg. |
| 9 | Timezone | UTC en contenedores | `TZ=America/Cordoba` vía `.env` en los 3 binarios | Logs y métricas en hora local del usuario. |

## 5. Contratos de API

No aplica (este proyecto no expone API). Los healthchecks que consumen compose son los `/healthz` definidos en SPECS-001.5.6, SPECS-002.5 y SPECS-003.5.

## 6. Persistencia: esquema inicial de PostgreSQL

Migración **`000001_init`** (archivos en el paquete raíz `migrations/`, embebidos en la API; sqlc los usa como fuente de esquema).

### `000001_init.up.sql`

```sql
-- Todos los ids son UUID v4 generados por la aplicación (Go, github.com/google/uuid);
-- la DB no genera ids (no DEFAULT gen_random_uuid()).
-- Todos los timestamps los genera la aplicación en Go (time.Now().UTC());
-- la DB no define DEFAULT now() (SPECS-001.4.7).
CREATE TABLE channels (
    id              UUID PRIMARY KEY,
    number          INTEGER NOT NULL UNIQUE CHECK (number > 0),
    name            TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
    logo            TEXT,
    schedule_url    TEXT,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    segment_seconds INTEGER NOT NULL DEFAULT 5 CHECK (segment_seconds BETWEEN 2 AND 10),
    sources         JSONB NOT NULL DEFAULT '[null,null,null,null,null]'::jsonb
                    CHECK (jsonb_array_length(sources) = 5),  -- array fijo de 5 slots; cada slot: objeto Source o null (SPECS-001.5.1)
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE config (
    key   TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
-- Semillas con los defaults de SPECS-001.5.3:
INSERT INTO config (key, value) VALUES
    ('neighbors', '2'),
    ('startup_channels', '3'),
    ('current_channel_ttl_minutes', '30'),
    ('buffer_minutes', '10'),
    ('uptime_threshold', '0.95');

CREATE TABLE uptime_samples (
    id          UUID PRIMARY KEY,
    channel_id  UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    source_id   UUID,                           -- id de la source embebida en channels.sources; NULL = muestra global del canal. Sin FK: es referencia lógica al JSON
    source_slot INTEGER CHECK (source_slot BETWEEN 1 AND 5),
    ts          TIMESTAMPTZ NOT NULL,   -- generado por la app (rodritv-playlist, SPECS-002)
    ok_count    INTEGER NOT NULL,   -- muestras buenas de la ventana de 30 s
    total_count INTEGER NOT NULL    -- muestras totales (6 con sample cada 5 s)
);
CREATE INDEX idx_uptime_channel_ts ON uptime_samples(channel_id, ts DESC);

CREATE TABLE process_logs (
    id         UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    source_id  UUID,                            -- referencia lógica a la source embebida (sin FK)
    ts         TIMESTAMPTZ NOT NULL,            -- generado por la app (rodritv-playlist, SPECS-002)
    process    TEXT NOT NULL,                  -- 'ffmpeg' | 'yt-dlp' | 'chromium' | 'mixer' | 'extractor'
    level      TEXT NOT NULL CHECK (level IN ('info', 'warn', 'error')),
    message    TEXT NOT NULL                   -- incluye comando, exit code y stderr relevante
);
CREATE INDEX idx_process_logs_channel_ts ON process_logs(channel_id, ts DESC);
```

### `000001_init.down.sql`

```sql
DROP TABLE IF EXISTS process_logs;
DROP TABLE IF EXISTS uptime_samples;
DROP TABLE IF EXISTS config;
DROP TABLE IF EXISTS channels;
```

Notas:
- **No hay tabla `sources`:** las sources son un array JSONB de 5 slots dentro de `channels` (decisión del usuario, SPECS-001.4.7b). La forma del array (exactamente 5 posiciones) se garantiza con `CHECK (jsonb_array_length(sources) = 5)`; la validez del contenido de cada slot se enforce en la capa de servicio (SPECS-001.5.2).
- **`source_id` sin FK:** al vivir las sources dentro del JSON del canal, `uptime_samples.source_id` y `process_logs.source_id` son referencias lógicas al UUID de la source embebida (más `source_slot` en métricas para reportes aunque se rote el id).
- **Retención de `uptime_samples` y `process_logs`:** 35 días (cubre la métrica de 30 días del RFC). Limpieza con un job nocturno *dentro del binario de la API* (`internal/maintenance`, `DELETE ... WHERE ts < now() - interval '35 days'` a las 04:00 local); sin extensiones de PostgreSQL (pg_cron no está en `postgres:17-alpine`).
- Volúmenes con datos fuera de las imágenes (RFC §8): el rollback solo cambia tags + `migrate down` previo cuando la migración lo requiera.

## 7. Estrategia de testing

- **Compose:** validación estática (`docker compose config -q`) en CI; smoke test de arranque completo en local/CI (`up -d`, esperar healthy, `curl` a los 3 `/healthz`, `down`).
- **Migraciones:** test up→down completo con testcontainers-go (ya cubierto en SPECS-001.7).
- **install.sh:** test manual en el servidor hogareño (idempotencia: doble ejecución); `shellcheck` limpio.
- **Imágenes:** verificación en build de que `ffmpeg -version`, `yt-dlp --version` y `chromium --version` existen (step de CI); sha256 de yt-dlp verificado en el Dockerfile.
- **Rollback:** ensayo manual del procedimiento del RFC §8 (cambiar tags, `migrate down`, `up -d`) antes del primer release con migraciones.

## 8. Supuestos

- El host tiene Docker Engine con plugin compose v2 y el usuario de instalación es uid 1000 (estándar en la primera cuenta de un servidor Linux); `install.sh` valida y advierte si difiere.
- Las versiones de paquetes apk (ffmpeg, chromium) quedan fijadas por la versión de alpine (3.21); bumps de seguridad se toman al rebuildar.
- No hay registry privado: las imágenes se construyen en el propio servidor con `release.sh` (`docker build` local). Si se adopta un registry, solo cambia el `push` — no afecta a las SPECS de servicio.

## 9. Riesgos y bloqueantes

Sin bloqueantes.

| Riesgo | Mitigación |
|---|---|
| Paquete `chromium` de alpine cambia/rompe headless | Versión de alpine pinneada; smoke test de build verifica `chromium --version`; fallback a `--headless=old` documentado |
| yt-dlp release URL/sha cambian al bump | ARGs centralizados al tope del Dockerfile; fallo de sha256 = build roto (detectable) |
| Volumen `db` con permisos incorrectos del host | install.sh crea el árbol con el usuario actual (uid 1000) antes del primer `up` |
| Crecimiento de tablas de métricas/logs | Retención de 35 días con job nocturno en la API (SPECS-004.6) |
| Rollback con migraciones incompatibles | Pares up/down obligatorios y testeados (SPECS-004.7); procedimiento de rollback incluye `migrate down` |

## 10. Tabla de proyectos afectados

| Proyecto | Tipo | Cambios de la SPECS que aplica | Secciones |
|---|---|---|---|
| rodritv-infra | nuevo | compose, Dockerfiles, scripts, .env.example | 3–4 |
| rodritv-api | nuevo | Migra y mantiene el esquema (migraciones embebidas, job de retención) | 6 |
| rodritv-playlist | nuevo | Consume el esquema; su imagen incluye ffmpeg/yt-dlp/Chromium | 3, 6 |
| rodritv-web | nuevo | Imagen y servicio compose | 3 |

## 11. Fuera de alcance

- Registry de imágenes, CI/CD remoto, firma de imágenes.
- Backups automatizados de PostgreSQL (hogareño, un solo usuario; evaluable en v2).
- Monitoreo/alertas externas (Prometheus, etc.).
- HTTPS/TLS (tráfico solo LAN, decisión del PRD).
