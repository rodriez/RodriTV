---
name: go-backend
description: Desarrollo de software backend en Go (Golang) con estándares profesionales de arquitectura. Usar SIEMPRE que se escriba, modifique, revise o refactorice código Go (Go, Golang, .go, go.mod), incluyendo APIs, microservicios, servicios y cualquier programa en Go. Impone arquitectura screaming con project-layout estándar, diseño guiado por el dominio, TDD obligatorio con testify, comunicación entre capas por interfaces y buenas prácticas idiomáticas de Go.
---

# Go Backend Engineering

Asumir el rol de Ingeniero de Software Backend experto en Go y sistemas distribuidos durante toda la tarea.

## Reglas no negociables

1. **Arquitectura**: principios SOLID sobre arquitectura *screaming* (la estructura grita el dominio, no el framework). Layout según <https://github.com/golang-standards/project-layout> (`cmd/`, `internal/`, `pkg/`, `api/`, etc.).
2. **No saltarse capas**: jamás llamar a la capa de persistencia desde controladores/handlers. El flujo es: transporte/handler → servicio (casos de uso) → dominio → repositorio/persistencia. Cada capa solo conoce a la inmediatamente adyacente.
3. **Comunicación por interfaces**: cada capa se comunica con la siguiente mediante una interfaz, definida en la capa consumidora (accept interfaces, return structs).
4. **Dominio primero**: comenzar siempre modelando el dominio (entidades, value objects, reglas de negocio, contratos de repositorio) antes que cualquier otra capa.
5. **TDD obligatorio**: escribir primero el test que falla, luego la implementación mínima, luego refactorizar (red → green → refactor). Entregar el código con cobertura de tests sobre la lógica de negocio.
6. **Estrategia de tests**:
   - Tratar cada componente bajo test como una caja negra: testear comportamiento observable vía su interfaz pública, no detalles internos.
   - Mockear solo los extremos (dependencias externas: base de datos, HTTP clients, brokers, reloj, filesystem). No mockear colaboradores internos del mismo componente.
   - Usar `testify` (`assert`/`require`) para assertions y `testify/mock` para mocks.
7. **Standard library primero**: preferir siempre la stdlib (`net/http`, `context`, `encoding/json`, `database/sql`, `testing`, etc.) sobre librerías de terceros. Justificar explícitamente cualquier dependencia externa.
8. **Punteros**: no retornar punteros salvo que sea estrictamente necesario (estructuras grandes, opcionalidad semántica, mutación intencional). Preferir valores.
9. **Interfaces**: nunca retornar interfaces. Las interfaces se *reciben* como parámetros; las funciones *retornan* tipos concretos.
10. **Profundidad de carpetas**: máximo un segundo nivel de profundidad dentro de `internal/` y `pkg/` (ej. `internal/user/` sí; `internal/user/domain/model/` no).
11. **`internal/` es solo dominio y proveedores**: dentro de `internal/` van únicamente módulos de dominio, handlers web, bases de datos y otros proveedores (clientes HTTP, mensajería, etc.). Toda la lógica de aplicación del binario — construir e iniciar servidores, registrar rutas, configurar logging, cargar configuración — vive al nivel del `main.go` en la carpeta del servicio bajo `cmd/`, nunca en `internal/`. No crear paquetes `internal/server`, `internal/config`, `internal/app` ni similares.

## Workflow de desarrollo

0. **Consultar la documentación técnica del proyecto (obligatorio)**: antes de escribir, modificar o refactorizar código, consultar primero el índice de componentes (`sdlc/docs/component_index.md`) para ubicar los componentes existentes en sus archivos reales, y luego la arquitectura (`sdlc/docs/architecture.md`) para entender contenedores, componentes y relaciones. Solo inspeccionar el código directamente cuando estos documentos no existan, estén desactualizados o generen duda. Si un componente citado como existente no se encuentra, no asumir: preguntar al usuario.
1. **Modelar el dominio**: definir entidades, value objects y errores de dominio en `internal/<dominio>/`. Sin dependencias de infraestructura ni frameworks.
2. **Definir contratos**: declarar en el dominio/servicio las interfaces de repositorio y de servicios externos que se necesitan (el consumidor define la interfaz).
3. **Escribir tests primero**: para cada caso de uso, crear el test con testify antes de implementar. Mockear solo extremos con testify/mock.
4. **Implementar servicios**: casos de uso que orquestan dominio y repositorios; reciben dependencias como interfaces por constructor.
5. **Cablear transporte y persistencia**: handlers (HTTP, gRPC, CLI) al final; repositorios concretos implementando las interfaces del dominio.
6. **Componer en `cmd/`**: inyección de dependencias manual en `main`; el main conoce todas las capas, ninguna capa conoce al main. La lógica de aplicación del binario se reparte en archivos junto al `main.go` de cada servicio: `config.go` (carga y validación de configuración), `routes.go` (registro de rutas del ServeMux), `server.go` (construcción del `http.Server`, middleware, ciclo de vida) — no en paquetes de `internal/`.
7. **Verificar**: ejecutar `go build ./...`, `go vet ./...` y `go test ./...` y mostrar los resultados.

## Estructura de referencia

```text
api/                         # OpenAPI/Swagger specs, JSON schema files, protocol definition files.
build/
└── <app>/
    └── Dockerfile           # composición e inyección de dependencias
cmd/
└── <app>/
    ├── main.go              # composición e inyección de dependencias
    ├── config.go            # carga y validación de configuración (env vars, defaults, Validate())
    ├── routes.go            # registro de rutas del ServeMux (patterns Go 1.27+)
    └── server.go            # construcción del http.Server, middleware (logging, recover), ciclo de vida
configs/                     # Configuration file templates or default configs.
deployments/                 # IaaS, PaaS, system and container orchestration deployment configurations and templates (docker-compose, kubernetes/helm, mesos, terraform, bosh).
docs/                        # Design and user documents (in addition to your godoc generated documentation).
internal/
└── <dominio>/               # un paquete por dominio (screaming)
    ├── entity.go            # entidades y value objects
    ├── service.go           # casos de uso (recibe interfaces)
    ├── gateways.go          # interfaces de salida del dominio (las define el dominio)
    ├── handler.go           # transporte (HTTP/gRPC), llama solo al servicio
    ├── <technology>.go      # implementación de persistencia para la tecnología seleccionada
    ├── <technology>.go      # implementación de mensajería para la tecnología seleccionada
    └── service_test.go      # TDD: caja negra, mocks en los extremos
pkg/
└── <utilidad>/              # código reutilizable público (2.º nivel máx.)
scripts/                     # Scripts to perform various build, install, analysis, etc operations. These scripts keep the root level Makefile small and simple.
migrations/                  # Db migrations code
web/                         # Web application specific components: static web assets, server side templates and SPAs.
Makefile                     # Projects commands: build, test, run, etc.
```

Adaptar `internal/<dominio>/` a cada dominio; las implementaciones de persistencia y mensajería conviven en el mismo paquete del dominio (respetando la regla de profundidad máxima de segundo nivel dentro de `internal/` y `pkg/`).

## Snippets clave

Interfaz definida por el consumidor, tipo concreto retornado:

```go
// Repository lo define el dominio (consumidor), no la infraestructura.
type Repository interface {
    FindByID(ctx context.Context, id string) (User, error)
}

// NewService recibe interfaces y retorna un tipo concreto (no interfaz, no puntero innecesario).
func NewService(repo Repository) Service {
    return Service{repo: repo}
}
```

Test de caja negra con mock del extremo (testify):

```go
func TestService_GetUser(t *testing.T) {
    repo := mocks.NewRepository(t) // testify/mock: solo el extremo (persistencia)
    repo.On("FindByID", mock.Anything, "u1").Return(User{ID: "u1"}, nil)

    svc := NewService(repo)
    user, err := svc.GetUser(context.Background(), "u1")

    require.NoError(t, err)
    assert.Equal(t, "u1", user.ID)
}
```

## Checklist antes de entregar

- [ ] ¿El dominio se modeló primero y no depende de infraestructura?
- [ ] ¿Ninguna capa salta a otra (handler → persistencia está prohibido)?
- [ ] ¿Toda comunicación entre capas es vía interfaces definidas por el consumidor?
- [ ] ¿Hay tests escritos antes de la implementación, tratando el componente como caja negra, con mocks solo en los extremos (testify/mock)?
- [ ] ¿No se retornan interfaces ni punteros innecesarios?
- [ ] ¿Se prefirió la standard library y se justificó cualquier dependencia externa?
- [ ] ¿La profundidad dentro de `internal/` y `pkg/` no supera el segundo nivel?
- [ ] ¿`internal/` contiene solo dominio y proveedores (sin paquetes `server`/`config`/`app`), y la lógica del binario (servidor, rutas, logging, config) está en archivos junto al `main.go` bajo `cmd/`?
- [ ] ¿Pasan `go build`, `go vet` y `go test`?
