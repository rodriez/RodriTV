# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

RodriTV is a TV channel aggregator: it ingests feeds from various sources and is meant to broadcast them as a unified IPTV playlist. So far only the channel CRUD API exists (Go 1.27, PostgreSQL via pgx, standard library `net/http` routing).

## Commands

```sh
make init     # download modules, install mockgen, regenerate mocks
make mocks    # regenerate gomock mocks (go generate ./...)
make sql      # regenerate internal/sqlc from scripts/db (needs sqlc on PATH)

go build ./...
go vet ./...
go test ./...

# Tests are testify suites: -run picks the suite, -testify.m picks the method
go test ./internal/tvchannels -run TestService
go test ./internal/tvchannels -run TestHandler -testify.m TestAddChannelWithInvalidJson

DATABASE_URL=postgres://... go run ./cmd/api   # listens on :9290
```

The server pings PostgreSQL at startup and exits if it cannot connect. There is no migration runner: apply `scripts/db/migrations/` by hand. `SERVER_PORT` is read into the config, but the port is hardcoded to 9290.

## Architecture

`cmd/api` is the composition root. It keeps dependencies in package-level vars (`main.go`) and wires them with `init*` functions: pgx pool → `sqlc.Queries` → repository → service → handler. Routes are registered on the default mux in `routes.go`.

`internal/tvchannels` is a single package split into layers by filename prefix:

- `api.*`: HTTP handler, request/response contracts, and mapping of errors to `ApiError` responses.
- `domain.*`: `TvChannel` and `TvSignal` entities, validation, `Service`, and the typed errors.
- `infrastructure.*`: repository over the sqlc queries, plus conversion from DB `Record` to domain type.

Layers are connected by interfaces declared in `*.gateways.go`: the handler depends on `TvChannelServiceGateway` (`api.gateways.go`) and the service on `TvChannelStoreGateway` (`domain.gateways.go`).

Things that span several files:

- **Errors:** the type of an error decides the HTTP status. `FromError` maps `ValidationError` to 400, `MissingResourceError` to 404 and anything else to 500. Add new errors as typed constants in `domain.errors.go`. The service logs store errors and returns `ErrInternal`.
- **Signals:** `TvSignalSet` is a fixed `[5]TvSignal` array, and zero-valued slots mean "unused" (skipped by validation and responses). It is stored as one JSON column on `channels`, using the Go field names because `TvSignal` has no JSON tags. The request/response contracts expose only `url`, `priority` and `needBrowser`, as an anonymous struct that tests repeat literally, tags included.
- **New channels:** the service assigns the ID (standard library `uuid`) and forces status `PAUSED`, ignoring what the client sent.

## Generated code

- `internal/sqlc/` comes from `scripts/db/queries/` and `scripts/db/migrations/` (config in `scripts/db/sqlc.yaml`). Edit the SQL, then run `make sql`.
- `internal/tvchannels/*.gateways_test.go` are mockgen output despite the `_test.go` names. Change the interface, then run `make mocks`.

Hand-written tests are in package `tvchannels_test` and use testify suites with these gomock mocks.
