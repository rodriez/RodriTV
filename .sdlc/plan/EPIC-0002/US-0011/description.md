# US-0011: SourceProcess: ingesta ffmpeg supervisada y sondeo de perfil
## Descripcion
Paquete `internal/supervisor` (parte de ingestión): `SourceProcess` ejecuta ffmpeg como subproceso supervisado por source — resuelve la URL vía `Extractor`, decide remux (`-c copy`) o transcode normalizador según el sondeo `ffprobe`, escribe segmentos TS normalizados en `~/.rodritv/streaming/ch-<n>/src-<id>/`, reinicia con backoff exponencial con jitter, marca la source `degraded` tras 5 fallos consecutivos y re-resuelve la URL por intervalo del extractor o ante errores 403/404.
## Proyecto
rodritv-playlist
## Depende de
US-0008 (Source), US-0009 (proclog, telemetry), US-0010 (Extractor)
## Valor
Al desplegarla, una source activa produce segmentos TS normalizados de duración `segment_seconds` de forma robusta (reinicios acotados, logs crudos y estructurados de cada fallo), lista para que el `ChannelMixer` (US-0012) la consuma.
## Checklist:
- [ ] T-0034: Crear sondeo de perfil con ffprobe en `internal/mixer/probe.go`
- [ ] T-0035: Crear `SourceProcess` con ffmpeg supervisado y backoff en `internal/supervisor/sourceproc.go`
- [ ] T-0036: Agregar re-resolución de URL por intervalo y ante 403/404 a `SourceProcess`
