# US-0015: StreamServer: playlist M3U, HLS por canal, logos y healthz
## Descripcion
Superficie HTTP del servicio para SS-IPTV (`internal/streamserver` + rutas en `cmd/playlist`): renderiza `/playlist.m3u` en memoria por request (canales habilitados ordenados por número, con tags `tvg-chno`/`tvg-name`/`tvg-logo` y URL absoluta al HLS del canal), sirve `index.m3u8` y segmentos por canal desde `${STREAMING_DIR}` registrando la request como "canal actual" en el `PowerScheduler`, sirve logos desde `${DATA_DIR}/logos` con anti path-traversal, y cablea el `/healthz` real (DB + Engine). Cierra el wiring del `main` con el Engine completo.
## Proyecto
rodritv-playlist
## Depende de
US-0007 (server/routes), US-0008 (catalog), US-0014 (Engine/scheduler)
## Valor
Al desplegarla, el servicio cumple todo el contrato de SPECS-002.5: SS-IPTV puede descargar la lista, sintonizar canales (encendiéndolos bajo demanda al pedir segmentos) y consumir HLS continuo; el operador tiene `/healthz` real.
## Checklist:
- [ ] T-0046: Crear render de la M3U maestra en `internal/streamserver/playlist.go`
- [ ] T-0047: Crear handlers HLS, logos y registro de canal actual en `internal/streamserver/handlers.go`
- [ ] T-0048: Cablear rutas, healthz real y Engine en `cmd/playlist` (integración final)
