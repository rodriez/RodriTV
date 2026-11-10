---
name: component-indexer
description: Generate a component index (component_index.md) for a codebase. Use when the user asks to inventory, document, or map all components of a system — classes, structs, interfaces, functions, and methods with their signatures — from Java and Go source code. Produces a tree grouped by module/package and folder, used as the verified source of real component and method names before writing technical specs or tasks (SDLC support skill).
---

# Component Indexer

Generate `component_index.md`: a complete inventory of the system's components extracted from the actual source code (Java and Go). This index is the authoritative reference for real class/struct/method names — never guess names elsewhere when this index exists.

## Workflow

1. Obtain the project root. If the code is not already available locally, ask the user for the repository or its location.
2. Run the bundled script (pure Bash + awk, works on Linux and macOS with no extra dependencies):

```bash
bash scripts/build_component_index.sh <project_root> --output component_index.md
```

3. Verify the output: check that the number of scanned files is plausible for the project and spot-check a few known components against the real code. If components are missing (e.g. unconventional code layout), inspect manually and complete the index — never leave silent gaps.
4. Deliver `component_index.md` and summarize: modules found, components per module, and file counts.

## Output format

```
<MODULE_NAME> (<PATH_FOLDER>)
└── <COMPONENT_NAME> (<FILE_NAME>)
    └── <METHOD_NAME> (<METHOD_SIGN>)
```

- **MODULE_NAME**: Java package (e.g. `com.shop.orders`) or Go package name.
- **PATH_FOLDER**: path of the folder relative to the project root.
- **COMPONENT_NAME**: class, interface, enum, record (Java) or struct, interface, type, top-level function (Go). Go components carry a kind tag, e.g. `PaymentService [struct]`.
- **FILE_NAME**: file containing the component.
- **METHOD_NAME** / **METHOD_SIGN**: methods of the component with parameters and return types, e.g. `getOrder(Long id) : OrderDto` or `Charge(ctx context.Context, orderID string, amount int64) (string, error)`. Included only when the component has methods; constructors have no return type.

## Rules

- Index ALL `.java` and `.go` files, including test files; exclude only build/vendor directories (`.git`, `vendor`, `target`, `build`, etc. — already skipped by the script).
- The index must reflect the code as it is: exact names, exact signatures. Do not paraphrase, translate, or omit components.
- Signatures keep parameter names and types, and return values (Go includes multiple returns and `error`).
- For polyglot repos, both languages go in the same `component_index.md`, grouped by their respective modules.
- If the script finds zero files or clearly misses components, stop and tell the user instead of delivering an empty or partial index.
