# US-0021: Pantalla Monitoreo: métricas y logs con polling htmx
## Descripcion
Implementar la página `GET /monitoring` (vista de todos los canales con uptime 24h/7d/30d, estado de sources y contador de interrupciones) y los fragmentos con polling `GET /frag/channels/{id}/metrics?window=24h` (`hx-trigger="every 5s"`) y `GET /frag/channels/{id}/logs?level=all` (`hx-trigger="every 3s"`, filtrable por nivel). Los mismos fragmentos de métricas/logs se reutilizan en el bottom panel de la pantalla Canales (solo del canal seleccionado). Ventana de métricas seleccionable (1h/24h/7d/30d). Errores de la API → banner "API no disponible, reintentando" (el polling reintenta solo).
## Proyecto
rodritv-web
## Depende de
US-0016, US-0017, US-0018, US-0019 (reutiliza los partials del bottom panel)
## Valor
El admin ve la salud de todos los canales y, por canal, uptime por ventana, estado de sources, interrupciones y logs en vivo (polling), tanto en Monitoreo como en el bottom panel de Canales. Cierre de la EPIC: incluye la validación manual del layout responsivo en Android.
## Checklist:
- [ ] T-0064: Crear página de monitoreo con tarjetas de uptime por canal
- [ ] T-0065: Crear fragmentos de métricas y logs con polling htmx
