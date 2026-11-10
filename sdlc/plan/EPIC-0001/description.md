# EPIC-0001: API REST de gestión de canales, sources, configuración, métricas y logs

- SPECS de origen: SPECS-001 (`sdlc/specify/SPECS-001-rodritv-api.md`)
- RFC de origen: RFC-001
- Roles: Tech Lead + Software Engineer
- Fecha: 2026-11-09
- Proyectos involucrados: `rodritv-api` (binario `cmd/api`, puerto 9291)

## Objetivo técnico
Al completar todas las US queda construido el binario `cmd/api` que migra el esquema al arrancar (golang-migrate sobre `embed.FS`) y expone bajo `/v1` el CRUD de canales (con sources embebidas en JSONB), la configuración KV tipada, las métricas de uptime, los logs de procesos, el servicio de logos y `/healthz`, todo sobre PostgreSQL 17 con pgx/sqlc y tests unitarios + integración (testcontainers).

## Alcance
- Entra: plataforma base (config Viper, pool pgx, migraciones, servidor HTTP con middleware y mapeo uniforme de errores), dominios `channel`, `setting`, `metric`, `proclog`, `maintenance` (retención 35 días), queries sqlc y migración `000001_init` embebida.
- NO entra: lógica de streaming/failover (SPECS-002), escritura de `uptime_samples`/`process_logs` (SPECS-002), UI (SPECS-003), Dockerfile/compose (SPECS-004), autenticación.

## Dependencias externas (fuera del Plan)
Ninguna gestión no-codeable detectada. Advertencia técnica: los dominios `metric` y `proclog` son de solo lectura; sus datos los escribe `rodritv-playlist` (SPECS-002). Hasta que ese servicio exista, US-0005 solo es verificable con datos sintéticos (previsto en sus tests). Además, el contrato con el panel (`rodritv-web`) depende del formato uniforme de error y del versionado `/v1` — cualquier desviación obliga a reabrir la SPECS.

## Nota de verificación
El repositorio es greenfield (sin código Go). Todos los componentes citados son `[NUEVO]` y siguen la estructura de SPECS-001.3; no hubo nombres existentes que verificar.

## Historias de usuario (ordenadas por dependencias)
- [x] [US-0001: Plataforma base del binario `cmd/api`](./US-0001/description.md) ✅ (completada 2026-11-10)
- [x] [US-0002: Dominio channel: entidad, servicio y repositorio PostgreSQL](./US-0002/description.md) ✅ (completada 2026-11-10)
- [x] [US-0003: Transporte HTTP de canales y logos](./US-0003/description.md) ✅ (completada 2026-11-10)
- [US-0004: Dominio setting: configuración de comportamiento (KV tipado)](./US-0004/description.md)
- [US-0005: Dominios metric y proclog: consultas de uptime y logs (solo lectura)](./US-0005/description.md)
- [US-0006: Job de retención de métricas y logs (maintenance)](./US-0006/description.md)
