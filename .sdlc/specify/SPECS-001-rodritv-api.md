# SPECS-001: rodritv-api — API REST de gestión

- **RFC de origen:** RFC-001 — Motor de streaming con failover continuo para lista IPTV unificada (APPROVED, `.sdlc/define/RFC-001-streaming-engine-failover-iptv.md`)
- **Proyecto:** `rodritv-api` (binario `cmd/api`, imagen `rodritv-api`, puerto **9291**)
- **SPECS relacionadas:** SPECS-002 (rodritv-playlist), SPECS-003 (rodritv-web), SPECS-004 (rodritv-infra)
- **Rol:** Tech Lead
- **Estado:** APPROVED
- **Fecha / Versión:** 2026-11-09 / v1.0

## 1. Objetivo técnico

Servicio de gestión puro (sin streaming): API REST JSON sobre PostgreSQL para CRUD de canales y sources, configuración de comportamiento, y consulta de métricas de uptime y logs de procesos. Ejecuta las migraciones de base de datos (golang-migrate) al iniciar, antes de aceptar tráfico. Consumido únicamente por el backend del panel (`rodritv-web`) dentro de la red de docker-compose; sin autenticación (producto hogareño LAN, decisión del PRD).

## 2. Documentación interna consultada

| Documento | Ubicación | Qué aporta |
|---|---|---|
| Arquitectura de código | — (no existe) | Repo greenfield; se asume layout estándar Go (ver SPECS-001.8 Supuestos) |
| Contratos de APIs actuales | — (no existe) | No hay API previa; esta SPECS define el contrato inicial |
| Convenciones y acuerdos | `AGENTS.md` | Stack Go 1.27.3, módulo `github.com/rodriez/RodriTV`, gofmt/vet, tests con `testing` |

## 3. Arquitectura de código

### Estructura de paquetes

Arquitectura **screaming** (la estructura grita el dominio) según `golang-standards/project-layout`, con **máximo 2 niveles de profundidad** dentro de `internal/`. Un paquete por dominio; la implementación de persistencia convive en el mismo paquete del dominio:

```
cmd/api/main.go                     # composición e inyección de dependencias manual
cmd/api/routes.go                   # registro de rutas del ServeMux (patterns Go 1.27+: "GET /v1/channels/{id}", …)
cmd/api/config.go                   # Viper: env vars, defaults, validación (arranque)
cmd/api/server.go                   # construcción del http.Server, middleware (logging, recover), ciclo de vida, mapeo uniforme de errores
internal/channel/                   # dominio canales
  entity.go                         # Channel, value objects, errores de dominio
  service.go                        # casos de uso (validaciones, orquestación)
  gateways.go                       # interfaz Repository (la define el dominio/consumidor)
  handler.go                        # transporte HTTP; llama solo al servicio
  postgres.go                       # implementación Repository (usa el código generado por sqlc)
  queries.sql.go, models.go         # generados por sqlc en el paquete del dominio (commiteados)
  source.go                         # entidad Source embebida (slot 1–5, url, enabled) y sus validaciones de dominio
  service_test.go, handler_test.go, postgres_test.go
internal/setting/                   # dominio configuración de comportamiento (KV tipado)
internal/metric/                    # dominio métricas de uptime (consulta, solo lectura)
internal/proclog/                   # dominio consulta de logs de procesos (solo lectura)
internal/platform/
  db.go                             # pgxpool, ping con retry
  migrate.go                        # golang-migrate con iofs sobre el embed.FS del paquete migrations
migrations/                         # paquete raíz (project-layout): 000001_*.up.sql / .down.sql (SPECS-004.6) + migrations.go con embed.FS
scripts/sqlc/                       # fuentes .sql de las queries que compila sqlc (una por dominio: channel.sql, setting.sql, …)
```

### Patrones y buenas prácticas

Se aplican las reglas no negociables de la skill `go-backend`:

- **Dominio primero:** se modela primero cada dominio (`entity.go`: entidades, value objects, errores de dominio) sin dependencias de infraestructura ni frameworks; luego contratos, servicios y, al final, transporte y persistencia.
- **Capas sin saltos:** el flujo es estrictamente `handler → service → domain → repository`. Cada capa solo conoce a la inmediatamente adyacente; los handlers jamás tocan el repositorio.
- **Comunicación por interfaces definidas por el consumidor:** cada dominio declara en `gateways.go` la interfaz `Repository` que necesita; `postgres.go` la implementa. *Accept interfaces, return structs*: las funciones retornan tipos concretos, nunca interfaces.
- **Sin punteros innecesarios:** se retornan valores salvo que haya opcionalidad semántica, mutación intencional o estructuras grandes.
- **Inyección de dependencias manual en `cmd/api/main.go`** (`NewService(repo)`, `NewHandler(svc)`); sin frameworks de DI. El main conoce todas las capas; ninguna capa conoce al main.
- **Standard library primero:** ServeMux, `encoding/json`, `context`, `slog`. Toda dependencia externa (pgx, sqlc, Viper, golang-migrate, testify) está justificada en SPECS-001.4.
- **Rutas junto al entrypoint:** el registro de rutas del ServeMux vive en `routes.go` dentro de la carpeta del `cmd` de cada servicio (`cmd/api/routes.go`, `cmd/playlist/routes.go`, `cmd/web/routes.go`); los paquetes `internal/` exponen handlers pero no conocen el árbol de rutas. Convención de todos los binarios del repo (SPECS-001, SPECS-002, SPECS-003).
- **Errores:** errores de dominio sentinelas/tipados (`ErrNotFound`, `ErrValidation{Field, Msg}`, `ErrConflict`); `errors.Is/As` en `cmd/api/server.go` para el mapeo HTTP uniforme. Wrapping con `fmt.Errorf("...: %w", err)`.
- **Context en toda la cadena:** `context.Context` como primer parámetro en servicios, repositorios y queries sqlc; timeouts con middleware (`http.Timeout`) y `context.WithTimeout` en handlers pesados.
- **Config junto al entrypoint:** la config de arranque vive en `config.go` dentro de la carpeta del `cmd` de cada servicio (`cmd/api/config.go`, `cmd/playlist/config.go`, `cmd/web/config.go`), no en `internal/` — es un detalle del binario, no del dominio. Viper con `AutomaticEnv`, prefijo `RODRITV_API`, bind explícito por clave, defaults declarados y `Validate()` al arrancar (fail fast). Convención de todos los binarios del repo (SPECS-001, SPECS-002, SPECS-003).
- **Logging:** `slog` con handler JSON; logger inyectado, nunca el default global en librerías internas. Campos estructurados: `request_id`, `method`, `path`, `status`, `duration_ms`.
- **sqlc:** las queries fuente `.sql` viven centralizadas en **`scripts/sqlc/`** (una por dominio) y sqlc genera los paquetes `.go` resultantes **dentro del paquete del dominio que corresponda** (`internal/<dominio>/`), commiteados (builds sin sqlc instalado); `sqlc.yaml` en la raíz del repo mapea cada archivo de `scripts/sqlc/` a su paquete destino; esquema desde `migrations/`.
- **Migraciones:** `golang-migrate` con driver `iofs` sobre el `embed.FS` del paquete raíz `migrations/`. Se ejecutan en `main` antes de levantar el server; si `migrate.Up()` falla (distinto de `ErrNoChange`), el proceso sale con código != 0 (compose lo reintenta por `restart: always`).

## 4. Decisiones técnicas

| # | Decisión | Alternativas consideradas | Decisión final | Justificación |
|---|---|---|---|---|
| 1 | Router HTTP | net/http ServeMux (Go 1.27+ patterns), chi v5, gorilla/mux | **`net/http` ServeMux estándar** con patterns de método y wildcard (`"GET /channels/{id}"`) | Decisión del RFC: la API es un CRUD simple; los patterns nativos de Go 1.27+ cubren métodos y path params sin dependencias externas. Middleware por composición de `http.Handler`. |
| 2 | Driver PostgreSQL | pgx/v5, lib/pq, database/sql | **pgx/v5** (`github.com/jackc/pgx/v5`, v5.x) con `pgxpool` | Decisión del usuario. Driver estándar moderno, pool nativo, tipos PostgreSQL completos. |
| 3 | Generación de acceso a datos | sqlc, queries manuales, GORM | **sqlc** (v1.x; queries fuente en `scripts/sqlc/`, código `.go` generado en el paquete del dominio correspondiente) | Decisión del usuario. Código tipado desde SQL, sin ORM, ideal para consultas de métricas. Código generado commiteado. |
| 4 | Migraciones | golang-migrate, goose, migraciones manuales | **golang-migrate v4** (`github.com/golang-migrate/migrate/v4`) con `iofs` + `embed.FS` | Decisión del RFC. Pares up/down embebidos en el binario; rollback con `down` (ver RFC §8). |
| 5 | Config de arranque | caarlos0/env, os.Getenv manual | **Viper** (`github.com/spf13/viper`, v1.x) | Decisión del usuario. Env + defaults + validación centralizada. |
| 6 | Logging | zerolog, zap | **log/slog** (stdlib) | Decisión del usuario. Cero dependencias, JSON estructurado. |
| 7 | IDs y timestamps | generados por la DB (serial, `DEFAULT now()`), generados por la aplicación | **UUID v4 y timestamps generados por Go** (`github.com/google/uuid`, `time.Now().UTC()`); la DB no genera ids ni fechas (sin `DEFAULT now()`) | Alineado con el esquema de SPECS-004.6: las sources embebidas en JSONB necesitan ids estables generados por la app, y los timestamps en Go permiten testear la lógica con relojes controlados y evitan depender del reloj de la DB. |
| 8 | Validación de requests | go-playground/validator, manual | **Manual en servicios** (`ErrValidation`) | Pocos recursos y reglas de dominio (array de 5 slots de sources, unicidad de `number`) que no expresa bien un validador de tags. |
| 9 | Respuesta de error | por handler, formato uniforme | **Formato uniforme** `{"error": {"code": "...", "message": "...", "fields": {...}?}}` | Un solo mapeo en `cmd/api/server.go`; el panel lo consume de forma predecible. |
| 10 | Versionado de API | `/v1/`, sin versión, header de versión | **Prefijo `/v1` en el path** para todos los recursos | Decisión del usuario: versionado explícito desde el primer release; permite evolucionar el contrato (`/v2`) sin romper clientes. `/healthz` queda fuera del versionado (endpoint operativo, no de negocio). |
| 11 | Tests de repositorio | pgxmock, testcontainers-go | **testcontainers-go** (postgres:17-alpine) para repos; fakes de repo para servicios | Queries sqlc se validan contra PostgreSQL real (mismo engine que producción); servicios se testean con fakes rápidos. |

## 5. Contratos de API

Base: `http://rodritv-api:9291`. **Todos los recursos bajo el prefijo `/v1`** (en las tablas siguientes las rutas se listan sin el prefijo por legibilidad: leer `/channels` como `/v1/channels`, `/config` como `/v1/config`, etc.). Única excepción: `/healthz` (SPECS-001.5.6), sin versionar. Todas las respuestas `application/json; charset=utf-8`. Errores siempre:

```json
{
  "error": {
    "code": "validation",
    "message": "descripción legible",
    "fields": {
      "name": "es requerido"
    }
  }
}
```

Códigos: `validation` (400), `not_found` (404), `conflict` (409), `internal` (500).

**Serialización:** los campos con valor `null` no aparecen en los payloads (`omitempty` en todos los DTOs; un campo opcional ausente es indistinguible de uno nulo, y el panel lo trata como "sin valor").

### 5.1 Canales

**Modelo `Channel`:**

```json
{
  "id": "3f6b2c1e-...",
  "number": 1,
  "name": "El Trece",
  "logo": "3f6b2c1e-....png",
  "schedule_url": "https://...",
  "enabled": true,
  "segment_seconds": 5,
  "sources": [
    {
      "id": "9a1c...",
      "slot": 1,
      "url": "https://.../index.m3u8",
      "enabled": true,
      "created_at": "2026-11-09T12:00:00Z",
      "updated_at": "2026-11-09T12:00:00Z"
    },
    null,
    null,
    null,
    null
  ],
  "created_at": "2026-11-09T12:00:00Z",
  "updated_at": "2026-11-09T12:00:00Z"
}
```

- `number`: entero positivo, único, define orden en la playlist M3U (requerido).
- `name`: 1–100 chars (requerido). `logo`: nombre de archivo dentro de `~/.rodritv/data/logos` (opcional), con la convención **`<channel_id>.<ext>`** — `<channel_id>` es el UUID del campo `id` del canal y `<ext>` la extensión original del archivo subido (ej. `3f6b2c1e-....png`). Al cambiar el logo de un canal se pisa el archivo anterior del mismo canal (se elimina si cambió la extensión); al eliminar el canal se elimina su logo. `schedule_url`: URL http(s) opcional. `segment_seconds`: 2–10, default 5.
- `sources`: array fijo de **5 slots** (modelo y reglas en 5.2); cada posición es un objeto `Source` o `null`. El array se persiste como JSONB dentro de `channels` (SPECS-004.6) — no existe tabla `sources`.

| Método | Ruta | Descripción | Request | Response |
|---|---|---|---|---|
| GET | `/channels` | Lista ordenada por `number` | — | `200 { "channels": [Channel] }` |
| POST | `/channels` | Crear | Channel (sin id/timestamps) | `201 Channel` · `400/409` (number duplicado) |
| GET | `/channels/{id}` | Obtener | — | `200 Channel` · `404` |
| PUT | `/channels/{id}` | Reemplazo total | Channel | `200 Channel` · `400/404/409` |
| DELETE | `/channels/{id}` | Eliminar (con sus sources) | — | `204` · `404` |

### 5.2 Sources (embebidas en el canal)

No hay endpoints propios de sources: viven dentro del array `sources` del canal (5.1) y se crean/editan/eliminan con `POST /channels` y `PUT /channels/{id}` (reemplazo total del array).

**Modelo `Source`:**

```json
{
  "id": "9a1c4e2d-...",
  "slot": 1,
  "url": "https://.../index.m3u8",
  "enabled": true,
  "created_at": "2026-11-09T12:00:00Z",
  "updated_at": "2026-11-09T12:00:00Z"
}
```

- `id`: UUID v4 generado por la aplicación al crear la source (estable mientras la source exista; referenciado por `uptime_samples.source_id` y `process_logs.source_id`, SPECS-004.6).
- `slot`: posición 1–5 dentro del array fijo de 5. El array siempre se serializa con las 5 posiciones (los slots vacíos como `null`). La **source principal es la del slot ocupado más bajo** (normalmente slot 1); el resto actúa como backups por orden de slot, que el motor enciende según la regla del 95% (SPECS-002). No existe campo `priority` ni `is_backup`: el orden de failover y el rol principal/backup se derivan del `slot`.
- `url`: URL http(s) de la source (requerida en slots ocupados). **No hay campo `type`:** el sistema detecta automáticamente el tipo de source (m3u8 / YouTube / web) a partir de la URL al encenderla (lógica del motor, SPECS-002).
- Validación en la capa de servicio: array de exactamente 5 posiciones; slots ocupados con `url` válida y `slot` igual a su posición (1-based) en el array.

### 5.3 Configuración de comportamiento

Claves (tabla `config`, KV tipado — esquema en SPECS-004.6):

| Clave | Tipo | Default | Rango | Descripción |
|---|---|---|---|---|
| `neighbors` | int | 2 | 0–5 | Canales vecinos encendidos hacia atrás/adelante |
| `startup_channels` | int | 3 | 0–20 | Canales encendidos al iniciar el servidor |
| `current_channel_ttl_minutes` | int | 30 | 5–120 | TTL del "canal actual" desde su última request de segmentos |
| `buffer_minutes` | int | 10 | 1–10 | Buffer máximo por canal en disco |
| `uptime_threshold` | float | 0.95 | 0.5–1.0 | Umbral de la regla del 95% |

| Método | Ruta | Descripción | Response |
|---|---|---|---|
| GET | `/config` | Config efectiva completa | `200` (abajo) |
| PUT | `/config` | Reemplazo total (valida rangos) | `200` config actualizada · `400` |

```json
{
  "neighbors": 2,
  "startup_channels": 3,
  "current_channel_ttl_minutes": 30,
  "buffer_minutes": 10,
  "uptime_threshold": 0.95
}
```

### 5.4 Métricas

| Método | Ruta | Descripción | Response |
|---|---|---|---|
| GET | `/channels/{id}/metrics?window=24h` | Uptime del canal y por source | `200` (abajo) · `400` (window inválida) · `404` |

`window`: `1h | 24h | 7d | 30d` (default `24h`).

```json
{
  "channel_id": "3f6b2c1e-...",
  "window": "24h",
  "uptime": 0.987,
  "sources": [
    {
      "source_id": "9a1c4e2d-...",
      "slot": 1,
      "uptime": 0.99,
      "active": true,
      "state": "running"
    },
    {
      "source_id": "b77f0a3c-...",
      "slot": 2,
      "uptime": 0.0,
      "active": false,
      "state": "standby"
    }
  ],
  "interruptions": 2
}
```

`uptime` se calcula sobre `uptime_samples` (muestras cada 30 s por source/canal — escritas por rodritv-playlist, SPECS-002). `interruptions`: cantidad de transiciones a "sin señal sana" en la ventana.

### 5.5 Logs

| Método | Ruta | Descripción | Response |
|---|---|---|---|
| GET | `/channels/{id}/logs?limit=100&offset=0&level=error` | Logs de procesos del canal (más recientes primero) | `200` (abajo) |

```json
{
  "logs": [
    {
      "ts": "2026-11-09T12:00:00Z",
      "source_id": "9a1c4e2d-...",
      "process": "ffmpeg",
      "level": "error",
      "message": "..."
    }
  ],
  "total": 42
}
```

`level`: `all | info | warn | error` (default `all`). `limit` 1–500 (default 100).

### 5.6 Health

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/healthz` | `200` si la DB responde `SELECT 1`; `503` si no. Usado por el healthcheck del compose (SPECS-004). |

```json
{
  "status": "ok"
}
```

## 6. Persistencia, integraciones y proveedores

- **Base de datos:** PostgreSQL 17 (`postgres:17-alpine`, contenedor `rodritv-db`, SPECS-004). Esquema completo y migraciones iniciales definidos en **SPECS-004.6**; esta SPECS consume las tablas `channels` (con las sources embebidas en JSONB), `config`, `uptime_samples`, `process_logs`.
- **Conexión:** `pgxpool` con `min_conns=2`, `max_conns=10`; `DATABASE_URL` desde env (construida por compose). Ping con retry exponencial (máx. 30 s) antes de migrar.
- **IDs y timestamps:** los genera la aplicación en Go (decisión 7). `created_at`/`updated_at` se asignan en la capa de persistencia con `time.Now().UTC()` al insertar/actualizar (el servicio puede inyectar un reloj para tests); el esquema no define `DEFAULT now()` (SPECS-004.6 debe reflejarlo).
- **Logos:** los archivos se guardan en `~/.rodritv/data/logos` (montado en el contenedor) con el nombre **`<channel_id>.<ext>`** (UUID del canal + extensión original del archivo subido; el `channel_id` es conocido porque el alta del logo es posterior a la del canal, y en la creación conjunta el UUID se genera en la app antes de persistir — decisión 7). `GET /channels` devuelve el nombre de logo y **`GET /logos/{filename}`** sirve los archivos para que el panel los muestre. Path traversal rechazado (solo `[a-zA-Z0-9._-]`); extensiones permitidas: `png`, `jpg`, `jpeg`, `webp`, `svg`.

### Variables de entorno (Viper, prefijo `RODRITV_API_`)

| Var | Default | Descripción |
|---|---|---|
| `PORT` | `9291` | Puerto HTTP |
| `DATABASE_URL` | — (requerida) | DSN PostgreSQL |
| `DATA_DIR` | `/data` | Raíz montada de `~/.rodritv/data` (logos en `${DATA_DIR}/logos`) |
| `LOG_LEVEL` | `info` | `debug|info|warn|error` |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown |

## 7. Estrategia de testing

**TDD obligatorio** (skill `go-backend`): para cada caso de uso se escribe primero el test que falla, luego la implementación mínima, luego refactor (red → green → refactor). Assertions con `testify/assert` + `testify/require`.

- **Caja negra:** cada componente se testea por su comportamiento observable vía su interfaz pública, no por detalles internos.
- **Mocks solo en los extremos** con `testify/mock`: el único extremo de esta SPECS es la persistencia (`Repository` de `gateways.go`). No se mockean colaboradores internos.
- **Servicios (`channel`, `setting`, `metric`, `proclog`):** unit tests con el repositorio mockeado (`testify/mock`). Casos: validaciones, array de 5 slots de sources (slots ocupados con URL válida y `slot` coherente con su posición), conflictos de `number`, rangos de config, cálculo de uptime con samples sintéticos.
- **Handlers:** `httptest` con el **servicio real** y el repositorio mockeado (el extremo sigue siendo la persistencia); se verifica status code, body JSON y formato de error uniforme por cada ruta y código de error.
- **`postgres.go` (implementación del Repository):** tests de integración con `testcontainers-go` (postgres:17-alpine) aplicando las migraciones reales; verifican cada método del repositorio (lo que cubre indirectamente las queries). **El código generado por sqlc (`queries.sql.go`, `models.go`) no se testea directamente**: es código generado, sin lógica propia.
- **Migraciones:** test que aplica `up` completo y luego `down` completo sobre contenedor efímero (garantiza reversibilidad, requisito del plan de rollback del RFC).
- **Cobertura objetivo:** ≥ 80% en la lógica de negocio (servicios de `channel`, `setting`, `metric`); handlers ≥ 70%. **El código generado por sqlc se excluye del coverage:** los archivos generados llevan el encabezado estándar `// Code generated by sqlc. DO NOT EDIT.` y se filtran en el cálculo de cobertura (herramientas como `gocov`/`go-test-coverage` o post-procesado del `coverprofile` descartando archivos marcados como generados); de lo contrario inflarían artificialmente la métrica.

## 8. Supuestos

- No existe documentación interna previa (greenfield); se asumen convenciones estándar Go (gofmt, go vet limpio, layout `cmd/` + `internal/`).
- Las versiones exactas de librerías se resuelven con `go mod tidy` en Implement; las indicadas son major versions objetivo.
- La escritura de `uptime_samples` y `process_logs` es responsabilidad exclusiva de rodritv-playlist (SPECS-002); la API es de solo lectura sobre esas tablas.

## 9. Riesgos y bloqueantes

Sin bloqueantes. Riesgos heredados del RFC que tocan a la API: ninguno directo (los riesgos de streaming son de SPECS-002). Riesgo menor: contención de migraciones si se escala a múltiples réplicas — no aplica (una sola réplica, RFC §7).

## 10. Tabla de proyectos afectados

| Proyecto | Tipo | Cambios de la SPECS que aplica | Secciones |
|---|---|---|---|
| rodritv-api | nuevo | Todo el servicio: paquetes, contratos, migraciones al arranque | 3–6 |
| rodritv-infra | nuevo | Imagen `rodritv-api`, servicio compose, volumen `data`, healthcheck | 5.6, 6 (detalle en SPECS-004) |

## 11. Fuera de alcance

- Cualquier lógica de streaming, encendido de canales o failover (SPECS-002).
- Autenticación/autorización (descartado por el PRD; documentar "solo LAN").
- Servir la playlist M3U o segmentos HLS (puerto 9290, SPECS-002).
- UI del panel (SPECS-003).
- Endpoints de escritura de métricas/logs (las escribe el motor directamente a la DB).
