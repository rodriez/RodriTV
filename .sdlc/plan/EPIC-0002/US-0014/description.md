# US-0014: PowerScheduler y Engine: encendido acotado y config en caliente
## Descripcion
Raíz del motor (`internal/engine`): `PowerScheduler` calcula el conjunto de canales encendidos (canal actual por última request de segmentos con TTL, ±N vecinos circulares en el orden por número, primeros M canales al inicio) y `Engine` (goroutine raíz con errgroup) aplica los diffs encendiendo/apagando `ChannelSupervisor` sin reiniciar el proceso, reaccionando al tick de 15 s, a los snapshots del watcher de catalog y a los cambios de canal actual.
## Proyecto
rodritv-playlist
## Depende de
US-0008 (watcher/catalog), US-0013 (ChannelSupervisor)
## Valor
Al desplegarla, el motor completo funciona: enciende solo los canales necesarios, reconfigura en caliente (neighbors, startup_channels, TTL, buffer, umbral) ante cambios en la DB en ≤ 15 s y apaga canales de forma limpia. El consumo de recursos queda acotado por diseño.
## Checklist:
- [ ] T-0044: Crear `PowerScheduler` con conjunto encendido en `internal/engine/scheduler.go`
- [ ] T-0045: Crear `Engine` con diffs de supervisores y config en caliente en `internal/engine/engine.go`
