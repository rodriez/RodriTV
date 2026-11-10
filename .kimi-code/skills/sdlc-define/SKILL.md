---
name: sdlc-define
description: Fase "Define" de un SDLC basado en SDD (flujo discovery, define, specify, plan, implement). Usar cuando el usuario quiera crear, redactar, revisar o aprobar un RFC (Request for Comments) a partir de un PRD, diseñar la arquitectura de una solución con el modelo C4 (diagramas Mermaid), definir el stack tecnológico, identificar riesgos, o planificar rollout y rollback de una funcionalidad. El agente asume el rol de Software Architect.
---

# SDLC — Define (RFC)

Segunda fase del flujo SDD: `discovery -> define -> specify -> plan -> implement`.
Rol asumido: **Software Architect**. Entrada: PRD (salida del Discovery). Salida: documento **RFC**.

Objetivo: traducir los requerimientos del PRD a una propuesta técnica, combinando el lenguaje ubicuo del negocio con lenguaje técnico. No entrar al nivel de detalle de código/módulos/paquetes: eso pertenece a la fase Specify (SPECS).

## Reglas del flujo

- El ciclo debe ser lo más corto posible; el tamaño lo da el resultado del Discovery. Favorecer propuestas iterables por versiones para entregar valor rápido.
- Puede existir **más de un RFC para un mismo PRD** (propuestas alternativas), pero **solo 1 puede estar en estado APPROVED**. Al aprobar uno, verificar que los demás queden REJECTED.
- Estados del RFC (flujo de aprobación, en este orden):
  `PENDING_REVIEW -> IN_REVIEW -> APPROVED | REJECTED`
  - Todo RFC nuevo nace en `PENDING_REVIEW`. Nunca crearlo directamente en `APPROVED`.
  - Pasar a `IN_REVIEW` cuando el usuario lo someta a revisión.
  - `APPROVED` / `REJECTED` solo tras revisión explícita del usuario.
- La aprobación del RFC habilita la fase Specify (SPECS). No avanzar a Specify sin un RFC en `APPROVED`.

## Workflow

### 1. Verificar la entrada

- Confirmar que existe un PRD en estado `APPROVED`. Si no existe o no está aprobado, indicarlo y detenerse (el Define depende del Discovery).
- Leer el PRD completo: problema, funcionalidad deseada, cómo se va a probar, métricas objetivo y bosquejo de diseño si lo hay.

### 2. Relevar el contexto del proyecto (preguntar antes de diseñar)

Preguntar siempre:

1. **¿El proyecto ya existe?**
   - Si existe: **pedir la documentación existente** (arquitectura actual, stack, ADRs, diagramas, READMEs, RFCs anteriores). Analizarla antes de proponer.
   - Si existe **pero no tiene documentación**: recomendar construir primero la documentación base (al menos: contexto del sistema, stack actual y arquitectura existente en C4 nivel 1-2) **antes de redactar el RFC**. Explicar que diseñar sin conocer el estado actual genera propuestas frágiles. Proceder solo si el usuario lo acepta explícitamente.
   - Si es proyecto nuevo (greenfield): continuar directamente.
2. **Plan de rollout**: cómo se desplegará (feature flags, canary, blue/green, porcentajes, entornos, criterios de avance).
3. **Plan de rollback**: cómo se revierte si algo falla (estrategia de reversa, migraciones de datos reversibles, kill switch, criterios y responsable de disparar el rollback).

Si el usuario no tiene definidos rollout/rollback, proponer opciones razonables según el tipo de solución y registrarlas en el RFC.

### 3. Diseñar la propuesta

Redactar el RFC siguiendo la plantilla de [references/rfc-template.md](references/rfc-template.md). Contenido mínimo obligatorio:

- **Arquitectura C4 en Mermaid**: al menos nivel 1 (System Context) y nivel 2 (Containers); nivel 3 (Components) solo donde la complejidad lo justifique. Usar sintaxis `C4Context` / `C4Container` de Mermaid.
- **Stack y tecnologías**: lenguajes, frameworks, librerías, bases de datos, infraestructura y proveedores, con justificación de cada decisión relevante.
- **Problemas y riesgos**: riesgos técnicos, de negocio y operativos, con mitigaciones.
- **Plan de rollout y rollback** (relevado en el paso 2).

### 4. Propuestas alternativas (opcional)

Cuando haya más de una solución viable, generar varios RFC (uno por propuesta), cada uno con su propio documento, estado independiente y sección de trade-offs. Recordar al usuario que **solo 1 puede ser aprobado**; al aprobar uno, marcar el resto como `REJECTED`.

### 5. Revisión y aprobación

Presentar el RFC en estado `PENDING_REVIEW` y pedir revisión explícita. Gestionar el ciclo de estados según la decisión del usuario. Si se rechaza, registrar el motivo en el documento.

## Convenciones de salida

- Un archivo Markdown por RFC, en el directorio de documentación del proyecto (o el que el usuario indique).
- Nombre sugerido: `RFC-<numero>-<titulo-kebab-case>.md` (ej. `RFC-001-checkout-async.md`).
- Encabezado del documento con: título, estado, PRD de origen, autor/rol (Software Architect), fecha y versión.
- Idioma: el del PRD y del usuario; combinar lenguaje ubicuo del negocio con lenguaje técnico.
