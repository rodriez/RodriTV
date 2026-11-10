# Architecture — RodriTV

- Fecha de generación: 2026-11-10
- Alcance: repositorio `github.com/rodriez/RodriTV` (módulo `cmd/api`)
- Última actualización: se agrega el binario `cmd/api` con `Config`/`LoadConfig` (Viper) y stub de `main` (T-0001, US-0001).

## System Context

```mermaid
C4Context
    title System Context — RodriTV

    Person(user, "Usuario IPTV", "Consume la playlist unificada")
    System(rodritv, "RodriTV", "Agregador de canales de TV que expone una playlist IPTV unificada")
    System_Ext(feeds, "Fuentes externas", "Feeds de canales de terceros")

    Rel(user, rodritv, "Consume playlist", "HTTP")
    Rel(rodritv, feeds, "Ingiere feeds", "HTTP")
```

| Elemento | Responsabilidad | Tecnología | Interacciones |
|---|---|---|---|
| RodriTV | Agregador de canales con playlist IPTV unificada | Go 1.27 | HTTP → fuentes externas (inferido del PRD/RFC) |

## Container

```mermaid
C4Container
    title Container — RodriTV

    Container(api, "cmd/api", "Go", "Binario de la API: configuración, ciclo de vida y endpoints HTTP")
    System_Ext(feeds, "Fuentes externas", "Feeds de canales de terceros")

    Rel(api, feeds, "Ingiere feeds", "HTTP (planificado)")
```

| Contenedor | Responsabilidad | Tecnología | Interacciones |
|---|---|---|---|
| cmd/api | Binario de la API (config, arranque, HTTP) | Go 1.27, Viper | HTTP → fuentes externas (planificado, aún sin código) |

## Component — cmd/api

```mermaid
C4Component
    title Component — cmd/api

    Component(config, "Config / LoadConfig", "Go, Viper", "Carga y valida la configuración desde variables de entorno RODRITV_API_*")
    Component(main, "main", "Go", "Punto de entrada (stub, pendiente de composición en T-0005)")

    Rel(main, config, "Cargará y validará config al arrancar", "llamada directa (planificado)")
```

| Componente | Responsabilidad | Tecnología | Interacciones |
|---|---|---|---|
| Config / LoadConfig (cmd/api/config.go) | Struct de configuración y carga con Viper: defaults (`PORT=9291`, `DATA_DIR=/data`, `LOG_LEVEL=info`, `SHUTDOWN_TIMEOUT=10s`), `AutomaticEnv` con prefijo `RODRITV_API`, validación fail-fast (`Validate`) | Go, spf13/viper | Variables de entorno del proceso |
| main (cmd/api/main.go) | Punto de entrada del binario; stub vacío hasta T-0005 | Go stdlib | — |

## Notas / supuestos

- La interacción con fuentes externas y la base de datos PostgreSQL proviene del PRD/RFC; aún no existe código de ingesta ni de persistencia (llega en T-0002/T-0003).
- El servidor HTTP y el endpoint `/healthz` llegan en T-0004; la composición del ciclo de vida en T-0005.
