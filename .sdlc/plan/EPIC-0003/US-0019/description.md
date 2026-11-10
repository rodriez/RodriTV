# US-0019: Pantalla Canales: páginas, fragmentos y CRUD con grilla de 5 slots
## Descripcion
Implementar la pantalla principal del panel: páginas completas `GET /` (redirect), `GET /channels` y `GET /channels/{id}` (layout + lista en sidebar + detalle del canal seleccionado), fragmentos htmx de lectura (`GET /frag/channels`, `GET /frag/channels/{id}`) y fragmentos de escritura del CRUD (`POST /frag/channels`, `POST /frag/channels/{id}`, `POST /frag/channels/{id}/delete`). El formulario del canal incluye la grilla de 5 slots de sources (URL + habilitada por slot) que se envía como array completo en el reemplazo total (`PUT /v1/channels/{id}` vía `APIClient`). Errores de la API: 400 → formulario re-renderizado con errores por campo; 404 → aviso + lista refrescada; 409 → error en campo `number`; 5xx/red → banner "API no disponible".
## Proyecto
rodritv-web
## Depende de
US-0016, US-0017, US-0018
## Valor
El admin puede listar, crear, editar (incluidas las 5 sources por slot) y eliminar canales desde el browser, con validación inline y recuperación ante errores de la API.
## Checklist:
- [ ] T-0059: Crear handlers de páginas de canales en `internal/panel/pages.go`
- [ ] T-0060: Crear fragmentos de lectura de canales y sus partials en `internal/panel/fragments.go`
- [ ] T-0061: Crear fragmento de alta de canal (POST /frag/channels)
- [ ] T-0062: Crear fragmentos de edición y borrado de canal con grilla de 5 slots
