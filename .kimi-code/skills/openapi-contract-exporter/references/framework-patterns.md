# Framework Discovery Patterns

How to locate endpoints, contracts, and error responses in Java/Spring Boot and Go codebases.

## Contents
- Java / Spring Boot
- Go (net/http, Gin, Echo, Chi)
- Error-response discovery (both)
- Naming sanitization

## Java / Spring Boot

**Endpoints**
- Class-level: `@RestController`, `@Controller` + `@RequestMapping` (base path).
- Method-level: `@GetMapping`, `@PostMapping`, `@PutMapping`, `@PatchMapping`, `@DeleteMapping`, `@RequestMapping(method=...)`. Concatenate class base path + method path.
- Params: `@PathVariable`, `@RequestParam` (name, `required`, `defaultValue`), `@RequestHeader`.
- Body: `@RequestBody` type; media types from `consumes`/`produces`.

**Contracts**
- DTO/record classes: fields + Jackson annotations (`@JsonProperty`, `@JsonIgnore`, `@JsonInclude`).
- Validation: Jakarta/Bean Validation (`@NotNull`, `@NotBlank`, `@Size`, `@Min/@Max`, `@Pattern`, `@Email`) → OpenAPI `required`, `minLength/maxLength`, `minimum/maximum`, `pattern`, `format`.
- Enums → `enum:` with serialized values (`@JsonValue` if present).

**Errors**
- `@ControllerAdvice` / `@RestControllerAdvice` classes: `@ExceptionHandler` methods map exceptions → `@ResponseStatus`/ResponseEntity status and error body type.
- `ResponseStatusException` thrown in services/controllers.
- Validation errors: `MethodArgumentNotValidException` → 400 shape.
- Spring Security: 401/403 when endpoints are secured (`@PreAuthorize`, SecurityFilterChain rules).

## Go

**Endpoints**
- net/http: `http.HandleFunc`, `mux.Handle`, `ServeMux` patterns (Go 1.22+ method patterns: `"GET /users/{id}"`).
- Gin: `router.GET/POST/...`, `group := r.Group("/api")`.
- Echo: `e.GET(...)`, groups.
- Chi: `r.Get/Post/...`, `r.Route`, `r.Mount`.
- Follow router composition into sub-groups to build full paths.

**Contracts**
- Request/response structs: `json:"name"` tags are the external field names; `binding:"required"` / validator tags → `required` and constraints.
- Decode sites (`json.NewDecoder(r.Body).Decode(&v)`, `c.Bind(&v)`) identify request bodies; encode sites (`c.JSON(status, v)`, `json.NewEncoder(w).Encode(v)`) identify responses **and their status codes**.

**Errors**
- Every `c.JSON(http.StatusBadRequest, ...)`, `http.Error(w, msg, code)`, custom error types, and error-handling middleware → collect each distinct status + body shape.
- Centralized error handlers/middleware often define the canonical error schema — reuse it under `components/schemas`.

## Error-response discovery (both)

1. Grep for status-code literals and error constructors (`StatusBadRequest`, `ResponseStatus`, `HttpStatus.`, `4xx/5xx`).
2. Group by (status code, body shape); each group becomes one response entry with a real example from the code.
3. Add 401/403 per security enforcement and 400 for validated request bodies, even if not explicitly coded, when the framework produces them automatically (bean validation, binding failures) — note this in the response description.

## Naming sanitization

- Strip suffixes: `Dto`, `DTO`, `Model`, `Entity`, `Controller`, `Service`, `ServiceImpl`, `Repository`, `Impl`, `Bean`, `Po`, `VO`, `Req`, `Resp`. Strip prefixes: `Abstract`, `Base` (when noise), `I` (interface marker).
- Keep semantic qualifiers: `CreateUser` and `UpdateUser` are fine if they carry meaning; `UserDto` → `User`.
- Collision after stripping (`UserDto` and `UserEntity` both → `User`): keep one as `User`, qualify the other by usage (`UserProfile`, `UserSummary`) based on its fields.
