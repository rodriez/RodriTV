# US-0022: Dockerfiles multi-stage de las 3 imágenes
## Descripcion
Crear los Dockerfiles multi-stage de los 3 binarios Go en `build/`: etapa de build común sobre `golang:1.27-alpine` (`CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`, caché de capas con `go mod download` primero) y runtime `alpine:3.21` con `ca-certificates`, usuario no-root `rodritv` (uid 1000), binario en `/app/`, `EXPOSE` y `HEALTHCHECK` contra `/healthz`. La imagen de playlist agrega `ffmpeg`, `chromium` y yt-dlp como binario release pinneado con sha256 verificado.
## Proyecto
rodritv-infra
## Depende de
— (primera US de la EPIC; los `docker build` completos requieren los binarios de EPIC-0001/0002/0003, pero los archivos se mergean de forma independiente)
## Valor
Las 3 imágenes `rodritv-api:<tag>`, `rodritv-playlist:<tag>`, `rodritv-web:<tag>` se construyen de forma reproducible con versiones pinneadas; rollback de imagen = rollback de yt-dlp (RFC §8).
## Checklist:
- [ ] T-0066: Crear `build/api/Dockerfile` multi-stage
- [ ] T-0067: Crear `build/web/Dockerfile` multi-stage
- [ ] T-0068: Crear `build/playlist/Dockerfile` con ffmpeg, Chromium y yt-dlp pinneado
