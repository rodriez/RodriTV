# US-0024: Scripts `install.sh` y `release.sh`
## Descripcion
Crear los scripts operativos del repo: `scripts/install.sh` (bash, idempotente) que verifica Docker + compose v2 en el host, crea el árbol `~/.rodritv/{db,data,streaming,logs}` y `~/.rodritv/data/logos`, copia el compose y genera `.env` con contraseña de DB aleatoria si no existe, y levanta el stack; y `scripts/release.sh` que toma un tag semántico como argumento, lo valida contra `^\d+\.\d+\.\d+$` y construye las 3 imágenes con ese mismo tag.
## Proyecto
rodritv-infra
## Depende de
US-0023 (install.sh copia `deployments/docker-compose.yml` y genera `.env` desde `deployments/.env.example`; release.sh usa los Dockerfiles de US-0022)
## Valor
Instalación del stack en el servidor hogareño con un solo comando, idempotente y sin regenerar credenciales; releases versionados consistentes de las 3 imágenes con un mismo tag (RFC §7).
## Checklist:
- [ ] T-0071: Crear `scripts/install.sh` idempotente
- [ ] T-0072: Crear `scripts/release.sh` con validación de tag semántico
