# US-0001: Plataforma base del binario `cmd/api` (arranque, DB, migraciones, health)

## Descripcion
Construcción de la plataforma base del binario `cmd/api`: carga de configuración con Viper, pool de conexiones pgx con retry, migración embebida del esquema inicial y servidor HTTP con middleware compuesto, mapeo uniforme de errores y endpoint `/healthz`.

## Proyecto
rodritv-api

## Depende de
—

## Valor
Binario que arranca, valida config, conecta a PostgreSQL con retry, aplica la migración `000001_init` y expone `GET /healthz`. Deployable sin romper nada (repo vacío).

## Checklist:
- [ ] T-0001: Crear `Config` con Viper en `cmd/api/config.go` [NUEVO]
- [ ] T-0002: Crear pool pgx con retry en `internal/platform/db.go` [NUEVO]
- [ ] T-0003: Crear migración `000001_init` y runner golang-migrate embebido [NUEVO]
- [ ] T-0004: Crear servidor HTTP con middleware, mapeo uniforme de errores y `/healthz` [NUEVO]
- [ ] T-0005: Crear `cmd/api/main.go` con composición y ciclo de vida [NUEVO]
