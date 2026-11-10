# US-0025: Makefile raíz del repo
## Descripcion
Crear el `Makefile` en la raíz del repo como punto de entrada único para desarrollo y release: targets `up`/`down`/`logs`/`ps`/`clean` sobre `docker compose -f deployments/docker-compose.yml`, `build`/`test`/`vet`/`generate` sobre la toolchain Go y sqlc, `release TAG=x.y.z` delegando en `scripts/release.sh`, y `help` como target por defecto. Todos los targets `.PHONY`.
## Proyecto
rodritv-infra
## Depende de
US-0023 (targets de compose), US-0024 (target `release` invoca `scripts/release.sh`). El target `generate` queda operativo cuando exista `sqlc.yaml` (EPIC-0001 / T-0008).
## Valor
Un solo punto de entrada documentado (`make help`) para levantar el entorno, correr tests con coverage (excluyendo código generado por sqlc) y cortar releases.
## Checklist:
- [ ] T-0073: Crear `Makefile` raíz con todos los targets
