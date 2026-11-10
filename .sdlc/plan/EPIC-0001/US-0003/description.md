# US-0003: Transporte HTTP de canales y logos

## Descripcion
Exposición del dominio channel vía HTTP: handler REST de canales con DTOs y formato de error uniforme, almacenamiento y serving seguro de logos, y registro de rutas con inyección de dependencias en `main.go`.

## Proyecto
rodritv-api

## Depende de
US-0002

## Valor
API REST de canales usable: `GET/POST /v1/channels`, `GET/PUT/DELETE /v1/channels/{id}` y `GET /v1/logos/{filename}` con formato de error uniforme.

## Checklist:
- [ ] T-0011: Implementar `Handler` de canales en `internal/channel/handler.go` [NUEVO]
- [ ] T-0012: Implementar almacenamiento y serving de logos (`LogoStore` + `GET /v1/logos/{filename}`) [NUEVO]
- [ ] T-0013: Registrar rutas de canales en `cmd/api/routes.go` e inyectar dependencias en `main.go` [NUEVO]
