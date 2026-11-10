---
name: openapi-contract-exporter
description: Generate a third-party-facing openapi.yaml (OpenAPI 3.1) from source code by discovering every endpoint and deriving the real request/response contracts — including all error responses — from the actual code components. Use when the user asks to generate, export, or document an OpenAPI/Swagger specification (openapi.yaml, openapi.json, Swagger file) from an existing Java/Spring Boot or Go codebase, for external consumers, API portals, or contract documentation.
---

# OpenAPI Contract Exporter

Generate an `openapi.yaml` (OpenAPI 3.1) that renders cleanly in Swagger UI and is written for third-party consumers, derived entirely from what the code actually does — never from assumptions.

## Workflow

### 1. Gather required inputs

Ask the user (if not already provided) for:
- **Authentication mechanism** exposed by the API (Bearer/JWT, API Key, OAuth2, Basic, mTLS, none, or a mix). Map it to `components.securitySchemes` and apply `security` globally or per-operation as the code enforces it.
- **Servers/hosts to expose** (e.g. `https://api.example.com/v1`, plus sandbox/staging if requested). Use them in `servers:` with `description` per entry.
- **Output folder** where `openapi.yaml` will be written.
- **API title, version, and contact/license info** for the `info` block (propose sensible defaults from the project name and let the user confirm).

Do not start analysis until auth, servers, and output folder are known.

### 2. Discover every endpoint

Scan the whole codebase for HTTP route definitions. Do not rely on a single file or on existing docs — the code is the source of truth. Framework-specific discovery patterns: see [references/framework-patterns.md](references/framework-patterns.md).

Build an inventory of: HTTP method, path (with path variables), handler, request body type, response types per status code, query/path/header params, and consumed/produced media types. Cross-check that no discovered route is omitted from the final spec.

### 3. Derive real contracts from components

For each endpoint, resolve the actual types:
- Request bodies, response bodies, query/path params → trace to the concrete structs/classes and their fields, types, validation constraints (required, format, min/max, pattern, enum), and JSON serialization names (e.g. `@JsonProperty`, `json:"..."` tags). Use serialized names, never internal field names.
- **Error responses**: find every error the code can return — exception handlers, error middleware, validation failures, `4xx/5xx` returns, custom error types. Each distinct error shape and status code gets a response entry with a **concrete example** taken from the code (real messages, real codes). Group reusable error schemas under `components/schemas` and reference them.
- Status codes: only document codes the code can actually produce (including 401/403 when security applies).

### 4. Sanitize naming for third parties

The spec is for external consumers. Never leak internal implementation vocabulary:
- Strip prefixes/suffixes like `Dto`, `DTO`, `Model`, `Entity`, `Controller`, `Service`, `Repository`, `Impl`, `Request`/`Response` (when used as suffix noise), `Bean`, `Po`, `Vo` from schema names and operation summaries. `CreateUserRequestDto` → schema `CreateUser` or `User`; pick the clearest domain name.
- Rename schemas to domain terms (`User`, `Order`, `PaymentError`), de-duplicating collisions with a meaningful qualifier.
- `operationId`: camelCase verb+noun (`createUser`, `getOrderById`), no technical suffixes.
- Descriptions must describe business behavior, not classes/methods. Never mention packages, classes, functions, or files in descriptions.

### 5. Write openapi.yaml (OpenAPI 3.1)

Structure:
- `openapi: 3.1.0`, `info` (title, version, description, contact/license if provided).
- `servers` from user input.
- `tags` grouping operations by domain resource.
- `paths`: every discovered endpoint with params, `requestBody` (with `required` and `content`), and `responses` covering success **and every error found in step 3**, each with schema and `examples` (or a realistic `example`).
- `components`: `schemas` (sanitized names, validation constraints, `required` arrays, enums), `securitySchemes` per the chosen auth, reusable `responses`/`parameters` only when shared by 3+ operations.
- Global `security` requirement unless the API is public.
- Prefer `$ref` over inline duplication. YAML must be valid and Swagger-UI renderable: no tabs, consistent 2-space indent, quote strings containing `:`.

### 6. Validate before delivering

- Parse the YAML and validate it as OpenAPI 3.1 (e.g. `python -c` with `openapi-spec-validator`, or `npx @redocly/cli lint`). Fix every error; treat warnings about missing descriptions as errors for a third-party doc.
- Verify: endpoint count in spec == endpoint count discovered in code; every schema name passes the no-technical-suffix rule; every operation has at least one error response when the code can produce one.
- Save the file as `openapi.yaml` in the user-chosen folder and report the endpoint coverage summary (total endpoints, per-tag counts, auth scheme, servers).

## Rules

- Never invent endpoints, fields, or error codes not present in the code. If a handler's return type is opaque (e.g. raw `interface{}`/`Object`), inspect what is actually written to the response; if still ambiguous, ask the user rather than guessing.
- Never include internal-only endpoints (health, metrics, actuator, debug, admin) unless the user explicitly asks — but always list what was excluded.
- Deprecations, hidden/internal flags in code annotations are honored.
