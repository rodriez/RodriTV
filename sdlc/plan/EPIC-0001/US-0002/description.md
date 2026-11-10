# US-0002: Dominio channel: entidad, servicio y repositorio PostgreSQL

## Descripcion
Implementación del dominio `channel`: entidades con invariantes, errores de dominio, interfaz `Repository` definida por el consumidor, queries sqlc, repositorio PostgreSQL con mapeo de errores y servicio de orquestación con validaciones y normalización.

## Proyecto
rodritv-api

## Depende de
US-0001

## Valor
CRUD completo de canales con sources embebidas (5 slots) persistido en `channels`, con validaciones de dominio y tests de integración reales. Aún sin rutas HTTP (código muerto temporal hasta US-0003).

## Checklist:
- [x] T-0006: Modelar entidades y errores de dominio en `internal/channel/entity.go` y `source.go` [NUEVO]
- [x] T-0007: Definir la interfaz `Repository` en `internal/channel/gateways.go` [NUEVO]
- [x] T-0008: Escribir queries sqlc de canales en `scripts/sqlc/channel.sql` y generar código [NUEVO]
- [x] T-0009: Implementar `Repository` PostgreSQL en `internal/channel/postgres.go` [NUEVO]
- [ ] T-0010: Implementar `Service` de canales en `internal/channel/service.go` [NUEVO]
