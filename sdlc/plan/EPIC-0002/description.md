# EPIC-0002: Motor de streaming con failover continuo + servidor M3U/HLS (`rodritv-playlist`)

- SPECS de origen: SPECS-002 (`sdlc/specify/SPECS-002-rodritv-playlist.md`)
- RFC de origen: RFC-001 (`sdlc/define/RFC-001-streaming-engine-failover-iptv.md`)
- Roles: Tech Lead + Software Engineer
- Fecha: 2026-11-10
- Proyectos involucrados: `rodritv-playlist` (binario `cmd/playlist`, puerto 9290)

## Objetivo técnico
Al completar todas las US queda construido el binario `cmd/playlist` que: lee canales/sources/config de comportamiento desde PostgreSQL con aplicación en caliente (polling 15 s); mantiene encendido el canal actual + N vecinos vía `PowerScheduler`; por canal levanta un `ChannelSupervisor` con un `SourceProcess` ffmpeg por source activa (extractores m3u8/youtube/web, backoff, regla del 95% con histéresis); compone la salida HLS continua del canal con `ChannelMixer` (MEDIA-SEQUENCE y PROGRAM-DATE-TIME monótonos, EXT-X-DISCONTINUITY en failover, slate de "problemas técnicos", retención de buffer); persiste `uptime_samples` y `process_logs` en PostgreSQL y logs crudos en `~/.rodritv/logs`; y expone en `:9290` `/playlist.m3u`, `/channels/{n}/index.m3u8`, `/channels/{n}/seg-{seq}.ts`, `/logos/{filename}` y `/healthz`.

## Alcance
- Entra: plataforma base del binario (config Viper, pool pgx, servidor HTTP, graceful shutdown), dominios `catalog` (lectura + watch), `telemetry` + `proclog` (escritura), `extractor`, `supervisor` (SourceProcess + uptime), `mixer` (incluye spike PoC del gate del RFC), `engine` (PowerScheduler), `streamserver` y el `main` de composición.
- NO entra: migraciones de DB ni CRUD de canales/config (SPECS-001, EPIC-0001), UI del panel (SPECS-003), Dockerfile/compose/binarios pinneados (SPECS-004), autenticación en :9290, EPG.

## Dependencias externas (fuera del Plan)
Ninguna gestión no-codeable detectada. Advertencias técnicas:
- **El esquema de DB lo crea `rodritv-api`** (migración `000001_init` embebida en el binario `cmd/api`, EPIC-0001 / SPECS-004.6). `rodritv-playlist` NO migra: asume el esquema aplicado (compose ordena db → api → playlist). Los tests de integración de `catalog`/`telemetry` (testcontainers) aplican los mismos SQL de migración.
- La PoC del mixer (T-0037) requiere **2 URLs de sources reales** del set del RFC y ffmpeg/ffprobe instalados en el entorno de desarrollo; es prueba de aceptación manual (gate del RFC riesgo n.º 1) antes de cerrar US-0012.
- Los binarios `ffmpeg`, `ffprobe`, `yt-dlp`, `chromium` deben existir en `$PATH` del entorno de desarrollo para los tests de integración (en producción los provee la imagen, SPECS-004).

## Nota de verificación
El repositorio es greenfield (sin código Go). Todos los componentes citados son `[NUEVO]` y siguen la estructura de SPECS-002.3; no hubo nombres existentes que verificar. La numeración continúa la de EPIC-0001 (US-0007..US-0015, T-0020..T-0048).

## Historias de usuario (ordenadas por dependencias)
- [US-0007: Plataforma base del binario `cmd/playlist`](./US-0007/description.md)
- [US-0008: Dominio catalog: lectura de canales/config y watch en caliente](./US-0008/description.md)
- [US-0009: Dominios telemetry y proclog: escritura de métricas y logs](./US-0009/description.md)
- [US-0010: Extractores de stream (m3u8, YouTube, web headless)](./US-0010/description.md)
- [US-0011: SourceProcess: ingesta ffmpeg supervisada y sondeo de perfil](./US-0011/description.md)
- [US-0012: ChannelMixer: salida HLS continua, failover y slate](./US-0012/description.md)
- [US-0013: ChannelSupervisor y regla del 95%](./US-0013/description.md)
- [US-0014: PowerScheduler y Engine: encendido acotado y config en caliente](./US-0014/description.md)
- [US-0015: StreamServer: playlist M3U, HLS por canal, logos y healthz](./US-0015/description.md)
