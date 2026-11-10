# US-0017: Cliente API tipado `internal/apiclient`
## Descripcion
Construir el gateway de salida hacia `rodritv-api`: modelos espejo del contrato JSON de SPECS-001.5 (Channel con array fijo de 5 slots de Source, Config, Metrics, Logs, error uniforme), la interfaz `APIClient` definida por el consumidor en `gateways.go` y su implementación HTTP tipada en `client.go` (un método por endpoint, `http.Client` con timeout configurable, propagación de `context.Context`, decodificación de errores `error.code/message/fields` a errores Go tipados).
## Proyecto
rodritv-web
## Depende de
— (independiente de US-0016; US-0019..US-0021 la consumen)
## Valor
Los handlers del panel obtienen una interfaz única y testeable para hablar con la API; los errores de la API llegan tipados (validation con fields, not_found, conflict, internal/red) listos para traducirse a fragmentos.
## Checklist:
- [ ] T-0052: Crear modelos espejo del contrato en `internal/apiclient/models.go`
- [ ] T-0053: Crear interfaz `APIClient` y base del cliente HTTP en `internal/apiclient/gateways.go` y `client.go`
- [ ] T-0054: Implementar métodos de canales en `internal/apiclient/client.go`
- [ ] T-0055: Implementar métodos de config, métricas y logs en `internal/apiclient/client.go`
