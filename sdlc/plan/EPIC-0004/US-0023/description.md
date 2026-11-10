# US-0023: `docker-compose.yml` y `.env.example` de 4 servicios
## Descripcion
Crear `deployments/docker-compose.yml` con los 4 servicios (db `postgres:17-alpine` sin puerto publicado, api, playlist, web) sobre la red bridge interna `rodritv`, con healthchecks, orden de arranque (`depends_on` con condiciones), volúmenes bajo `~/.rodritv/`, límites de memoria, `shm_size` para Chromium, `stop_grace_period` acorde a `SHUTDOWN_TIMEOUT` y `TZ` desde `.env`. Acompañar con `deployments/.env.example` commiteado sin secretos (tags de imagen, credenciales de DB, timezone).
## Proyecto
rodritv-infra
## Depende de
US-0022 (el compose referencia los Dockerfiles de `build/` como contexto de build y las variables de entorno `/healthz` definidas en SPECS-001/002/003)
## Valor
`docker compose -f deployments/docker-compose.yml up -d --build` levanta el stack completo en orden correcto: db healthy → api migra y queda healthy → playlist/web; la DB nunca es accesible desde la LAN.
## Checklist:
- [ ] T-0069: Crear `deployments/docker-compose.yml` con 4 servicios, healthchecks y orden de arranque
- [ ] T-0070: Crear `deployments/.env.example` sin secretos
