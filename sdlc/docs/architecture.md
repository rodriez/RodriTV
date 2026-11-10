# Architecture — RodriTV

- Fecha de generación: 2026-11-10
- Alcance: repositorio `github.com/rodriez/RodriTV` (módulo `cmd/api`)
- Última actualización: se agrega `PostgresRepository` (T-0009, US-0002), implementación de `Repository` sobre las queries sqlc y `pgx`, con serialización JSONB, reloj inyectable y mapeo de errores de persistencia al dominio.

## System Context

```mermaid
C4Context
    title System Context — RodriTV

    Person(user, "Usuario IPTV", "Consume la playlist unificada")
    System(rodritv, "RodriTV", "Agregador de canales de TV que expone una playlist IPTV unificada")
    System_Ext(feeds, "Fuentes externas", "Feeds de canales de terceros")

    Rel(user, rodritv, "Consume playlist", "HTTP")
    Rel(rodritv, feeds, "Ingiere feeds", "HTTP")
```

| Elemento | Responsabilidad | Tecnología | Interacciones |
|---|---|---|---|
| RodriTV | Agregador de canales con playlist IPTV unificada | Go 1.27 | HTTP → fuentes externas (inferido del PRD/RFC) |

## Container

```mermaid
C4Container
    title Container — RodriTV

    Container(api, "cmd/api", "Go", "Binario de la API: configuración, ciclo de vida y endpoints HTTP")
    ContainerDb(db, "PostgreSQL", "PostgreSQL", "Base de datos relacional (acceso vía internal/platform)")
    System_Ext(feeds, "Fuentes externas", "Feeds de canales de terceros")

    Rel(api, db, "Pool de conexiones con retry", "pgx/PostgreSQL wire protocol")
    Rel(api, feeds, "Ingiere feeds", "HTTP (planificado)")
```

| Contenedor | Responsabilidad | Tecnología | Interacciones |
|---|---|---|---|
| cmd/api | Binario de la API (config, arranque, HTTP) | Go 1.27, Viper, pgx/v5 | pgx → PostgreSQL; HTTP → fuentes externas (planificado) |
| PostgreSQL | Persistencia relacional | PostgreSQL | ← pgxpool (MinConns=2, MaxConns=10) desde `internal/platform` |

## Component — cmd/api

```mermaid
C4Component
    title Component — cmd/api

    Component(config, "Config / LoadConfig", "Go, Viper", "Carga y valida la configuración desde variables de entorno RODRITV_API_*")
    Component(main, "main / run", "Go", "Composición y ciclo de vida: config -> logger -> pool -> migraciones -> server; devuelve exit code 0/1")
    Component(lifecycle, "runServer", "Go, os/signal", "Arranca el server en goroutine; signal.NotifyContext(SIGINT, SIGTERM) -> Shutdown drenando conexiones dentro de ShutdownTimeout")
    Component(logcfg, "newLogger / parseLogLevel", "Go, log/slog", "Logger JSON a stdout con el nivel de config (debug/info/warn/error; default info)")
    Component(server, "NewServer", "Go, net/http", "Arma el http.Server (ReadHeaderTimeout 5s) con middleware requestID -> logging -> recoverPanic")
    Component(routes, "registerRoutes / healthz", "Go, net/http ServeMux", "Registra GET /healthz: Ping a la DB -> 200 {\"status\":\"ok\"} o 503")
    Component(werr, "writeError / writeJSON", "Go, encoding/json", "Mapeo uniforme de errores: validation 400, not_found 404, conflict 409, resto 500")
    Component(dbpool, "NewPool (internal/platform)", "Go, pgx/v5 pgxpool", "Crea el pool de conexiones (MinConns=2, MaxConns=10) y espera al Ping con backoff exponencial (250ms x2, tope 5s, max 30s)")
    Component(migfs, "FS (migrations)", "Go, embed", "Migraciones SQL embebidas (000001_init up/down, SPECS-004.6)")
    Component(migrateup, "MigrateUp (internal/platform)", "Go, golang-migrate/v4", "Aplica las migraciones pendientes; ignora ErrNoChange")
    Component(domain, "Channel / Source (internal/channel)", "Go", "Dominio de canales: entidades con invariantes (Validate), NewChannel, errores ErrNotFound/ErrConflict/ErrValidation")
    Component(repo, "Repository (internal/channel)", "Go interface", "Contrato de persistencia definido por el dominio: List, GetByID, Create, Update, Delete, ExistsNumber; errores ErrNotFound/ErrConflict")
    Component(pgrepo, "PostgresRepository (internal/channel)", "Go, pgx/v5, encoding/json", "Implementa Repository sobre queries sqlc; serializa sources JSONB, genera UUID/timestamps y mapea errores PostgreSQL al dominio")
    Component(sqlc, "Queries / DBTX / ChannelRow (internal/channel)", "Go generado por sqlc, pgx/v5", "Queries CRUD y ExistsChannelNumber; modelos de canales, config, logs y uptime; sources como json.RawMessage")
    ContainerDb(pgdb, "PostgreSQL", "Base de datos relacional")

    Rel(pgrepo, repo, "Implementa el contrato de persistencia", "Go interface")
    Rel(pgrepo, sqlc, "Ejecuta operaciones CRUD y ExistsChannelNumber", "llamada directa")
    Rel(sqlc, pgdb, "Ejecuta queries de canales via DBTX/pgxpool", "pgx/PostgreSQL wire protocol")
    Rel(main, config, "Carga y valida config al arrancar (LoadConfig)", "llamada directa")
    Rel(main, logcfg, "Construye el logger slog JSON", "llamada directa")
    Rel(main, dbpool, "Crea el pool con retry (NewPool)", "llamada directa")
    Rel(main, migrateup, "Aplica migraciones (MigrateUp); si falla, exit 1", "llamada directa")
    Rel(main, server, "Construye el servidor (NewServer)", "llamada directa")
    Rel(main, lifecycle, "Delega arranque y shutdown (runServer)", "llamada directa")
    Rel(server, routes, "Registra rutas en el ServeMux", "llamada directa")
    Rel(routes, pgdb, "Ping de salud via interfaz Pinger", "pgx/PostgreSQL wire protocol")
    Rel(dbpool, pgdb, "Ping con retry hasta 30s", "pgx/PostgreSQL wire protocol")
    Rel(migrateup, migfs, "Lee los .sql embebidos", "iofs source")
    Rel(migrateup, pgdb, "Aplica migraciones DDL", "golang-migrate postgres driver")
```

| Componente | Responsabilidad | Tecnología | Interacciones |
|---|---|---|---|
| Config / LoadConfig (cmd/api/config.go) | Struct de configuración y carga con Viper: defaults (`PORT=9291`, `DATA_DIR=/data`, `LOG_LEVEL=info`, `SHUTDOWN_TIMEOUT=10s`), `AutomaticEnv` con prefijo `RODRITV_API`, validación fail-fast (`Validate`) | Go, spf13/viper | Variables de entorno del proceso |
| main / run (cmd/api/main.go) | Composición del binario: `LoadConfig` → logger slog JSON → `platform.NewPool` → `platform.MigrateUp` (error ≠ `ErrNoChange` → exit 1) → `NewServer` → `runServer`; inyección manual de dependencias, sin frameworks DI | Go stdlib, os/signal | orquesta todos los componentes; devuelve exit code 0/1 |
| runServer (cmd/api/main.go) | Arranca `ListenAndServe` en goroutine; `signal.NotifyContext(SIGINT, SIGTERM)` dispara `Shutdown` drenando conexiones en vuelo dentro de `ShutdownTimeout`; errores de arranque se propagan | Go, net/http, os/signal | señales del SO → graceful shutdown |
| newLogger / parseLogLevel (cmd/api/main.go) | Logger `slog` JSON a stdout con el nivel de `Config.LogLevel` (desconocido → info) | Go, log/slog | usado por toda la cadena de middleware y handlers |
| NewServer (cmd/api/server.go) | Construye el `http.Server` (ReadHeader 5s, Read 10s, Write 30s, Idle 120s) sobre el ServeMux envuelto en middleware: `requestID` (genera/propaga `X-Request-ID`), `logging` (slog JSON con `request_id`, `method`, `path`, `status`, `duration_ms`), `recoverPanic` (panic → 500 uniforme + stack en log) | Go, net/http, log/slog, google/uuid | compone `registerRoutes` |
| registerRoutes / healthz (cmd/api/routes.go) | Registra `GET /healthz`: ejecuta `Ping` (timeout 2s) contra la interfaz `Pinger` (la satisface `*pgxpool.Pool`); 200 `{"status":"ok"}` o 503 `{"status":"unavailable"}` | Go, net/http ServeMux | pgx Ping → PostgreSQL |
| writeError / writeJSON (cmd/api/server.go) | Mapeo uniforme de errores con `errors.Is/As`: validación → 400 `validation` (con `fields`), `not_found` → 404, `conflict` → 409, resto → 500 `internal` (sin filtrar el mensaje interno). JSON siempre `application/json; charset=utf-8` | Go, encoding/json | consumido por todos los handlers |
| NewPool (internal/platform/db.go) | Pool pgx con `MinConns=2`/`MaxConns=10`; `Ping` con backoff exponencial (inicio 250ms, ×2, tope 5s) hasta 30s; devuelve error si se agota el tiempo o se cancela el contexto | Go, jackc/pgx/v5 pgxpool, log/slog | pgx → PostgreSQL |
| FS (migrations/migrations.go) | `embed.FS` con las migraciones SQL (`000001_init.up.sql` / `000001_init.down.sql`, DDL de SPECS-004.6) | Go, embed | leído por `MigrateUp` vía iofs |
| MigrateUp (internal/platform/migrate.go) | Aplica todas las migraciones pendientes del FS embebido; ignora `migrate.ErrNoChange` (retorna nil), propaga el resto de errores | Go, golang-migrate/migrate/v4 (iofs + postgres) | DDL → PostgreSQL |
| Channel / Source (internal/channel/entity.go, source.go) | Dominio de canales: `Channel` (5 slots fijos de `*Source`, serialización JSON con nulls) y `Source` con invariantes en `Validate` (`Number`>0, `Name` 1–100, `SegmentSeconds` 2–10, `ScheduleURL` http(s) opcional, `Slot` 1–5 = posición, `URL` http(s)); `NewChannel` aplica defaults (habilitado, `SegmentSeconds=5`). Errores: sentinelas `ErrNotFound`/`ErrConflict` y tipado `ErrValidation{Field, Msg}` (con `ValidationFields()` para el mapeo HTTP) | Go, google/uuid | sin dependencias ni consumidores aún (hasta T-0010) |
| Repository (internal/channel/gateways.go) | Interfaz de salida del dominio hacia la persistencia, definida por el consumidor: `List` (ordenada por `number`), `GetByID`, `Create`, `Update`, `Delete`, `ExistsNumber` (excluye un ID para conflicto 409). Contratos de error documentados: `ErrNotFound` en GetByID/Update/Delete, `ErrConflict` en Create/Update | Go interface, google/uuid | implementada por `PostgresRepository` |
| PostgresRepository (internal/channel/postgres.go) | Repositorio PostgreSQL: UUID v4 y timestamps UTC con reloj inyectable; `sources` serializado como array JSONB de 5 posiciones; SQLSTATE 23505 de `number` → `ErrConflict`, no rows/0 filas → `ErrNotFound` | Go, pgx/v5, encoding/json | queries sqlc → PostgreSQL |
| Queries / DBTX / ChannelRow (internal/channel) | Código generado por sqlc desde `scripts/sqlc/channel.sql`: operaciones CRUD, `ExistsChannelNumber`, modelos de tablas y abstracción `DBTX` compatible con pgx; `ChannelRow` evita colisión con la entidad `Channel` | Go, sqlc v1.30.0, pgx/v5 | consumido por `PostgresRepository`; pgx → PostgreSQL |

## Notas / supuestos

- La interacción con fuentes externas proviene del PRD/RFC; aún no existe código de ingesta.
- Los errores de dominio (`ErrNotFound`, `ErrConflict`, `ErrValidation`) ya existen en `internal/channel/entity.go` (T-0006); `writeError` aún usa placeholders provisionales en `cmd/api/server.go` hasta que lleguen los handlers (US-0003).
