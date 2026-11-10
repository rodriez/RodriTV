# US-0016: Plataforma base del binario `cmd/web`
## Descripcion
Crear el esqueleto desplegable del binario `cmd/web`: configuración con Viper (prefijo `RODRITV_WEB_`), `http.Server` con middleware de logging `slog` y recover, graceful shutdown, registro de rutas en `routes.go` y `main.go` de composición con inyección de dependencias manual. Al finalizar, el proceso arranca, expone `GET /healthz` (200, independiente de la API) y se apaga limpiamente.
## Proyecto
rodritv-web
## Depende de
— (primera US de la EPIC; sigue las convenciones ya establecidas en EPIC-0001/EPIC-0002)
## Valor
Binario `rodritv-web` deployable que arranca con config validada (fail fast), sirve `/healthz` en :9292 y loguea requests en JSON.
## Checklist:
- [ ] T-0049: Crear `Config` con Viper en `cmd/web/config.go`
- [ ] T-0050: Crear `http.Server` con middleware y ciclo de vida en `cmd/web/server.go`
- [ ] T-0051: Crear `cmd/web/routes.go` y `cmd/web/main.go` con `/healthz`
