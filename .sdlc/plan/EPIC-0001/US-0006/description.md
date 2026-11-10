# US-0006: Job de retención de métricas y logs (maintenance)

## Descripcion
Job interno de mantenimiento dentro del binario de la API que elimina diariamente las muestras de uptime y los logs de procesos más antiguos que la retención configurada (35 días).

## Proyecto
rodritv-api

## Depende de
US-0001 (plataforma).

## Valor
Limpieza automática nocturna de `uptime_samples` y `process_logs` (retención 35 días, SPECS-004.6) dentro del binario de la API.

## Checklist:
- [ ] T-0019: Crear job de retención en `internal/maintenance` [NUEVO]
