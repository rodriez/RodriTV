# US-0012: ChannelMixer: salida HLS continua, failover y slate
## Descripcion
`ChannelMixer` (`internal/mixer`): compone la salida HLS del canal en `~/.rodritv/streaming/ch-<n>/index.m3u8` + `seg-<seq>.ts` copiando/renombrando los segmentos de la source activa con secuencia de medios y PROGRAM-DATE-TIME monótonos independientes de la source; emite `#EXT-X-DISCONTINUITY` en el límite de failover; conmuta al slate de "problemas técnicos" (loop lavfi cacheado) cuando ninguna source está sana; y aplica retención de `buffer_minutes` con pruning de disco. Incluye el spike PoC del mixer con 2 sources reales (gate del RFC).
## Proyecto
rodritv-playlist
## Depende de
US-0011 (segmentos normalizados de SourceProcess), US-0009 (process_logs)
## Valor
Al desplegarla, el canal produce una playlist HLS que el cliente (SS-IPTV) puede consumir de forma continua aunque la source cambie, se caiga o se recupere, sin gaps ni saltos de secuencia, y con señal de problemas técnicos en caída total.
## Checklist:
- [ ] T-0037: Spike PoC: mezcla continua con 2 sources reales (gate RFC riesgo n.º 1)
- [ ] T-0038: Crear `ChannelMixer` con publicación de segmentos y secuencia monótona en `internal/mixer/mixer.go`
- [ ] T-0039: Crear generación del slate de problemas técnicos en `internal/mixer/slate.go`
- [ ] T-0040: Implementar retención y pruning del buffer en `internal/mixer/mixer.go`
