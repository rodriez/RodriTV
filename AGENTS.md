# AGENTS.md — RodriTV

## Project overview

RodriTV is a TV channel aggregator capable of ingesting feeds from various sources and broadcasting them through a unified IPTV playlist (see `README.md`).

**Current state:** early development. `cmd/api/` contains the start of the `rodritv-api` binary (Viper-based `Config`/`LoadConfig`, stub `main`). Update this file as real code lands.

## Technology stack

- **Language:** Go (`go 1.27.3`, see `go.mod`)
- **Module path:** `github.com/rodriez/RodriTV`
- **Dependencies:** `github.com/spf13/viper` (config), `github.com/jackc/pgx/v5` (PostgreSQL pool), `github.com/golang-migrate/migrate/v4` (migraciones embebidas), `github.com/stretchr/testify` + `github.com/testcontainers/testcontainers-go` (tests) — see `go.mod`.
- Intended layout suggests a web UI (`web/`) and containerized deployment (`deployments/`), but neither exists yet.

## Repository layout

```
cmd/          # entry points / main packages, standard Go convention
  └── api/    # binario rodritv-api (config.go, main.go)
internal/     # private application code, standard Go convention
  ├── channel/  # dominio de canales, Service, contrato Repository, queries sqlc y PostgresRepository
  └── platform/ # infraestructura compartida (db.go: pool pgx con retry; migrate.go: runner golang-migrate embebido)
migrations/   # migraciones SQL embebidas (000001_init up/down, SPECS-004.6) exportadas como embed.FS
scripts/sqlc/ # queries SQL usadas por sqlc para generar acceso a datos
sqlc.yaml     # configuración de generación Go/pgx en internal/channel
web/          # (empty) intended web frontend
deployments/  # (empty) intended deployment manifests
docs/         # functional/user documentation (docs/funcional/)
build/        # (empty) build artifacts / packaging
sdlc/        # SDLC workflow artifacts (discovery/define/specify/plan docs)
  └── docs/   # generated technical documentation (see below)
.kimi-code/   # Kimi Code CLI agent configuration and skills
go.mod        # Go module definition
README.md     # One-line project description
LICENSE
```

## Generated technical documentation

All generated/maintained technical documentation lives in `sdlc/docs/` — never at the repo root or in `docs/` (which is reserved for functional documentation):

- `sdlc/docs/component_index.md` — component inventory (regenerate with the `component-indexer` skill after code changes).
- `sdlc/docs/architecture.md` — C4 architecture diagrams (update with the `c4-architecture-documenter` skill when components/containers/relations change).

## Mandatory: consult technical documentation first

Before writing, modifying, or refactoring any code, the agent MUST consult the project's technical documentation to locate components and understand the architecture:

1. `sdlc/docs/component_index.md` — first source to map components (structs, interfaces, functions, methods) to their real files.
2. `sdlc/docs/architecture.md` — C4 architecture diagrams, to understand containers, components and relations.
3. Direct code inspection — only as a last resort, when the docs don't exist, are outdated, or raise doubts.

If a component cited as existing is not found in the docs or the code, do not assume — ask the user.

## Build and test commands

Standard Go toolchain:

- Build: `go build ./...`
- Test: `go test ./...`
- Module tidy: `go mod tidy`

## Code style guidelines

No project-specific conventions exist yet. Until established, follow standard Go practice: `gofmt`/`goimports` formatting, `go vet` clean, idiomatic package layout (`cmd/` for binaries, `internal/` for non-public code).

## Testing instructions

Tests use Go's built-in `testing` package with `testify` (`assert`/`require`), in `*_test.go` files colocated with the code. Run `go test ./...`.

## Security considerations

- Never commit secrets; `.gitignore` currently only excludes `.vscode`.
- IPTV aggregation can involve third-party feeds and copyrighted streams — validate and sanitize external feed inputs, and be mindful of content-licensing constraints when implementing ingestion.
