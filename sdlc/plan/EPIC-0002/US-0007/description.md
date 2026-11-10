# US-0007: Plataforma base del binario `cmd/playlist`
## Descripcion
Base transversal del servicio `rodritv-playlist`: configuración por variables de entorno con Viper (prefijo `RODRITV_PLAYLIST_`), pool de conexiones PostgreSQL compartido, servidor HTTP en :9290 con middleware (logging, recover) y `/healthz`, y `main` con inyección de dependencias manual y graceful shutdown.
## Proyecto
rodritv-playlist
## Depende de
— (primera US de la EPIC; comparte convenciones de plataforma con US-0001 de EPIC-0001, pero no código)
## Valor
Al desplegarla, el binario arranca, valida su configuración, conecta a PostgreSQL, responde `/healthz` (200/503) y se apaga limpiamente con señal, matando procesos hijos. Aún sin motor ni playlist.
## Checklist:
- [ ] T-0020: Crear `Config` con Viper en `cmd/playlist/config.go`
- [ ] T-0021: Crear pool pgx compartido en `internal/platform/db.go`
- [ ] T-0022: Crear servidor HTTP :9290 con middleware y `/healthz` en `cmd/playlist/server.go`
- [ ] T-0023: Componer `cmd/playlist/main.go` con DI y graceful shutdown
