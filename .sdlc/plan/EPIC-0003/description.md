# EPIC-0003: Servidor del panel de control (`rodritv-web`)

- SPECS de origen: SPECS-003 (`.sdlc/specify/SPECS-003-rodritv-web.md`)
- RFC de origen: RFC-001 (`.sdlc/define/RFC-001-streaming-engine-failover-iptv.md`)
- Roles: Tech Lead + Software Engineer
- Fecha: 2026-11-10
- Proyectos involucrados: `rodritv-web` (binario `cmd/web`, puerto 9292)

## Objetivo técnico
Al completar todas las US queda construido el binario `cmd/web` que: sirve en `:9292` el panel de control renderizado en servidor con `html/template` + `embed.FS` (layout tipo VSCode: sidebar / main / bottom panel), enriquecido con htmx vendored + JS vanilla + CSS propio (sin build de frontend); consume la API REST de `rodritv-api` (`/v1`, :9291) como único backend de datos a través del cliente tipado `internal/apiclient`; expone las 3 pantallas (Canales con CRUD y grilla de 5 slots de sources, Configuración global, Monitoreo con métricas y logs con polling htmx); responde `/healthz` y `/static/*`; y funciona responsivo en PC y smartphone Android.

## Alcance
- Entra: plataforma base del binario (config Viper prefijo `RODRITV_WEB_`, `http.Server` con middleware logging/recover, graceful shutdown, `/healthz`), paquete `internal/apiclient` (interfaz `APIClient`, modelos espejo del contrato SPECS-001.5, cliente HTTP tipado con timeout y mapeo de errores `error.code/message/fields`), paquete `internal/panel` (handlers de páginas y fragmentos, `render.go` con FuncMap), templates (`layout.html`, `pages/`, `partials/`), assets estáticos (`panel.css` mobile-first, `panel.js`, htmx vendored pinneado) embebidos con `embed.FS` en `web/web.go`.
- NO entra: la API REST consumida (SPECS-001, EPIC-0001), el motor de streaming (SPECS-002, EPIC-0002), Dockerfile/imagen `rodritv-web` ni servicio compose (SPECS-004), autenticación del panel, upload de logos desde la UI, SSE/WebSocket, i18n (UI en español hardcoded).

## Dependencias externas (fuera del Plan)
Ninguna gestión no-codeable detectada. Advertencias técnicas:
- **El contrato de `rodritv-api` (SPECS-001.5) debe estar implementado** (EPIC-0001) para pruebas end-to-end manuales; los tests de esta EPIC no lo requieren (usan `httptest.Server` y mocks de `APIClient`).
- El asset `web/static/vendor/htmx.min.js` requiere **una descarga puntual** de la release pinneada de htmx (v2.x) durante T-0056; queda vendored en el repo y el build final no necesita internet.
- La validación manual en smartphone Android (criterio del RFC) se realiza al cerrar US-0021 con el compose de SPECS-004; no es una T de código.

## Nota de verificación
El repositorio es greenfield (`cmd/`, `internal/`, `web/` vacíos; no existe código Go). Todos los componentes citados son `[NUEVO]` y siguen la estructura de SPECS-003.3; no hubo nombres existentes que verificar. La numeración continúa la de EPIC-0002 (US-0016..US-0021, T-0049..T-0065).

## Historias de usuario (ordenadas por dependencias)
- [US-0016: Plataforma base del binario `cmd/web`](./US-0016/description.md)
- [US-0017: Cliente API tipado `internal/apiclient`](./US-0017/description.md)
- [US-0018: Render de templates, layout VSCode y assets estáticos](./US-0018/description.md)
- [US-0019: Pantalla Canales: páginas, fragmentos y CRUD con grilla de 5 slots](./US-0019/description.md)
- [US-0020: Pantalla Configuración global](./US-0020/description.md)
- [US-0021: Pantalla Monitoreo: métricas y logs con polling htmx](./US-0021/description.md)
