---
name: c4-architecture-documenter
description: Generate or update an architecture.md file documenting the code architecture of a system with C4 diagrams in Mermaid, by inspecting the actual Java and Go source code. Use when the user asks to document, diagram, or visualize the architecture of a codebase, generate C4 diagrams, or keep architecture documentation in sync with the code (SDLC support skill). Covers all components and the communication between them (REST, gRPC, messaging, databases, external systems).
---

# C4 Architecture Documenter

Generate `architecture.md`: C4 architecture diagrams (Mermaid) plus a component/communication reference, always derived from the real source code (Java and Go). If the file already exists, update it so it reflects the current state of the code — never leave stale diagrams.

## Workflow

1. **Locate the code.** Get the project root(s). If the code is not available locally, ask the user for the repository or its location. For multi-repo systems, document all repos the user includes in scope.
2. **Locate the output.** Ask the user in which folder to save `architecture.md`, unless they already indicated it. Check whether the file already exists there — that decides between *create* and *update* mode.
3. **Inventory the code.** Enumerate all `.java` and `.go` files (skip `.git`, `vendor`, `target`, `build`, `node_modules`). If a `component_index.md` exists (component-indexer skill), use it as the starting map and verify it is not outdated against the file list; otherwise scan the code directly.
4. **Inspect components and their communication.** Read the code enough to identify, for each module/service:
   - Deployable units (applications, services, workers, schedulers) and their responsibilities.
   - Components inside each unit (controllers/handlers, services/use cases, repositories/gateways, mappers, config, scheduled jobs, consumers/producers).
   - Communication edges, with protocol and direction: REST endpoints, gRPC services, message topics/queues, database access, caches, and calls to external systems. Extract these from real artifacts: route annotations (`@RequestMapping`, `@GetMapping`…, Gin/Echo/net-http routes), gRPC `.proto`/generated clients, messaging listeners/publishers, repository beans/DAO wiring, HTTP clients (Feign, RestTemplate, WebClient, Go http clients), and configuration (application.yml, env) for target systems.
5. **Build the diagrams.** Use Mermaid C4 syntax (see [references/mermaid-c4-syntax.md](references/mermaid-c4-syntax.md) for the exact syntax and a full example). Generate the levels that apply:
   - **System Context**: the system, its users/actors, and external systems.
   - **Container**: one diagram per system (or per bounded context in large monorepos) showing applications, services, databases, queues, and their interactions.
   - **Component**: one diagram per non-trivial container, showing its internal components and edges.
   Keep each diagram readable: if a component diagram exceeds ~20 nodes, split it by layer or functional area.
6. **Write `architecture.md`.** Structure:
   - Title, generation date, and scope (repos/modules covered).
   - Diagrams (Context → Container → Component), each in a ` ```mermaid ` block.
   - After each diagram, a concise reference table: component/container, responsibility, technology (from the code: framework, language), and key interactions (protocol → target).
   - Notes section for anything the code does not reveal (e.g. infrastructure-only aspects), marked as assumptions.
7. **Verify.** Re-check every node and edge in the diagrams against the code: names must match real classes/structs/services; every edge must have a code artifact behind it. Confirm the Mermaid is syntactically valid (balanced braces, declared elements used in relations). Then deliver the file path and a summary of what was documented or what changed.

## Update mode (existing architecture.md)

- Read the existing file and preserve its structure and any sections the user maintains manually.
- Diff it against the current code: add new components/edges, remove deleted ones, rename what changed. Never keep a node or relation that no longer exists in the code.
- Update the generation date and add a brief "Last update" line listing what changed (added/removed/modified components and relations).
- If a documented component no longer exists in the code and it is unclear whether it was removed or renamed, ask the user before deleting it from the document.

## Rules

- Every element in a diagram must come from the code: use real names (class/struct/application names), never invented ones. Do not paraphrase or translate identifiers.
- Document ALL components in scope, not a curated selection. Completeness beats elegance.
- Every communication edge carries its protocol and direction, and must be traceable to code (annotation, client wiring, config). If an interaction is only hinted by configuration, mark it in the notes as inferred.
- Both Java and Go components coexist in the same diagrams when the system is polyglot; tag containers with their technology (e.g. `[Container: Java, Spring Boot]`, `[Container: Go, Gin]`).
- Ask the user when: the output folder is unknown; the scope (repos/modules) is ambiguous; a discrepancy between existing docs and code cannot be resolved from the code itself.
- If the codebase is empty or the scanned files don't match what the user described, stop and report instead of delivering a partial diagram silently.
