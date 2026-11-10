---
name: sdlc-plan
description: Etapa Plan del SDLC (discovery, define, specify, plan, implement). Usar cuando el usuario pida planificar, dividir o descomponer una SPECS aprobada en una EPIC con US (historias de usuario) y T (sub-tareas). Cubre el rol de Tech Lead + Software Engineer que genera tareas pequeñas y desplegables, con la regla de que ninguna US toca más de un proyecto, excluyendo toda tarea que no implique codificar (permisos, coordinación con otros equipos, gestiones), verificando contra el código real que los nombres de clases, estructuras y métodos citados existan, preguntando al usuario cuando el código haya cambiado y un nombre ya no sea usable, e incluyendo en cada T pseudo código de la lógica, criterios de aceptación y definición de finalizado (DoD).
---

# SDLC Plan (TASKS)

## Contexto del flujo

Flujo SDLC basado en SDD: `discovery -> define -> specify -> plan -> implement`.

- Entrada: SPECS en estado APPROVED (una o varias, una por proyecto, del mismo RFC).
- Salida: **una EPIC por SPECS** (o por conjunto de SPECS del mismo RFC) con sus US asociadas y, por cada US, las T (sub-tareas) a completar. Todas las US que salgan de una SPECS deben estar englobadas en la misma épica. La salida se materializa como una **estructura de directorios** (un directorio por EPIC y por US, con `description.md` en cada uno y un archivo `.md` por cada T), nunca como un único archivo.
- Roles: Tech Lead en conjunto con Software Engineer. El lenguaje es **puramente técnico**.

Para la plantilla completa de salida, leer [references/tasks-template.md](references/tasks-template.md).

## Reglas de división (innegociables)

1. **Una US nunca toca más de 1 proyecto.** Si un cambio abarca varios proyectos, dividirlo en una US por proyecto y marcar la dependencia entre ellas (orden y contrato que las vincula).
2. **Cada T es pequeña y deployable**: puede mergearse y desplegarse de forma independiente sin romper el sistema (usar feature flags, endpoints versionados o código muerto temporal si hace falta para mantener la desplegabilidad).
3. **Solo tareas con impacto en código.** Nunca incluir tareas que no impliquen codificar: pedir permisos, solicitar accesos, coordinar con otros equipos, gestiones de infraestructura manuales, reuniones, etc. Si la SPECS depende de algo así, registrarlo como "Dependencia externa (fuera del Plan)" y advertirlo al usuario, pero jamás como T.
4. **Descripción sumamente técnica**: la T describe exactamente qué cambios de código se hacen, dónde y cómo.
5. **Nombres reales verificados**: toda clase, estructura, método, interfaz o archivo citado en una T debe existir en el código actual o marcarse explícitamente como `[NUEVO]`. Verificar buscando en el repositorio antes de escribir la tarea.
6. Cada T incluye **criterios de aceptación** y **definición de finalizado (DoD)**.

## Flujo de trabajo

### 1. Verificar precondiciones

- Confirmar que la(s) SPECS están en estado APPROVED. Si no, detenerse e indicarlo.
- Leer las SPECS completas y el RFC de origen para contexto.

### 2. Verificar componentes contra el proyecto

Antes de escribir cualquier tarea, verificar los componentes que la SPECS menciona. Fuentes de verificación, **en orden de preferencia**:

1. **Índice de componentes y clases**: si el proyecto dispone de un índice o catálogo de componentes/clases (documento generado, archivo de índice, wiki técnica, etc.), consultarlo primero para localizar y verificar nombres de clases, estructuras, métodos, interfaces, paquetes y archivos.
2. **Documentación interna** (arquitectura de código, contratos de APIs) si el índice no existe o no cubre el componente.
3. **Inspección directa del código real**: solo como **última opción**, cuando las fuentes anteriores no estén disponibles, estén desactualizadas o generen duda.

- Verificar que los nombres coinciden con los de la SPECS.

**Si un nombre ya no existe o el código cambió desde que se escribió la SPECS**, no asumir: preguntar al usuario presentando:
1. el nombre citado en la SPECS y el componente candidato encontrado en el código (si existe),
2. si se trata del mismo componente renombrado/refactorizado o de uno nuevo,
3. si es el mismo: cómo debe agregarse la lógica (método nuevo, extensión del existente, etc.),
4. esperar la decisión del usuario y registrarla en la tarea.

### 3. Definir la EPIC y las US

- Crear la EPIC: objetivo técnico, SPECS de origen, proyectos involucrados.
- Dividir en US: cada US es implementable y desplegable por sí sola y **toca un solo proyecto**.
- Ordenar las US por dependencias (indicar precedencia cuando exista).

### 4. Dividir cada US en T

Para cada T:

- **Título**: verbo + componente técnico, ej. `Agregar método validateStock() a OrderValidator`.
- **Descripción técnica**: archivos y componentes a tocar (existentes verificados o marcados `[NUEVO]`), con **pseudo código** de la lógica a aplicar cuando no sea trivial:

```
// OrderService.createOrder(order: OrderDTO): Result<Order, OrderError>
items = inventoryRepo.findBySkus(order.skus)
si algún item.stock < item.cantidad:
    retornar Error(INSUFFICIENT_STOCK, skusFaltantes)
order = Order.create(orderDTO)          // valida invariantes de dominio
orderRepo.save(order)
eventBus.publish(OrderCreated(order.id))
```

- **Criterios de aceptación**: verificables, en formato dado/cuando/entonces o checklist técnico.
- **Definición de finalizado (DoD)**: condiciones de cierre (tests, lint, revisión, deploy). Ver plantilla.
- Estimar tamaño: si una T no puede completarse en ~1 día de trabajo, seguir dividiéndola.

### 5. Emitir el Plan

El Plan NO se genera en un solo archivo. Se crea una estructura de directorios con un archivo por elemento:

```
<EPIC-0000>/
    |-- description.md          # Contiene la descripcion de la epica
    |__ <US-0000>/
            |-- description.md  # Contiene la descripcion de la US
            |__ <T-0000>.md     # Contiene la descripcion de la tarea
```

- `EPIC-0000/description.md`: encabezado del Plan (SPECS de origen, RFC de origen, roles: Tech Lead + Software Engineer, fecha, proyectos involucrados) seguido del template de EPIC (objetivo técnico, alcance, dependencias externas).
- `US-0000/description.md`: una carpeta por US, ordenadas por dependencias, con su template (descripción, proyecto, dependencias, valor, checklist de sus T).
- `US-0000/T-0000.md`: un archivo por sub-tarea con su template (descripción técnica, pseudo código, criterios de aceptación, DoD, proyecto/módulo, tamaño, dependencias).
- Usar [references/tasks-template.md](references/tasks-template.md) para los templates exactos de cada archivo.
- Las dependencias externas no-codeables detectadas se listan en la sección "Dependencias externas (fuera del Plan)" del `description.md` de la EPIC (como advertencia, nunca como T).

## Reglas

- Nunca contradecir la SPECS aprobada; si el código obliga a un cambio de diseño, indicar que la SPECS debe reabrirse (volver a Specify). Ajustes menores de naming se resuelven preguntando (paso 2).
- No implementar: el Plan produce tareas, no código. El pseudo código ilustra la lógica, no es la implementación final.
- Nunca crear una T cuyo entregable no sea un cambio de código verificable.
