# RodriTV — Documento de Requisitos de Producto (PRD)

**Estado:** APPROVED
**Fecha:** 2026-11-09
**Fuente:** `docs/funcional/general.md` + entrevista de discovery

## Declaración del problema

El contenido audiovisual gratuito está disperso en muchos sitios web, plataformas de video en vivo y enlaces de stream directos, cada uno con su propio reproductor y sus propios problemas de confiabilidad. El televidente mira este contenido en un televisor LG (webOS) a través de la aplicación SS-IPTV, que solo puede consumir una lista unificada de canales IPTV — no puede navegar sitios web ni extraer streams por sí misma.

Hoy no hay forma de ver estas fuentes cómodamente en el televisor: cada fuente vive detrás de una página web distinta, los streams se caen sin aviso y, cuando uno se cae, no hay nada a lo que cambiar. La alternativa actual (abrir cada sitio en una computadora) rompe por completo la experiencia de mirar TV.

Una demo técnica ya demostró la premisa central: el contenido de páginas web, YouTube en vivo y listas de stream directas puede unificarse en una sola lista de canales IPTV que el televisor LG reproduce correctamente (video y audio) mediante SS-IPTV.

## Usuarios e interesados

- **Televidente / usuario único** — mira TV en su casa y administra el panel de control. Es un producto hogareño de un solo usuario; no hay otros consumidores ni administradores en v1. Espera una experiencia de TV sin interrupciones visibles.
- La responsabilidad por las licencias de los streams de terceros es asumida por el televidente y queda **fuera del alcance** del producto.

## Funcionalidad deseada (v1)

### Lista unificada de canales
- El producto expone una única lista de canales IPTV de acceso libre (sin autenticación) con todos los canales habilitados, ordenados según la numeración elegida por el televidente comenzando en 1.
- La lista se consume desde SS-IPTV en el televisor LG; solo un canal se transmite al televisor por vez (SS-IPTV maneja la sintonización con el control remoto).

### Canales y sources
- Cada canal soporta: nombre, logo, un enlace opcional para extraer su programación, un número de orden comenzando en 1, una opción para activar/desactivar el canal completo, y hasta 5 sources.
- Cada source soporta: un número de orden comenzando en 1 (1 = source principal), la ubicación del stream, un tiempo de buffering configurable, una opción para activar/desactivarla, y si requiere extracción mediante navegador o cuenta con un extractor dedicado.
- Tipos de source soportados en v1: listas de stream directas (m3u8 — se retransmiten tal cual), YouTube en vivo, y fuentes privadas/web cuando existe un extractor dedicado.
- Las sources de YouTube soportan la opción de "streaming rotativo": dado el canal de YouTube, el producto descubre por sí mismo el enlace de la transmisión en vivo actual.

### Reproducción continua y sin interrupciones
- Se prioriza la fluidez sobre la actualidad: cada canal tiene un buffer configurable de hasta 10 minutos. El televidente acepta explícitamente que la reproducción puede ir hasta ese tiempo atrás de la transmisión real en vivo.
- **Conmutación transparente entre sources (necesidad central):** el sistema debe poder cambiar la señal de salida de un canal de una source a otra en cualquier momento — ante una caída, una degradación o una decisión operativa — **sin generar ninguna interrupción perceptible para el cliente** (sin corte, congelamiento, salto de audio o necesidad de re-sintonizar en el televisor). La mezcla entre señales es continua.
- Las sources de un canal se mantienen encendidas precisamente para hacer posible esa conmutación instantánea: cuando una se cae, la reproducción continúa desde la siguiente sin que el televidente note la interrupción.
- Si un canal no tiene ninguna source funcionando, el televidente ve una señal de video de "problemas técnicos" mientras el producto sigue intentando restablecer la transmisión automáticamente.
- Cada canal registra el uptime de sus sources. Si el uptime de la source principal supera el 95%, las demás sources pueden permanecer apagadas y la señal de problemas técnicos cubre las interrupciones. Si el uptime cae por debajo del 95%, las sources de respaldo se encienden hasta que el uptime global del canal se recupere por encima del 95%.
- Cada canal guarda logs y métricas de sus procesos en ejecución, visibles en el panel de control.

### Gestión de encendido de canales
- El producto considera exactamente un canal como "el que se está consumiendo" y mantiene encendidos ese canal más una cantidad configurable de vecinos hacia atrás y hacia adelante, para un cambio de canal instantáneo.
- Al iniciar el servidor se asume que el canal 1 es el actual, por lo que los vecinos configurados se encienden alrededor de él.
- Una configuración aparte define cuántos canales se encienden al iniciar el servidor.
- SS-IPTV transmite solo un canal por vez, por lo que la cantidad de canales transmitidos simultáneamente se mantiene acotada.

### Panel de control
- Panel de control web para crear, editar y borrar canales y sus sources, y para cambiar la configuración global (cantidad de vecinos, canales encendidos al inicio).
- En PC, el panel sigue un diseño similar a VSCode: menú de íconos a la izquierda, área central que renderiza las pantallas seleccionadas, y panel derecho con formularios de canales, configuración y logs.
- El panel también debe funcionar desde un smartphone Android con diseño responsivo.

## Cómo se probará (aceptación a nivel producto)

La entrega de v1 se acepta cuando, frente al televisor LG con SS-IPTV, el televidente pueda confirmar:

1. La lista unificada reproduce video y audio correctamente para canales alimentados por cada tipo de source soportado (página web con extractor dedicado, YouTube en vivo, lista de stream directa) — replicando lo que la demo ya validó, pero producido por el producto real.
2. Zapping: el cambio de canal es inmediato gracias a la configuración de vecinos alrededor del canal actual.
3. Failover: derribar la source principal de un canal no produce una interrupción visible; la reproducción continúa desde la siguiente source.
4. Caída total: derribar todas las sources de un canal muestra la señal de problemas técnicos, y la reproducción se reanuda automáticamente cuando una source se recupera.
5. Panel: se puede crear un canal nuevo completo (con sus sources) desde el panel usando el smartphone, y está reproduciéndose en el televisor en menos de 5 minutos.

Rechazo inmediato: congelamientos o cortes visibles en el televisor durante la reproducción normal, o una lista que SS-IPTV no pueda cargar.

## Métricas de éxito

| Métrica | Línea base | Objetivo | Medición |
|---|---|---|---|
| Interrupciones visibles por hora de consumo | Desconocida (hoy: la reproducción simplemente se corta) | < 1 | 30 días de uso real |
| Uptime por canal (según medición del propio producto) | n/d | ≥ 95% | 30 días de uso real |

Si las interrupciones/hora no bajan de 1 tras 30 días, se reevaluará si el error estuvo en el problema (fuentes poco confiables) o en la solución (estrategia de failover/buffer).

## Boceto de diseño

Diseño de tres regiones estilo VSCode para el panel en PC (menú de íconos a la izquierda, pantallas al centro, formularios/logs a la derecha), con fallback responsivo en smartphone. No se proporcionó otro boceto.

## Fuera de alcance / versiones futuras

- **v2:** aplicación cliente nativa para LG webOS, reemplazando SS-IPTV.
- Soporte multi-usuario, autenticación en la lista de canales, televidentes adicionales.
- Aplicación o validación de licencias de contenido — responsabilidad del televidente.
- Renderizado de guía de programación (EPG) en el cliente de TV (en v1 solo se guarda el enlace opcional de programación por canal).
- Extractores para fuentes más allá de las configuradas en el set inicial de canales.

## Set inicial de canales (referencia)

Telefe, El Trece, TN, América TV, El Nueve, NeoTV, CM, Cosmos, Disney Channel, ESPN Premium, TNT Sports, Fox Sports, TyC Sports — cada uno con las sources listadas en `docs/funcional/general.md`.
