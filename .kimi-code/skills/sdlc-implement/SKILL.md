---
name: sdlc-implement
description: Etapa Implement del SDLC (discovery, define, specify, plan, implement). Usar cuando el usuario pida implementar, desarrollar o ejecutar una T (sub-tarea) o US del Plan sobre el repositorio. Cubre el rol de Software Engineer que lee la tarea, localiza los archivos a modificar usando primero el índice de componentes (si existe), lee los archivos necesarios, aplica los cambios de código descritos en la tarea usando la SKILL del lenguaje correspondiente si existe, y al finalizar invoca las SKILLs de soporte para actualizar el índice de componentes (component-indexer), la arquitectura de código (c4-architecture-documenter) y el contrato OpenAPI (openapi-contract-exporter), cerrando con un PR y su deploy. Detecta cuándo un ajuste necesario cambia radicalmente el scope de la T/US/EPIC y obliga a re-validar RFC y SPECS en lugar de implementar.
---

# SDLC Implement (PR + deploy)

## Contexto del flujo

Flujo SDLC basado en SDD: `discovery -> define -> specify -> plan -> implement`.

- Entrada: una **T (sub-tarea)** de una US del Plan (con descripción técnica, pseudo código, criterios de aceptación y DoD). La SPECS y el RFC de origen dan contexto.
- Salida: cambios de código aplicados y testeados, documentación técnica actualizada, y un **PR (pull request)** con su procedente deploy.
- Rol: Software Engineer. Etapa **puramente técnica**.

## Flujo de trabajo

### 1. Leer la tarea y su contexto

- Leer la T completa: archivos y componentes citados (existentes o `[NUEVO]`), pseudo código, criterios de aceptación y DoD.
- Leer la US que la contiene y, si es necesario para entender el diseño, la SPECS de origen.
- Si faltan datos esenciales (tarea ambigua, componente no localizable por ninguna vía), preguntar al usuario antes de codificar.

### 2. Localizar los archivos a modificar

Fuentes de localización, **en orden de preferencia**:

1. **Índice de componentes**: si el proyecto tiene un índice de componentes/clases (ej. `component_index.md` generado por `component-indexer`), consultarlo primero para mapear los componentes citados en la T a sus archivos reales.
2. **Documentación interna** (arquitectura de código, ej. `architecture.md`) si el índice no existe o no cubre el componente.
3. **Inspección directa del código**: solo como última opción, cuando las fuentes anteriores no existan, estén desactualizadas o generen duda.

Si un componente citado como existente no se encuentra, no asumir: aplicar el mismo criterio del Plan — preguntar al usuario si es el mismo componente renombrado/refactorizado y cómo agregar la lógica.

### 3. Leer los archivos necesarios

- Leer completos los archivos a modificar y los archivos directamente acoplados (interfaces, DTOs, tests existentes, llamadores del método a cambiar).
- Leer solo lo necesario: no cargar el repositorio entero si el índice ya ubica los puntos de cambio.

### 4. Aplicar los cambios

- **Antes de escribir código, verificar si existe alguna SKILL que especifique cómo desarrollar en el lenguaje del proyecto** (convenciones de Java, Go, etc.). Si existe, leerla y aplicarla obligatoriamente.
- Implementar los cambios descritos en la T: el pseudo código de la tarea guía la lógica; la implementación final sigue las convenciones y buenas prácticas del stack.
- Respetar nombres, paquetes y estructura definidos en la SPECS; los componentes marcados `[NUEVO]` se crean donde la SPECS indica.
- Agregar/actualizar los tests que exijan los criterios de aceptación y el DoD.
- **Ajustes menores permitidos**: renombres triviales, detalles de implementación que no alteran comportamiento. Registrar cualquier desviación de la tarea en el resumen final.
- **Si el ajuste necesario cambia radicalmente el scope de la T, US o EPIC: detenerse.** No implementar. Explicar al usuario por qué e indicar que deben re-validarse y modificarse el RFC y las SPECS (volver a Define/Specify) antes de continuar.

### 5. Verificar

- Compilar/build del proyecto y ejecución de tests relevantes (al menos los que cubren lo tocado).
- Verificar contra los **criterios de aceptación** y el **DoD** de la T: la tarea no está terminada hasta cumplirlos todos.

### 6. Actualizar la documentación técnica (obligatorio al finalizar)

Invocar las SKILLs de soporte, en este orden:

1. **component-indexer**: regenerar/actualizar el índice de componentes para reflejar clases, métodos y estructuras nuevas o modificadas.
2. **c4-architecture-documenter**: actualizar la arquitectura de código (`architecture.md`) si los cambios agregaron/modificaron componentes, contenedores o relaciones.
3. **openapi-contract-exporter**: actualizar el archivo OpenAPI si la tarea agregó, modificó o eliminó endpoints o contratos de API.

Si alguna de las tres no aplica (ej. no hubo cambios de API), indicarlo explícitamente en el resumen en lugar de invocarla. Ajustar también cualquier otra documentación existente afectada por el cambio.

### 7. Cerrar con PR y deploy

- Preparar el PR: rama, commits, título y descripción que referencien la T y la US, incluyendo el resumen de cambios y desviaciones.
- Ejecutar el deploy correspondiente según el proceso del proyecto, o indicar al usuario los pasos exactos si el deploy no puede ejecutarse desde el entorno actual.

## Reglas

- Nunca implementar algo que contradiga la SPECS aprobada: el código sigue a la SPECS; si la realidad del código la invalida, se reabre Specify.
- No tocar código fuera del alcance de la T salvo lo estrictamente necesario para que compile y pase tests; documentar cualquier excepción.
- Una T por vez: no adelantar trabajo de otras T de la US.
- Nunca dejar el build roto ni tests fallando: la T termina en estado desplegable.
