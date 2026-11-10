# US-0010: Extractores de stream (m3u8, YouTube, web headless)
## Descripcion
Paquete `internal/extractor` con la interfaz `Extractor` (`Resolve` + `RefreshInterval`, según SPECS-002.3.3), una factory por tipo de source y las tres implementaciones: m3u8 pass-through con re-resolución de URLs con token, YouTube vía yt-dlp (streaming rotativo con re-resolución cada 10 min) y web vía chromedp headless con pool acotado a 2 browsers.
## Proyecto
rodritv-playlist
## Depende de
US-0008 (tipos `Source`/`SourceType` de catalog)
## Valor
Al desplegarla, cualquier source configurada puede resolverse a una URL de stream lista para ffmpeg, con re-resolución automática según su naturaleza (tokens caducables, vivos rotativos de YouTube, players JS).
## Checklist:
- [ ] T-0030: Crear interfaz `Extractor` y factory en `internal/extractor/extractor.go`
- [ ] T-0031: Implementar extractor m3u8 con detección de token en `internal/extractor/m3u8.go`
- [ ] T-0032: Implementar extractor YouTube con yt-dlp en `internal/extractor/youtube.go`
- [ ] T-0033: Implementar extractor web con chromedp en `internal/extractor/web.go`
