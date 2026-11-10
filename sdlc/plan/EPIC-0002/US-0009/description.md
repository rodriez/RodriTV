# US-0009: Dominios telemetry y proclog: escritura de métricas y logs
## Descripcion
Dominio de escritura del motor: interfaces gateway de telemetría (muestras de uptime agregadas cada 30 s y eventos de log de procesos), implementación PostgreSQL con sqlc sobre `uptime_samples` y `process_logs`, y el logger de logs crudos por proceso a `~/.rodritv/logs` con rotación lumberjack y buffer de últimas 50 líneas de stderr para el resumen estructurado.
## Proyecto
rodritv-playlist
## Depende de
US-0007
## Valor
Al desplegarla, SourceProcess (US-0011) y ChannelSupervisor (US-0013) pueden persistir muestras de uptime (consumidas luego por la API de métricas, SPECS-001.5.4) y logs de procesos con el detalle exacto del error (comando, exit code, stderr), tanto en DB como en archivo rotado.
## Checklist:
- [ ] T-0028: Crear queries sqlc `telemetry.sql` e implementar gateways en `internal/telemetry`
- [ ] T-0029: Crear logger de procesos con rotación en `internal/proclog/logger.go`
