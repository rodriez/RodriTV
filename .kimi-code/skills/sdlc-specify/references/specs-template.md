# Plantilla del documento SPECS

Usar esta estructura para cada documento SPECS. Una SPECS por proyecto cuando el RFC afecta varios.

## Encabezado

- **RFC de origen:** <id/título del RFC APPROVED>
- **Proyecto:** <nombre del proyecto cubierto por esta SPECS>
- **SPECS relacionadas:** <otras SPECS del mismo RFC, si existen>
- **Rol:** Tech Lead
- **Estado:** PENDING_REVIEW | IN_REVIEW | APPROVED | REJECTED
- **Fecha / Versión:** <fecha> / <vX>

## 1. Objetivo técnico

Resumen técnico de lo que se construirá, derivado del RFC. Sin lenguaje de negocio.

## 2. Documentación interna consultada

| Documento | Ubicación | Qué aporta |
|---|---|---|
| Arquitectura de código | | composición de capas/módulos |
| Contratos de APIs actuales | | endpoints y esquemas vigentes |
| Convenciones y acuerdos | | naming, estilo, testing, errores, logging |

## 3. Arquitectura de código

- Estructura de carpetas/módulos resultante.
- **Nombres exactos** de módulos, paquetes, clases y componentes a crear o modificar.
- Patrones aplicados (y por qué encajan con la arquitectura existente).
- Buenas prácticas del lenguaje/stack definido en el RFC que se aplicarán (tipado, manejo de errores, inyección de dependencias, inmutabilidad, testing, etc.).

### 3.x Diagrama C4 de código (Mermaid) — si aplica

Incluir diagrama C4 (componente/código) en Mermaid cuando la estructura sea no trivial.

## 4. Decisiones técnicas

Tabla de decisiones tomadas. Cada fila debe estar resuelta (sin TBD):

| # | Decisión | Alternativas consideradas | Decisión final | Justificación |
|---|---|---|---|---|
| 1 | Librería de HTTP client | A, B, C | B | ... |

Para cada librería/proveedor elegido por caso de uso: nombre, versión objetivo y motivo.

## 5. Contratos de API (si aplica)

Endpoints nuevos o modificados: método, ruta, request/response (esquemas), códigos de error, versionado y compatibilidad con contratos existentes.

## 6. Persistencia, integraciones y proveedores (si aplica)

Esquemas, migraciones, colas/eventos, integraciones externas y proveedores concretos.

## 7. Estrategia de testing

Tipos de test por componente y cobertura esperada, alineada a las convenciones del proyecto.

## 8. Supuestos

Supuestos asumidos por falta de documentación interna, pendientes de validar.

## 9. Riesgos y bloqueantes

Decisiones no resueltas (bloqueantes), riesgos técnicos heredados del RFC y mitigaciones.

## 10. Tabla de proyectos afectados

| Proyecto | Tipo (existente/nuevo) | Cambios de la SPECS que aplica | Secciones de la SPECS |
|---|---|---|---|
| | | | |

## 11. Fuera de alcance

Lo que explícitamente no se hará, para delimitar la SPECS frente al Plan.
