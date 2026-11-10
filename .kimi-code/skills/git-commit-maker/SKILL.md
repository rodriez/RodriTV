---
name: git-commit-maker
description: Usar SIEMPRE que se necesite hacer un commit en git (crear, preparar o proponer un commit, o cuando el usuario diga "commitea", "hacé commit", "guardá los cambios"). Define el formato obligatorio del título [<EPIC>|<US>|<T>]: <titulo> y arma el mensaje siguiendo el template .github/PULL_REQUEST_TEMPLATE.md.
---

# git-commit-maker

Esta skill es obligatoria cada vez que haya que hacer un commit. No se debe ejecutar `git commit` sin seguir estos pasos.

## 1. Revisar los cambios

1. Ejecutar `git status` para ver qué hay modificado y qué está en staging.
2. Ejecutar `git diff --staged` (y `git diff` si no hay nada en staging) para entender qué cambió realmente.
3. Si no hay nada para commitear, avisar al usuario y detenerse.
4. Si hay cambios sin agregar, preguntar al usuario si debe incluirlos (`git add`) antes de continuar. No agregar archivos por iniciativa propia si no es evidente.

## 2. Armar el título

El título del commit debe tener SIEMPRE esta estructura:

```
[<EPIC>|<US>|<T>]: <titulo>
```

Donde:

- `<EPIC>`: identificador de la épica (ej. `EPIC-12`).
- `<US>`: identificador de la historia de usuario (ej. `US-345`).
- `<T>`: identificador de la tarea (ej. `T-678`).
- `<titulo>`: resumen breve y claro de lo que hace el commit, en modo imperativo, sin punto final y idealmente de hasta ~72 caracteres en total.

Ejemplo:

```
[EPIC-12|US-345|T-678]: Agregar validación de email en el formulario de registro
```

### Cómo obtener EPIC, US y T

Buscar en este orden y usar el primero que dé resultado:

1. Lo que el usuario haya indicado en la conversación.
2. El nombre de la rama actual (`git branch --show-current`), por ejemplo `feature/EPIC-12_US-345_T-678-validacion-email`.
3. Commits anteriores de la rama (`git log -n 10 --oneline`) que ya usen el formato.

Si no se puede determinar alguno de los tres valores, **preguntar al usuario**. Nunca inventar identificadores ni dejar placeholders (`<EPIC>`, `XXX`, etc.) en el commit final.

## 3. Armar el mensaje (cuerpo)

El cuerpo del commit debe respetar la estructura del template ubicado en:

```
.github/PULL_REQUEST_TEMPLATE.md
```

Pasos:

1. Leer el archivo `.github/PULL_REQUEST_TEMPLATE.md` en cada commit (no asumir su contenido, puede haber cambiado).
2. Mantener las mismas secciones, encabezados y orden que define el template.
3. Completar cada sección con información real basada en el diff y en lo que el usuario haya contado. Si un campo no aplica, escribir `N/A`; no borrar secciones.
4. Si el template tiene checkboxes (`- [ ]`), marcar (`- [x]`) solo los que efectivamente se cumplen.
5. No inventar información (tests ejecutados, capturas, tickets, etc.) que no se pueda respaldar. Si falta un dato necesario, preguntar al usuario.
6. Quitar los comentarios HTML del template (`<!-- ... -->`) en el mensaje final.

Si el archivo `.github/PULL_REQUEST_TEMPLATE.md` **no existe**, avisar al usuario y preguntar si quiere continuar solo con el título y un cuerpo breve describiendo los cambios.

## 4. Crear el commit

1. Mostrar al usuario el título y el mensaje completos antes de confirmar, salvo que ya haya pedido explícitamente que se commitee directamente.
2. Crear el commit con el título como primera línea, una línea en blanco y luego el cuerpo. Para mensajes multilínea usar un archivo temporal o varios `-m`:

   ```bash
   git commit -m "[EPIC-12|US-345|T-678]: Agregar validación de email" -m "<cuerpo según el template>"
   ```

3. Verificar con `git log -1` que el commit quedó con el formato correcto.

## Reglas

- Nunca usar `--no-verify`, `--amend` ni `push` salvo que el usuario lo pida explícitamente.
- Un commit = un cambio lógico. Si el diff mezcla cambios no relacionados, sugerir dividirlo en varios commits.
- El título siempre lleva los tres identificadores entre corchetes separados por `|`, seguidos de `: ` y el título.
- El cuerpo siempre sigue el template de `.github/PULL_REQUEST_TEMPLATE.md`.
