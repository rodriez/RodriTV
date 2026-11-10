# US-0005: Dominios metric y proclog: consultas de uptime y logs (solo lectura)

## Descripcion
Implementación de los dominios `metric` y `proclog` en modo solo lectura: agregación de muestras de uptime con cálculo de interrupciones, listado paginado de logs de procesos, y exposición vía HTTP con verificación de existencia del canal.

## Proyecto
rodritv-api

## Depende de
US-0001, US-0002 (necesita `channel.Service` para resolver 404 del canal).

## Valor
`GET /v1/channels/{id}/metrics` y `GET /v1/channels/{id}/logs` operativos sobre datos que escribirá rodritv-playlist.

## Checklist:
- [ ] T-0016: Crear dominio `metric`: queries sqlc y repositorio de uptime [NUEVO]
- [ ] T-0017: Crear `Service` y `Handler` de métricas y registrar la ruta [NUEVO]
- [ ] T-0018: Crear dominio `proclog`: repositorio, servicio, handler y ruta de logs [NUEVO]
