# US-0018: Render de templates, layout VSCode y assets estáticos
## Descripcion
Crear la base de presentación: paquete raíz `web/` con `embed.FS` de `templates/` y `static/` (`web/web.go`), htmx vendored y pinneado en `static/vendor/htmx.min.js`, el cargador de templates `internal/panel/render.go` (parseo único al arranque con `template.ParseFS`, FuncMap con helpers `formatUptime`, `formatTime`, `stateClass`, fail fast ante error de sintaxis) y el esqueleto visual: `layout.html` (sidebar / main / bottom), `panel.css` (grid de 3 regiones en desktop, colapso a columna con paneles plegables < 768 px) y `panel.js` (toggle de paneles, helpers htmx). Se sirve `/static/*` desde el embed con `Cache-Control: max-age=86400`.
## Proyecto
rodritv-web
## Depende de
US-0016 (rutas y servidor)
## Valor
El binario sirve un layout navegable vacío (sidebar/main/bottom responsivo) con estáticos embebidos y htmx disponible; los templates se validan en CI con un test de parseo.
## Checklist:
- [ ] T-0056: Crear paquete `web/` con embed.FS, htmx vendored y handler `/static/*`
- [ ] T-0057: Crear cargador de templates y FuncMap en `internal/panel/render.go`
- [ ] T-0058: Crear `layout.html`, `panel.css` (grid responsivo) y `panel.js`
