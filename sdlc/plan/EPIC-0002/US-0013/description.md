# US-0013: ChannelSupervisor y regla del 95%
## Descripcion
`ChannelSupervisor` (`internal/supervisor`): ciclo de vida de un canal encendido — levanta el `ChannelMixer` y un `SourceProcess` por source que deba estar activa según la regla del 95%, vigila la salud de cada source, ejecuta el failover lógico (decide qué source alimenta al mixer vía `SetActiveSource`, cayendo al slate cuando ninguna está sana) y persiste muestras de uptime agregadas cada 30 s. La regla del 95% con histéresis de 5 min enciende/apaga sources de backup según el uptime de la principal y del canal.
## Proyecto
rodritv-playlist
## Depende de
US-0011 (SourceProcess), US-0012 (ChannelMixer), US-0009 (telemetry)
## Valor
Al desplegarla, un canal encendido se autogestiona: failover automático entre sources por salud, slate en caída total, métricas de uptime persistidas para la API, y consumo de recursos adaptativo (backups apagados cuando la principal es estable).
## Checklist:
- [ ] T-0041: Crear ventana deslizante de uptime y sampler en `internal/supervisor/uptime.go`
- [ ] T-0042: Crear `ChannelSupervisor` con failover lógico en `internal/supervisor/supervisor.go`
- [ ] T-0043: Implementar la regla del 95% con histéresis en `internal/supervisor/supervisor.go`
