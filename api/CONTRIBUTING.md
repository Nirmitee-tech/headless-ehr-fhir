# Contributing to OpenEHR Server

Thank you for your interest in contributing to OpenEHR Server. This document provides guidelines and instructions for contributing to the project.

---

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Code Organization](#code-organization)
- [How to Add a New Domain](#how-to-add-a-new-domain)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)

---

## Development Environment Setup

### Prerequisites

- **Go 1.22+** -- [Install Go](https://go.dev/doc/install)
- **Docker and Docker Compose** -- Required for PostgreSQL, Redis, and Keycloak
- **golangci-lint** -- For linting (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/openehr/ehr.git
cd ehr

# Copy environment configuration
cp .env.example .env

# Start infrastructure services
docker compose up -d postgres redis keycloak

# Install Go dependencies
go mod download

# Build the server
make build

# Run database migrations
make migrate-up

# Run the server in development mode
make dev

# Verify it is running
curl http://localhost:8000/health
```

### Useful Make Targets

| Command | Description |
|---------|-------------|
| `make build` | Compile the binary to `./bin/ehr-server` |
| `make run` | Build and run the server |
| `make dev` | Build and run in development mode |
| `make test` | Run all tests with verbose output |
| `make test-short` | Run tests excluding long-running tests |
| `make lint` | Run golangci-lint |
| `make tidy` | Run `go mod tidy` |
| `make docker-up` | Start all Docker Compose services |
| `make docker-down` | Stop all Docker Compose services |
| `make docker-build` | Build the Docker image |
| `make migrate-up` | Apply pending database migrations |
| `make tenant-create NAME=<id>` | Create a new tenant schema |
| `make clean` | Remove build artifacts |

### Environment Variables

The server reads configuration from a `.env` file in the project root. See `.env.example` for all available variables. The most important ones for development:

- `DATABASE_URL` -- Points to your local PostgreSQL instance (default: `postgres://ehr:changeme@localhost:5433/ehr?sslmode=disable`)
- `ENV` -- Set to `development` to disable JWT requirement
- `PORT` -- Server listen port (default: `8000`)

---

## Code Organization

### Project Layout

The project follows Go conventions with an emphasis on domain-driven design:

```
internal/
  platform/       # Cross-cutting infrastructure
    auth/          # Authentication, RBAC, SMART on FHIR
    db/            # Connection pool, tenant middleware, migrations
    fhir/          # FHIR resource types, bundles, CapabilityStatement
    hipaa/         # Audit logging, PHI encryption
    middleware/    # HTTP middleware (logging, recovery, request ID, audit)
    plugin/        # Plugin registry for custom domain extensions
  domain/          # Business domains (19 total)
    identity/      # Example: patients, practitioners
    clinical/      # Example: conditions, observations
    ...
```

### The 5-File Domain Pattern

Every domain follows a consistent structure of five files. This makes the codebase predictable and easy to navigate:

**1. `model.go` -- Domain Types**

Contains struct definitions that map to database tables and methods to convert to FHIR representations.

```go
// model.go
type Patient struct {
    ID        uuid.UUID `db:"id" json:"id"`
    FHIRID    string    `db:"fhir_id" json:"fhir_id"`
    FirstName string    `db:"first_name" json:"first_name"`
    LastName  string    `db:"last_name" json:"last_name"`
    // ...
}

func (p *Patient) ToFHIR() map[string]interface{} {
    // Convert to FHIR R4 JSON representation
}
```

**2. `repo.go` -- Repository Interface**

Defines the persistence contract as a Go interface. This is what the service layer depends on, enabling mock-based testing.

```go
// repo.go
type PatientRepository interface {
    Create(ctx context.Context, p *Patient) error
    GetByID(ctx context.Context, id uuid.UUID) (*Patient, error)
    Update(ctx context.Context, p *Patient) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*Patient, int, error)
    Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Patient, int, error)
}
```

**3. `repo_pg.go` -- PostgreSQL Implementation**

Implements the repository interface using pgx queries. All queries use the tenant-scoped connection from context.

```go
// repo_pg.go
type patientRepoPG struct {
    pool *pgxpool.Pool
}

func NewPatientRepo(pool *pgxpool.Pool) PatientRepository {
    return &patientRepoPG{pool: pool}
}

func (r *patientRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Patient, error) {
    conn := db.ConnFromContext(ctx)
    // Execute query using tenant-scoped connection
}
```

**4. `service.go` -- Business Logic**

Contains validation, business rules, and orchestration. Depends only on repository interfaces, not implementations.

```go
// service.go
type Service struct {
    patients PatientRepository
}

func NewService(patients PatientRepository) *Service {
    return &Service{patients: patients}
}

func (s *Service) CreatePatient(ctx context.Context, p *Patient) error {
    if p.FirstName == "" || p.LastName == "" {
        return fmt.Errorf("first_name and last_name are required")
    }
    p.Active = true
    return s.patients.Create(ctx, p)
}
```

**5. `handler.go` -- HTTP Handlers**

Registers routes and translates between HTTP and service calls. Handles both operational REST and FHIR endpoints.

```go
// handler.go
type Handler struct {
    svc *Service
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
    patients := api.Group("/patients")
    patients.POST("", h.CreatePatient)
    patients.GET("/:id", h.GetPatient)
    // ...

    fhirGroup.GET("/Patient", h.SearchPatientsFHIR)
    fhirGroup.GET("/Patient/:id", h.GetPatientFHIR)
    // ...
}
```

---

## How to Add a New Domain

Follow these steps to add a new clinical domain to the system.

### Step 1: Create the Domain Directory

```bash
mkdir internal/domain/mydomain
```

### Step 2: Define the Model (`model.go`)

Create your domain types with database tags, JSON tags, and FHIR conversion methods if applicable:

```go
package mydomain

import (
    "time"
    "github.com/google/uuid"
)

type MyResource struct {
    ID        uuid.UUID `db:"id" json:"id"`
    FHIRID    string    `db:"fhir_id" json:"fhir_id"`
    Name      string    `db:"name" json:"name"`
    Status    string    `db:"status" json:"status"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
```

### Step 3: Define the Repository Interface (`repo.go`)

```go
package mydomain

import (
    "context"
    "github.com/google/uuid"
)

type MyResourceRepository interface {
    Create(ctx context.Context, r *MyResource) error
    GetByID(ctx context.Context, id uuid.UUID) (*MyResource, error)
    Update(ctx context.Context, r *MyResource) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*MyResource, int, error)
}
```

### Step 4: Implement the Repository (`repo_pg.go`)

```go
package mydomain

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/ehr/ehr/internal/platform/db"
)

type myResourceRepoPG struct {
    pool *pgxpool.Pool
}

func NewMyResourceRepo(pool *pgxpool.Pool) MyResourceRepository {
    return &myResourceRepoPG{pool: pool}
}

func (r *myResourceRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MyResource, error) {
    conn := db.ConnFromContext(ctx)
    // Use conn for tenant-scoped queries
}
```

### Step 5: Implement the Service (`service.go`)

```go
package mydomain

type Service struct {
    resources MyResourceRepository
}

func NewService(resources MyResourceRepository) *Service {
    return &Service{resources: resources}
}
```

### Step 6: Implement the Handler (`handler.go`)

```go
package mydomain

import "github.com/labstack/echo/v4"

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
    g := api.Group("/my-resources")
    g.POST("", h.Create)
    g.GET("", h.List)
    g.GET("/:id", h.GetByID)
    g.PUT("/:id", h.Update)
    g.DELETE("/:id", h.Delete)
}
```

### Step 7: Write a Database Migration

Create a new migration file in the `migrations/` directory following the naming convention:

```
migrations/NNN_tX_mydomain_tables.sql
```

Where `NNN` is the next available sequence number and `tX` is the tier designation (t0-t4).

### Step 8: Wire It Up in `main.go`

Add the domain import and registration in `cmd/ehr-server/main.go`:

```go
import "github.com/ehr/ehr/internal/domain/mydomain"

// In runServer():
myRepo := mydomain.NewMyResourceRepo(pool)
mySvc := mydomain.NewService(myRepo)
myHandler := mydomain.NewHandler(mySvc)
myHandler.RegisterRoutes(apiV1, fhirGroup)
```

### Step 9: Write Tests

Create `service_test.go` with mock repositories (see Testing Guidelines below).

---

## Testing Guidelines

### Test Structure

Tests are organized at three levels:

1. **Unit tests** -- Located alongside source files (`*_test.go`). Use mock repositories to test service logic in isolation.
2. **Integration tests** -- Located in `test/integration/`. Test against a real PostgreSQL database.
3. **End-to-end tests** -- Located in `test/e2e/`. Test the full HTTP API.

### Writing Unit Tests with Mock Repositories

The recommended approach is to create in-memory mock implementations of your repository interfaces:

```go
// service_test.go
package mydomain

import (
    "context"
    "testing"
    "github.com/google/uuid"
)

type mockMyResourceRepo struct {
    resources map[uuid.UUID]*MyResource
}

func newMockRepo() *mockMyResourceRepo {
    return &mockMyResourceRepo{
        resources: make(map[uuid.UUID]*MyResource),
    }
}

func (m *mockMyResourceRepo) Create(_ context.Context, r *MyResource) error {
    r.ID = uuid.New()
    m.resources[r.ID] = r
    return nil
}

func (m *mockMyResourceRepo) GetByID(_ context.Context, id uuid.UUID) (*MyResource, error) {
    r, ok := m.resources[id]
    if !ok {
        return nil, fmt.Errorf("not found")
    }
    return r, nil
}

// Implement remaining interface methods...

func TestCreateMyResource(t *testing.T) {
    svc := NewService(newMockRepo())
    r := &MyResource{Name: "Test"}
    err := svc.Create(context.Background(), r)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if r.ID == uuid.Nil {
        t.Error("expected ID to be set")
    }
}
```

### Test Conventions

- Use the standard `testing` package. Do not introduce external assertion libraries without discussion.
- Name test functions descriptively: `TestCreatePatient_NameRequired`, `TestDeletePatient`.
- Use `t.Fatalf` for setup failures that prevent the test from continuing. Use `t.Errorf` for assertion failures that should not stop the test.
- Test both success and error paths.
- Validate returned errors contain meaningful messages.
- For handler tests, use `httptest.NewRecorder()` and `echo.New()` to test HTTP behavior.

### Running Tests

```bash
# Run all tests
make test

# Run tests for a specific domain
go test ./internal/domain/identity/... -v

# Run short tests only (skip integration tests)
make test-short

# Run with race detector
go test -race ./...
```

---

## Pull Request Process

1. **Fork and branch.** Create a feature branch from `main`:
   ```bash
   git checkout -b feature/my-domain
   ```

2. **Follow conventions.** Use the 5-file domain pattern. Keep handler logic thin and push business rules into the service layer.

3. **Write tests.** All new domains must include unit tests for the service layer at minimum. Aim for coverage of all validation paths and key business logic.

4. **Lint and test locally.** Ensure your changes pass before pushing:
   ```bash
   make lint
   make test
   ```

5. **Write a clear commit message.** Use conventional-commit style prefixes:
   - `feat:` for new features or domains
   - `fix:` for bug fixes
   - `refactor:` for code restructuring
   - `docs:` for documentation changes
   - `test:` for test additions or fixes

6. **Open a pull request.** Provide:
   - A summary of what changed and why
   - Migration details if database schema changes are included
   - Any new environment variables or configuration required
   - Test evidence (test output, curl examples, etc.)

7. **Address review feedback.** Respond to all review comments. Push additional commits rather than force-pushing amended commits so reviewers can see incremental changes.

---

## Code Style

- Follow standard Go formatting. Run `gofmt` (or `goimports`) on all files before committing.
- Run `golangci-lint run ./...` and resolve all warnings.
- Use meaningful variable and function names. Avoid single-letter variables except for loop indices and short lambda parameters.
- Keep functions focused. If a function exceeds ~60 lines, consider extracting helper functions.
- Document all exported types and functions with Go doc comments.
- Use `context.Context` as the first parameter for any function that performs I/O or may be cancelled.
- Return `error` as the last return value. Wrap errors with `fmt.Errorf("context: %w", err)` to preserve the error chain.
- Use structured logging via `zerolog`. Do not use `fmt.Println` or `log.Println` in production code.
- Repository implementations must use `db.ConnFromContext(ctx)` to get the tenant-scoped database connection.
- FHIR endpoints must return `OperationOutcome` resources for errors (use `fhir.ErrorOutcome()`, `fhir.NotFoundOutcome()`).
- Operational REST endpoints must use `echo.NewHTTPError()` for errors and `pagination.NewResponse()` for list results.
