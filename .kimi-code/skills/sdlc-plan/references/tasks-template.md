# Plantilla de salida del Plan (TASKS)

El Plan NO se emite en un solo archivo. Se genera una estructura de directorios con un archivo `description.md` por EPIC y por US, y un archivo `<T-0000>.md` por cada sub-tarea:

```
<EPIC-0000>/
    |-- description.md          # Contiene la descripcion de la epica
    |__ <US-0000>/
            |-- description.md  # Contiene la descripcion de la US
            |__ <T-0000>.md     # Contiene la descripcion de la tarea
```

El encabezado del Plan (SPECS de origen, RFC, roles, fecha, proyectos involucrados) se incluye al inicio del `description.md` de la EPIC.

## Template: `EPIC-0000/description.md`

```markdown
# EPIC-<n>: <título>

## Objetivo técnico 
    <qué capacidad queda construida al completar todas las US>
## Alcance: 
    <qué entra y qué NO entra>
## Dependencias externas (fuera del Plan)
    <gestiones no-codeables detectadas, si las hay>
```

## Template: `US-0000/description.md`

```markdown
# US-<n>: <título>
## Descripcion
    <breve parrafo con una descripcion funcional de la historia>
## Proyecto
    <único proyecto que toca — NUNCA más de uno>
## Depende de
    <US previas, si aplica>
## Valor
    <qué queda funcionando al desplegarla>
## Checklist:
- [ ] <T-0000: Titulo>
```

## Template: `US-0000/T-0000.md`

```markdown
# T-<n>: <verbo + componente> (ej. "Agregar método validateStock() a OrderValidator")
## Descripción técnica:
<qué cambios de código se hacen, en qué archivos. Todo componente citado existe en el
código (verificado) o se marca [NUEVO]. Nombres exactos de clases, métodos, estructuras.>

## Pseudo código:
<lógica a aplicar — obligatorio cuando la lógica no es trivial>

## Criterios de aceptación:
- [ ] Dado <contexto>, cuando <acción>, entonces <resultado observable>
- [ ] <condición técnica verificable>

## Definición de finalizado (DoD):
- [ ] Tests unitarios de la lógica agregada/modificada en verde
- [ ] Tests existentes sin romper
- [ ] Lint/type-check sin errores
- [ ] Código mergeado tras code review
- [ ] Deployable de forma independiente (sin pasos manuales fuera del pipeline)

## Proyecto / módulo: <ruta>
## Tamaño: S (≤ 1 día) — si es mayor, dividir
## Depende de: <tarea>
```

## Ejemplo mínimo (`US-0003/T-0003.md`)

```markdown
# T-3: Agregar método ValidateStock() a OrderValidator
## Descripción técnica:
Agregar el método público ValidateStock(Order order) [NUEVO] a la clase existente
OrderValidator. Inyectar IInventoryRepository (ya existe) por constructor. El método
devuelve ValidationResult (existente).

## Pseudo código:
skus = order.Items.Select(i => i.Sku)
stocks = inventoryRepo.GetBySkus(skus)
para cada item en order.Items:
    si stocks[item.Sku] == null o stocks[item.Sku].Available < item.Quantity:
        agregar Error("INSUFFICIENT_STOCK", item.Sku)
retornar ValidationResult(errores)

## Criterios de aceptación:
- [ ] Dado un pedido con stock suficiente, ValidateStock devuelve resultado sin errores
- [ ] Dado un item sin stock, devuelve error INSUFFICIENT_STOCK con el sku afectado
- [ ] SKU inexistente en inventario se trata como stock 0

## Definición de finalizado (DoD): 
- [ ] Tests unitarios de la lógica agregada/modificada en verde
- [ ] Tests existentes sin romper
- [ ] Lint/type-check sin errores
- [ ] Código mergeado tras code review
- [ ] Deployable de forma independiente (sin pasos manuales fuera del pipeline)

### Proyecto / módulo: orders-service / src/domain/validation/OrderValidator.cs
### Tamaño: S
### Depende de: T-2
```
