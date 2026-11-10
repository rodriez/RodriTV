# US-0004: Dominio setting: configuración de comportamiento (KV tipado)

## Descripcion
Implementación del dominio `setting`: entidad tipada con validación de rangos, repositorio KV sobre la tabla `config`, servicio con defaults y handler REST para lectura y reemplazo total de la configuración.

## Proyecto
rodritv-api

## Depende de
US-0001 (plataforma). Independiente de US-0002/US-0003.

## Valor
`GET/PUT /v1/config` con validación de rangos y defaults sembrados por la migración.

## Checklist:
- [ ] T-0014: Crear dominio `setting`: entidad, gateways, queries sqlc y repositorio [NUEVO]
- [ ] T-0015: Crear `Service` y `Handler` de config y registrar `GET/PUT /v1/config` [NUEVO]
