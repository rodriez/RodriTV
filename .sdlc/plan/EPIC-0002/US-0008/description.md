# US-0008: Dominio catalog: lectura de canales/config y watch en caliente
## Descripcion
Dominio de lectura del motor: tipos de dominio (`Channel`, `Source`, `BehaviorConfig`), interfaces gateway definidas por el consumidor, implementación PostgreSQL con sqlc sobre las tablas `channels` (sources JSONB) y `config`, y un watcher que relee la configuración cada 15 s y notifica diffs para aplicar cambios en caliente sin reiniciar.
## Proyecto
rodritv-playlist
## Depende de
US-0007
## Valor
Al desplegarla, el servicio puede leer el catálogo completo (canales habilitados ordenados por número, con sus sources por slot) y la configuración de comportamiento, y reacciona a cambios en la DB en ≤ 15 s. Base de Engine (US-0014) y StreamServer (US-0015).
## Checklist:
- [ ] T-0024: Definir tipos de dominio e interfaces gateway en `internal/catalog/gateways.go`
- [ ] T-0025: Crear queries sqlc `catalog.sql` y generar código en `internal/catalog`
- [ ] T-0026: Implementar gateways PostgreSQL en `internal/catalog/postgres.go`
- [ ] T-0027: Implementar watcher de config en caliente en `internal/catalog/watch.go`
