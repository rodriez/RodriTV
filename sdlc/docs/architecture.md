# Architecture — RodriTV

- Fecha de generación: 2026-11-10
- Alcance: repositorio `github.com/rodriez/RodriTV` (módulo `cmd/api`)
- Última actualización: se agrega el paquete `migrations` (SQL `000001_init` embebido vía `FS`) y `MigrateUp` en `internal/platform` (runner golang-migrate iofs+postgres, T-0003, US-0001).

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
    Component(main, "main", "Go", "Punto de entrada (stub, pendiente de composición en T-0005)")
    Component(dbpool, "NewPool (internal/platform)", "Go, pgx/v5 pgxpool", "Crea el pool de conexiones (MinConns=2, MaxConns=10) y espera al Ping con backoff exponencial (250ms x2, tope 5s, max 30s)")
    Component(migfs, "FS (migrations)", "Go, embed", "Migraciones SQL embebidas (000001_init up/down, SPECS-004.6)")
    Component(migrateup, "MigrateUp (internal/platform)", "Go, golang-migrate/v4", "Aplica las migraciones pendientes; ignora ErrNoChange")
    ContainerDb(pgdb, "PostgreSQL", "Base de datos relacional")

    Rel(main, config, "Cargará y validará config al arrancar", "llamada directa (planificado)")
    Rel(dbpool, pgdb, "Ping con retry hasta 30s", "pgx/PostgreSQL wire protocol")
    Rel(migrateup, migfs, "Lee los .sql embebidos", "iofs source")
    Rel(migrateup, pgdb, "Aplica migraciones DDL", "golang-migrate postgres driver")
```

| Componente | Responsabilidad | Tecnología | Interacciones |
|---|---|---|---|
| Config / LoadConfig (cmd/api/config.go) | Struct de configuración y carga con Viper: defaults (`PORT=9291`, `DATA_DIR=/data`, `LOG_LEVEL=info`, `SHUTDOWN_TIMEOUT=10s`), `AutomaticEnv` con prefijo `RODRITV_API`, validación fail-fast (`Validate`) | Go, spf13/viper | Variables de entorno del proceso |
| main (cmd/api/main.go) | Punto de entrada del binario; stub vacío hasta T-0005 | Go stdlib | — |
| NewPool (internal/platform/db.go) | Pool pgx con `MinConns=2`/`MaxConns=10`; `Ping` con backoff exponencial (inicio 250ms, ×2, tope 5s) hasta 30s; devuelve error si se agota el tiempo o se cancela el contexto | Go, jackc/pgx/v5 pgxpool, log/slog | pgx → PostgreSQL |
| FS (migrations/migrations.go) | `embed.FS` con las migraciones SQL (`000001_init.up.sql` / `000001_init.down.sql`, DDL de SPECS-004.6) | Go, embed | leído por `MigrateUp` vía iofs |
| MigrateUp (internal/platform/migrate.go) | Aplica todas las migraciones pendientes del FS embebido; ignora `migrate.ErrNoChange` (retorna nil), propaga el resto de errores | Go, golang-migrate/migrate/v4 (iofs + postgres) | DDL → PostgreSQL |

## Notas / supuestos

- La interacción con fuentes externas proviene del PRD/RFC; aún no existe código de ingesta.
- El servidor HTTP y `/healthz` llegan en T-0004; la composición del ciclo de vida en T-0005.
