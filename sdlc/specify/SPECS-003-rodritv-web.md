# SPECS-003: rodritv-web — Servidor del panel de control

- **RFC de origen:** RFC-001 — Motor de streaming con failover continuo para lista IPTV unificada (APPROVED, `sdlc/define/RFC-001-streaming-engine-failover-iptv.md`)
- **Proyecto:** `rodritv-web` (binario `cmd/web`, imagen `rodritv-web`, puerto **9292**)
- **SPECS relacionadas:** SPECS-001 (rodritv-api), SPECS-002 (rodritv-playlist), SPECS-004 (rodritv-infra)
- **Rol:** Tech Lead
- **Estado:** APPROVED
- **Fecha / Versión:** 2026-11-09 / v1.0

## 1. Objetivo técnico

Servidor HTTP que renderiza en servidor el panel de control (layout tipo VSCode: barra lateral, área principal, panel inferior) con `html/template`, enriquecido con **htmx + JavaScript vanilla + CSS** (sin build de frontend). Consume la API REST de `rodritv-api` (:9291) como único backend de datos. Responsivo: usable en PC y en smartphone Android. Sin autenticación (LAN, decisión del PRD).

## 2. Documentación interna consultada

| Documento | Ubicación | Qué aporta |
|---|---|---|
| Arquitectura de código | — (no existe) | Repo greenfield; se asume layout estándar Go (ver SPECS-003.8 Supuestos) |
| Contratos de APIs actuales | SPECS-001.5 | Contrato completo de la API que este servidor consume |
| Convenciones y acuerdos | `AGENTS.md` | Stack Go 1.27.3, gofmt/vet, tests con `testing` |

## 3. Arquitectura de código

### Estructura de paquetes

Arquitectura **screaming** según `golang-standards/project-layout`, con **máximo 2 niveles de profundidad** dentro de `internal/`. Templates y assets estáticos viven en el directorio raíz `web/` (convención project-layout):

```
cmd/web/main.go                     # composición e inyección de dependencias manual
cmd/web/routes.go                   # registro de rutas del ServeMux (páginas y fragmentos htmx)
cmd/web/config.go                   # Viper: env vars, defaults, validación (arranque)
cmd/web/server.go                   # construcción del http.Server, middleware (logging, recover), ciclo de vida
internal/panel/                     # dominio del panel (transporte + render)
  pages.go                          # handlers de páginas completas (layout + contenido inicial)
  fragments.go                      # handlers de fragmentos htmx (canales, sources, config, métricas, logs)
  render.go                         # carga de templates con embed.FS, funciones helper (formato de %, fechas)
  pages_test.go, fragments_test.go, render_test.go
internal/apiclient/                 # gateway de salida hacia rodritv-api
  gateways.go                       # interfaz APIClient (la define el consumidor: panel)
  client.go                         # cliente HTTP tipado (un método por endpoint de SPECS-001.5)
  models.go                         # tipos espejo del contrato JSON de la API
  client_test.go
web/                                # raíz (project-layout): templates y assets
  web.go                            # embed.FS de templates/ y static/
  templates/layout.html             # esqueleto VSCode: sidebar / main / bottom panel
  templates/pages/*.html            # páginas: channels, config, monitoring
  templates/partials/*.html         # fragmentos htmx (filas, formularios, tarjetas, líneas de log)
  static/css/panel.css              # estilos propios (sin framework CSS)
  static/js/panel.js                # JS vanilla: toggle de paneles, selección de canal, helpers htmx
  static/vendor/htmx.min.js         # htmx vendored (pinneado, sin CDN)
```

### Patrones y buenas prácticas

Se aplican las reglas no negociables de la skill `go-backend`:

- **Comunicación por interfaces definidas por el consumidor:** el paquete `panel` define en `internal/apiclient/gateways.go` la interfaz `APIClient` que consume; `client.go` la implementa. Los handlers jamás hablan HTTP con la API directamente. *Accept interfaces, return structs*: nunca se retornan interfaces ni punteros innecesarios.
- **Inyección de dependencias manual en `cmd/web/main.go`**; el main conoce todas las capas, ninguna capa conoce al main.
- **Standard library primero:** ServeMux, `html/template`, `net/http`, `slog`. Dependencias externas justificadas en SPECS-003.4 (Viper, testify; htmx es asset estático, no dependencia Go).
- **Server-side rendering con htmx parcial:** las páginas se sirven completas en el primer `GET`; las interacciones (alta/edición/borrado, refresh de métricas/logs) son `hx-get`/`hx-post` que devuelven **fragmentos HTML** desde los handlers de `fragments.go`. Sin JSON en el frontend salvo lo estrictamente necesario.
- **Templates con `embed.FS`** (paquete raíz `web/`) compilados una vez al arranque (`template.ParseFS`); error de parseo = fallo al iniciar (fail fast). Funciones helper registradas (`formatUptime`, `formatTime`, `stateClass`).
- **Cliente de API tipado** (`internal/apiclient`): `http.Client` con timeout de 10 s, base URL por config, propagación de `context.Context`, decodificación de errores de la API (`error.code/message`) a errores Go que el handler traduce a mensajes en el fragmento de respuesta.
- **htmx vendored** en `web/static/vendor/` con versión pinneada: el panel debe funcionar sin internet (LAN hogareña).
- **CSS propio, mobile-first:** grid de 3 regiones en desktop (sidebar 260 px / main / bottom 220 px) que colapsa a una columna con paneles plegables en < 768 px.
- **Sin estado en el servidor web:** todo el estado vive en la API/DB; el servidor web es un proxy de renderizado (reinicio sin pérdida).
- Logging `slog` JSON igual que los demás binarios; errores de la API se muestran al usuario con su `message` dentro del fragmento (htmx `hx-target-error`).

### Pantallas

1. **Canales (página principal):** sidebar = lista de canales ordenada por número (nombre, logo, indicador enabled y estado on/off); main = detalle del canal seleccionado con formulario de edición (nombre, número, logo, URL de programación, habilitado, duración de segmento) y una grilla de sus 5 slots de sources (URL y habilitada por slot; el slot define el rol principal/backup — SPECS-001.5.2; alta/baja/cambio editando slots dentro del mismo formulario); bottom = métricas y logs del canal.
2. **Configuración global:** formulario con las claves de SPECS-001.5.3 (vecinos, canales al inicio, TTL, buffer, umbral) con validación de rangos en cliente y servidor.
3. **Monitoreo:** vista de todos los canales con uptime 24h/7d/30d, estado de sources y contador de interrupciones; bottom con logs filtrables por nivel.

## 4. Decisiones técnicas

| # | Decisión | Alternativas consideradas | Decisión final | Justificación |
|---|---|---|---|---|
| 1 | Stack UI | SPA React+Vite, htmx+templates | **htmx + `html/template` + JS vanilla + CSS propio** | Decisión del usuario/RFC: sin build de frontend, suficiente para el layout tipo VSCode. |
| 2 | Distribución de htmx | CDN, vendored | **Vendored en `static/vendor/htmx.min.js` (v2.x pinneada)** | El panel debe funcionar sin internet; build reproducible. |
| 3 | Actualización de métricas/logs | WebSocket, SSE, polling | **Polling htmx** (`hx-trigger="every 5s"` en métricas, `every 3s` en logs, solo del canal seleccionado) | Un solo usuario; polling es trivial de implementar y depurar; SSE queda como mejora futura. |
| 4 | Router | net/http ServeMux, chi | **`net/http` ServeMux estándar** (patterns Go 1.27+) | Decisión del RFC: sin router externo en ningún binario; los patterns nativos alcanzan para páginas y fragmentos. Consistente con SPECS-001.4.1. |
| 5 | Framework CSS | Bootstrap, Tailwind, propio | **CSS propio (~300 líneas, variables CSS)** | Layout fijo de 3 regiones; un framework añade peso sin beneficio para 3 pantallas. |
| 6 | Formularios | JSON + JS, form posts con htmx | **`hx-post` con `application/x-www-form-urlencoded`**, respuesta = fragmento re-renderizado (con errores inline si la API devuelve 400) | Camino nativo htmx; los errores de validación de la API se muestran junto al campo. |
| 7 | Subida de logos | multipart a la API, copia manual | **Fuera de alcance**: el admin copia el archivo a `~/.rodritv/data/logos` respetando la convención de nombre `<channel_id>.<ext>` (SPECS-001.5.1); el formulario muestra el nombre resultante | Un solo administrador; la API ya sirve `/logos/{filename}`. |
| 8 | Config de arranque | caarlos0/env, manual | **Viper** (prefijo `RODRITV_WEB_`) | Decisión del usuario, consistente en los 3 binarios. |
| 9 | Cliente API | generado (oapi-codegen), manual | **Manual tipado** (`internal/apiclient`) | ~12 endpoints; un cliente manual de ~200 líneas es más simple que codegen. |

## 5. Contratos de API (puerto 9292, consumidos por el browser)

Convención: rutas de **página** devuelven HTML completo; rutas bajo `/frag/` devuelven fragmentos para htmx.

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/` | Redirect a `/channels` |
| GET | `/channels` | Página principal (lista + detalle del primer canal) |
| GET | `/channels/{id}` | Página principal con canal `{id}` seleccionado |
| GET | `/frag/channels` | Fragmento: lista de canales (sidebar) |
| GET | `/frag/channels/{id}` | Fragmento: detalle + formulario del canal |
| POST | `/frag/channels` | Crear canal (form) → fragmento lista + detalle, o errores inline |
| POST | `/frag/channels/{id}` | Actualizar canal (form; htmx no envía PUT nativo de forms sin extensión) |
| POST | `/frag/channels/{id}/delete` | Eliminar canal → fragmento lista |
| — | — | Las sources se editan dentro del formulario del canal (SPECS-001.5.2): la grilla de 5 slots es parte de `POST /frag/channels/{id}` (el formulario envía el array completo) |
| GET | `/config` | Página de configuración global |
| POST | `/frag/config` | Guardar configuración → fragmento con confirmación o errores |
| GET | `/monitoring` | Página de monitoreo |
| GET | `/frag/channels/{id}/metrics?window=24h` | Fragmento: tarjeta de métricas (polling 5 s) |
| GET | `/frag/channels/{id}/logs?level=all` | Fragmento: líneas de log (polling 3 s) |
| GET | `/static/*` | CSS/JS/htmx (embed.FS, `Cache-Control: max-age=86400`) |
| GET | `/healthz` | `200` si el proceso sirve (no depende de la API; la UI muestra "API no disponible" en fragmentos con error) |

Manejo de errores de la API en fragmentos: 400 → re-render del formulario con errores por campo; 404 → fragmento con aviso y lista refrescada; 5xx/red → fragmento con banner "API no disponible, reintentando" (htmx reintenta por polling).

## 6. Persistencia, integraciones y proveedores

- **Sin base de datos propia.** Única integración: `rodritv-api` (`API_URL`, default `http://rodritv-api:9291` en la red compose); el cliente consume los recursos bajo el prefijo `/v1` (SPECS-001.5).
- Templates y estáticos embebidos con `embed.FS` (binario autocontenido).
- **Variables de entorno (Viper, prefijo `RODRITV_WEB_`):**

| Var | Default | Descripción |
|---|---|---|
| `PORT` | `9292` | Puerto HTTP |
| `API_URL` | `http://localhost:9291` | Base URL de rodritv-api |
| `API_TIMEOUT` | `10s` | Timeout de llamadas a la API |
| `LOG_LEVEL` | `info` | Nivel slog |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown |

## 7. Estrategia de testing

**TDD obligatorio** (skill `go-backend`): red → green → refactor por caso de uso. Assertions con `testify/assert` + `testify/require`.

- **Caja negra:** cada componente se testea por su comportamiento observable vía su interfaz pública.
- **Mocks solo en los extremos** con `testify/mock`: el único extremo es la API REST, a través de la interfaz `APIClient` (`internal/apiclient/gateways.go`). No se mockean colaboradores internos.
- **`apiclient`:** unit tests contra `httptest.Server` que emula la API (respuestas y errores del contrato de SPECS-001.5); verifica decodificación, timeouts y mapeo de errores.
- **Handlers (`panel`):** `httptest` con `APIClient` mockeado (testify/mock); verifica status, fragmento devuelto (contiene/omite elementos clave vía `strings.Contains`), mapeo de errores 400/404/500 a fragmentos adecuados.
- **`render.go`:** test que parsea todos los templates embebidos (detecta errores de sintaxis en CI) y prueba las funciones helper.
- **Manual (criterio de aceptación del RFC):** alta de canal desde smartphone Android y navegación general del layout responsivo.
- Cobertura objetivo: ≥ 75% en `panel` y `apiclient`.

## 8. Supuestos

- Greenfield: convenciones estándar Go asumidas.
- El contrato de la API (SPECS-001.5) es estable durante la implementación; cualquier cambio se refleja en `client/models.go`.
- El admin edita logos copiando archivos al volumen `data/logos` (sin upload en v1).
- Versiones exactas (htmx, Viper) fijadas en Implement; major versions indicadas son objetivo.

## 9. Riesgos y bloqueantes

Sin bloqueantes. Riesgos menores:

| Riesgo | Mitigación |
|---|---|
| El polling htmx genera carga innecesaria con varias pestañas abiertas | Un solo usuario (PRD); intervalos conservadores (3–5 s) y polling solo del canal seleccionado |
| Layout complejo en pantallas chicas | CSS mobile-first con paneles plegables; validación manual en Android (criterio de aceptación) |
| Versiones de htmx con breaking changes | Vendored y pinneado; actualización manual controlada |

## 10. Tabla de proyectos afectados

| Proyecto | Tipo | Cambios de la SPECS que aplica | Secciones |
|---|---|---|---|
| rodritv-web | nuevo | Todo el servicio: server, handlers, templates, estáticos, cliente API | 3–6 |
| rodritv-infra | nuevo | Imagen `rodritv-web`, servicio compose con `depends_on: api (healthy)` | 6 (detalle en SPECS-004) |

## 11. Fuera de alcance

- Autenticación del panel (PRD: sin auth, solo LAN).
- Upload de logos desde la UI (v2).
- SSE/WebSocket para métricas en tiempo real (mejora futura sobre polling).
- Edición de configuración de arranque (env vars); solo configuración de comportamiento vía API.
- Internacionalización (UI en español, hardcoded).
