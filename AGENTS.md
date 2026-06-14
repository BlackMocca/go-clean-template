# AGENTS.md — Go Clean Architecture Example

Reference implementation of Design-First REST API using Go + Echo + oapi-codegen with Clean Architecture principles.

---

## 1. Project Structure

```
example/
├── main.go                      ← Entry point: config → DB → migrate → wire layers → server
│
├── route/
│   └── web.go                   ← Manual route registration (public vs protected)
│
├── middleware/
│   ├── apikey.go                ← API Key auth (header X-API-Key + query ?apikey= fallback)
│   ├── middleware.go            ← Middleware interface + client struct
│   ├── openapi.go               ← OpenAPI schema validation middleware
│   └── transform_response.go    ← Converts Response types to JSON HTTP responses
│
├── service/
│   ├── web/                     ← Generated code (oapi-codegen) + handler implementation
│   │   ├── generate.go          ← //go:generate oapi-codegen for Echo server
│   │   ├── server.cfg.yaml      ← oapi-codegen config
│   │   ├── server.gen.go        ← Generated: ServerInterface, types, wrappers, GetSwagger()
│   │   └── handlers.go          ← Implements ServerInterface → delegates to usecase
│   │   └── handlers_test.go     ← Handler unit tests (httptest + mock Usecase)
│   │   └── mocks/               ← Auto-generated mocks for web dependencies (if any)
│   │
│   └── user/                    ← Domain layer: interfaces, usecase, repository
│       ├── types.go             ← Repository & Usecase interfaces + ListParams
│       ├── usecase.go           ← Business logic (email-duplicate check, delete validation)
│       ├── psql_repository.go   ← PostgreSQL implementation via sqlx
│       ├── psql_repository_test.go      ← Repository integration tests
│       ├── psql_repository_suit_test.go ← Test suite setup for repository
│       ├── usecase_test.go      ← Usecase unit tests (mock Repository)
│       └── mocks/               ← Auto-generated mocks via mockery (Repository, Usecase)
│           ├── Repository.go
│           └── Usecase.go
│
├── models/                      ← Domain types (shared across service layers)
│   ├── user.go                  ← User type with utils/models types (Timestamp, EnumScan, JsonScan)
│   ├── permission.go            ← Permission entity + Permissions slice helper
│   └── paginator.go             ← Generic Paginator[T] with offset/limit helpers
│
├── constants/                   ← Generated enum types from OpenAPI spec
│   ├── generate.go              ← //go:generate oapi-codegen for enums only
│   ├── enum.cfg.yaml            ← oapi-codegen config (models only, skip-prune: true)
│   └── enum.gen.go              ← Enum types + .Valid() methods
│
├── migrations/                  ← Database migrations (dual-track pattern)
│   ├── migrate.go               ← Dual-track migration runner (schema + seed data)
│   └── database/postgres/       ← Migration SQL files
│       ├── schema_migrations/   ← 000001–000004: extensions, permissions, users, user_permissions
│       │   ├── 000001_extension.up.sql / .down.sql
│       │   ├── 000002_create_permissions_table.up.sql / .down.sql
│       │   ├── 000003_create_users.up.sql / .down.sql
│       │   └── 000004_create_user_permissions.up.sql / .down.sql
│       └── master/              ← Seed data (up only, no down)
│           └── 000001_permission.up.sql
│
├── spec/openapi/                ← OpenAPI v3 specification (single source of truth!)
│   ├── openapi.yaml             ← Entry point: paths $ref, components, securitySchemes
│   ├── enums.yaml               ← Enum definitions with x-enum-varnames
│   ├── global/components/       ← Shared schemas across all APIs
│   │   ├── parameters/          ← page.yaml, page_size.yaml, document_id.yaml
│   │   ├── schemas/             ← uuid.yaml, timestamp.yaml, paginator.yaml, response_success.yaml, response_error.yaml
│   │   └── securitySchemes/     ← apikey.yaml, jwt.yaml
│   ├── components/              ← Domain-specific (this API only)
│   │   ├── parameters/          ← role.yaml, search_word.yaml
│   │   └── schemas/             ← user.yaml, permission.yml
│   └── paths/                   ← One-file-per-operation (operation objects only!)
│       ├── public/              ← health.yaml (GET /api/health)
│       └── users/               ← CRUD operations
│           ├── list_users.yaml, create_user.yaml, get_user_by_id.yaml, update_user.yaml, delete_user.yaml
│           ├── request/         ← create_user_json_body.yaml, update_user_json_body.yaml
│           └── response/        ← list_users_success.yaml, create_user_success.yaml, etc.
│
├── .mockery.yml                 ← Mockery configuration
├── Makefile                     ← Build/run/generate/test scripts
├── go.mod / go.sum              ← Go module definition
├── Dockerfile                   ← Multi-stage build
└── docker-compose.yml           ← Dev environment (app + postgres)
```

---

## 2. Design-First Workflow

### Step 1: Write OpenAPI Spec

Define all routes, schemas, enums under `spec/openapi/`. This is the **single source of truth**.

**Rules:**
- Use one-file-per-operation pattern: each endpoint has its own YAML file
- Operation files contain **only** operation fields (no top-level HTTP method key like `get:`)
- Use flow style for `$ref`: `{ $ref: './path/to/file.yaml' }` — never `key: $ref: '...'`
- Reference individual parameter files directly (name + in + schema are self-contained)
- Enum definitions use `x-enum-varnames` with format `$EnumName$EnumValue`

**See also**: [spec/openapi/rules.md](./spec/openapi/rules.md) for detailed spec writing rules.

### Step 2: Bundle + Generate Code

```bash
# Bundle all $ref into a single file (required by oapi-codegen)
redocly bundle spec/openapi/openapi.yaml -o spec/openapi/openapi_bundle.yaml

# Generate Go code — two targets:
#   a) Echo server code + types → service/web/server.gen.go
#   b) Enum constants only → constants/enum.gen.go
go generate ./...
```

**Generation pattern:** Each package uses a `generate.go` + `{name}.cfg.yaml` pair:
- `service/web/generate.go` → relative path `../../spec/openapi/openapi_bundle.yaml` (2 levels up)
- `constants/generate.go` → relative path `../spec/openapi/enums.yaml` (1 level up)

### Step 3: Implement Handlers

Build structs that implement the generated `ServerInterface`:

```go
type handler struct {
    userUs user.Usecase   // injected usecase (not repository directly)
}

func NewHandler(userUs user.Usecase) ServerInterface {
    return &handler{userUs: userUs}
}
```

**Handler method rules:**
- Every method returns `error` — but the value is always a `middleware.Response` struct
- Success with data → `return middleware.NewResponseJSON(data)`
- Empty result (not found) → `return middleware.NewResponseNoContent()` (HTTP 204, no body)
- Error → `return middleware.NewResponseError(httpStatus, code, errorType, message, data)`
- Use `ctx.Request().Context()` to pass context to usecase layer
- Convert query params with `fn.ConvertStruct[GeneratedType, DomainType](params)`
- Read request body with `ctx.Bind(&reqBody)` using generated request body type

### Step 4: Register Routes Manually

Separate public and protected routes explicitly:

```go
// Public — no middleware
func (r Route) RegisterPublicRoutes(w web.ServerInterfaceWrapper) {
    r.e.GET("/api/health", w.HealthCheck)
}

// Protected — API Key enforced at route level
func (r Route) RegisterProtectedRoutes(w web.ServerInterfaceWrapper) {
    r.e.GET("/api/v1/users", w.ListUsers, r.middl.APIKeyAuth)
    r.e.POST("/api/v1/users", w.CreateUser, r.middl.APIKeyAuth)
}
```

**Why manual registration over auto?**
- Clear public vs protected separation (auto registration uses `RegisterHandlers` which applies middleware to all routes)
- Per-route middleware control
- Ability to add custom logic between routes

### Step 5: Transform Response Format

All handlers return `middleware.Response` which is transformed by the `TransformResponse` middleware into JSON HTTP responses.

```go
// In main.go — register before route registration
e.Use(middl.TransformResponse)
```

**Response struct:**
```go
type Response struct {
    HttpStatus   int                 `json:"-"`          // controls HTTP status, NOT serialized to JSON
    ResponseCode constants.ApplicationCode  `json:"code"`         // e.g. "C20000"
    ErrorType    constants.ApplicationErrorType `json:"error_type,omitempty"` // e.g. "conflict"
    Message      string              `json:"message"`      // human-readable
    Data         interface{}         `json:"data"`         // payload
}
```

**Response constructors:**
| Constructor | HTTP Status | Code | Use Case |
|-------------|-------------|------|----------|
| `NewResponseJSON(data)` | 200 OK | C20000 | Success with body |
| `NewResponseNoContent()` | 204 No Content | (empty) | Empty result — no JSON body |
| `NewResponseError(400, ...)` | 400 Bad Request | C40000 | Validation error |
| `NewResponseError(401, ...)` | 401 Unauthorized | C40100 | Auth failure |
| `NewResponseError(409, ...)` | 409 Conflict | C409xx | Business conflict |
| `NewResponseError(500, ...)` | 500 Internal Error | C50000 | Unexpected error |

---

## 3. REST API Implementation Rules

### 3.1 Dependency Injection Flow

```
main.go (wiring)
    ↓ NewPsqlRepository(psqlClient)
psql_repository.go (implements Repository interface)
    ↓ NewUserUsecase(repo)
usecase.go (business logic, depends on Repository interface)
    ↓ NewHandler(userUs)
handlers.go (implements ServerInterface, delegates to Usecase)
```

**Rule:** Handlers NEVER depend directly on repositories — always go through the usecase layer. This keeps handlers thin and testable.

### 3.2 Error Handling Pattern

Usecase returns domain-specific errors using `errors.New(string(constants.ApplicationCodeXxx))`. The handler matches these strings to return proper HTTP responses:

```go
err := h.userUs.Save(ctx.Request().Context(), user)
if err != nil {
    if strings.EqualFold(err.Error(), string(constants.ApplicationCodeConflictUserEmailWasDuplicate)) {
        return middleware.NewResponseError(
            http.StatusConflict,
            constants.ApplicationCodeConflictUserEmailWasDuplicate,
            constants.ApplicationErrorConflict,
            fmt.Sprintf("email %s was duplicated", user.Email.ValueOrZero()),
            nil)
    }
    return middleware.NewResponseError(
        http.StatusInternalServerError,
        constants.ApplicationInternalServerError,
        constants.ApplicationErrorInternalServerError,
        err.Error(),
        nil)
}
```

**Rules:**
- Always use `strings.EqualFold` for case-insensitive error comparison (enum values are strings internally)
- Return application-level error codes (`C40901`, not just HTTP 409)
- Include the conflicting value in the message for conflict errors
- Never return raw database errors to clients — wrap in application-level messages

### 3.3 Data Conversion Patterns

Use `fn.CopyJSON` (from BlackMocca utils) for struct-to-struct conversion:

```go
// Generated type → domain model
var user = new(models.User)
if err := fn.CopyJSON(reqBody, user); err != nil { ... }

// Domain model → generated response type
resp := GetUserByIDSuccess{}
fn.CopyJSON(user, &resp.Data)
return middleware.NewResponseJSON(resp.Data)

// Query params → domain params
args, err := fn.ConvertStruct[ListUsersParams, user.ListParams](params)
```

**Rules:**
- Always check errors from `CopyJSON` / `ConvertStruct` and return 500 if conversion fails
- Domain models use BlackMocca utils types: `Timestamp`, `EnumScan[T]`, `JsonScan[T]`
- Response types are generated OpenAPI types — never expose internal domain types in responses

### 3.4 Request Body Handling

```go
var reqBody = UpdateUserByIDJSONRequestBody{}  // generated type
if err := ctx.Bind(&reqBody); err != nil {     // uses oapi-codegen runtime binder
    return middleware.NewResponseError(
        http.StatusBadRequest,
        constants.ApplicationCodeBadRequestInvalidArgument,
        constants.ApplicationErrorInvalidArgument,
        err.Error(),
        reqBody)  // echo back invalid input
}
```

### 3.5 Pagination Handling

Use the generic `Paginator[T]` from models package:

```go
func (p *Paginator[T]) Limit() int64 { return p.PageSize }
func (p *Paginator[T]) Offset() int64 { return (p.Page - 1) * p.PageSize }
func (p *Paginator[T]) SetTotalDocument(total int64) { ... }
```

Repository returns `models.Paginator[[]*models.User]` — handler passes it to `CopyJSON` which maps to the generated `ListUsersSuccess.Data` type.

### 3.6 Transaction Handling in Repository

Every write operation (Save, Delete) must use a transaction:

```go
func (p psqlRepository) Save(ctx context.Context, user *models.User) error {
    tx, err := p.client.GetClient().Beginx()
    if err != nil { return err }
    defer tx.Rollback()  // auto-rollback on any early return
    
    // ... perform operations ...
    
    return tx.Commit()  // only commit at the end
}
```

**Rules:**
- Always use `defer tx.Rollback()` immediately after Beginx
- Only call `tx.Commit()` as the final statement
- Handle errors from each SQL operation and return early (triggers rollback)
- Use `sqlx.Rebind(sqlx.DOLLAR, sql)` for parameterized queries

---

## 4. Unit Testing Rules

### 4.1 Mock Generation via Makefile

Mocks are auto-generated using [mockery](https://github.com/vektra/mockery). Configuration lives in `.mockery.yml`:

```yaml
# .mockery.yml
dir: "{{.InterfaceDir}}/mocks"
filename: "{{.InterfaceName}}.go"
structname: 'Mock{{.InterfaceName}}'
pkgname: '{{.SrcPackageName}}_test'  # _test suffix for test package
template: testify                     # uses github.com/stretchr/testify/mock
packages:
  example:
    config:
      recursive: true
      all: true
```

**Generate mocks:**
```bash
make gen-interface
# or directly: mockery
```

Mocks are generated into the same directory as the interface file, in a `mocks/` subdirectory. For example:
- `service/user/types.go` → interfaces `Repository`, `Usecase`
- Output: `service/user/mocks/Repository.go`, `service/user/mocks/Usecase.go`

### 4.2 Test Execution via Makefile

```bash
make test-unit
# Runs all *_test.go files (excluding mocks, gen files), generates cover.out + coverage.html
```

The Makefile automatically:
1. Finds all directories containing `*_test.go`
2. Excludes mock and generated code directories (`grep -vE 'mock|gen'`)
3. Runs tests with `-coverprofile=cover.out`
4. Generates HTML coverage report in `coverage.html`
5. Prints total test count and coverage percentage

### 4.3 Handler Testing Pattern (httptest + middleware.Response)

Handlers return `error` but the value is always `middleware.Response`. Test pattern:

```go
// Helpers for handler tests
func newRequest(method, url string, bodyJSON ...[]byte) *http.Request {
    if len(bodyJSON) > 0 && bodyJSON[0] != nil {
        r := httptest.NewRequest(method, url, bytes.NewReader(bodyJSON[0]))
        r.Header.Set("Content-Type", "application/json")
        return r
    }
    return httptest.NewRequest(method, url, nil)
}

func extractResponse(t *testing.T, handlerErr error) middleware.Response {
    t.Helper()
    resp, ok := handlerErr.(middleware.Response)
    require.True(t, ok, "expected middleware.Response but got %T", handlerErr)
    return resp
}

func assertFullResponse(t *testing.T, resp middleware.Response, httpStatus int, 
    code constants.ApplicationCode, errType constants.ApplicationErrorType, 
    message string, data interface{}) {
    assert.Equal(t, httpStatus, resp.HttpStatus)
    assert.Equal(t, code, resp.ResponseCode)
    if errType != "" { assert.Equal(t, errType, resp.ErrorType) }
    else { assert.Empty(t, string(resp.ErrorType)) }
    assert.Equal(t, message, resp.Message)
    assert.Equal(t, data, resp.Data)
}
```

**Test structure:**
```go
func TestHandler_ListUsers(t *testing.T) {
    t.Run("success: returns paginated users with all response fields", func(t *testing.T) {
        mockUc := mocks.NewMockUsecase(t)
        e := echo.New()
        handlerInst := web.NewHandler(mockUc)

        // 1. Setup mock expectations (see §4.4 multi-call ordering)
        mockUc.EXPECT().ListUsers(mock.Anything, mock.Anything).
            Run(func(ctx context.Context, params user.ListParams) {
                assert.Equal(t, int64(1), params.Page)
            }).
            Return(paginator, nil).Times(1)

        // 2. Create request and context
        req := newRequest(http.MethodGet, "/api/v1/users?page=1&page_size=10")
        rec := httptest.NewRecorder()
        c := e.NewContext(req, rec)

        // 3. Call handler directly (not via httptest server — faster)
        handlerErr := handlerInst.ListUsers(c, web.ListUsersParams{Page: 1, PageSize: 10})

        // 4. Extract response and assert EVERY field
        resp := extractResponse(t, handlerErr)
        
        assert.Equal(t, http.StatusOK, resp.HttpStatus)
        assert.Equal(t, constants.ApplicationCodeSuccess, resp.ResponseCode)
        assert.Empty(t, string(resp.ErrorType))
        assert.Equal(t, "successful", resp.Message)
        // NOTE: Data is map[string]interface{} from fn.CopyJSON — assert structure as needed
        
        mockUc.AssertNumberOfCalls(t, "ListUsers", 1)  // verify call count
    })
}
```

**Handler test rules:**
- Always use `extractResponse` to cast error → middleware.Response before asserting
- Assert ALL fields: HttpStatus, ResponseCode, ErrorType (if any), Message, Data
- Use `.Run(func(...))` to inspect arguments passed to mocks (context, params structs)
- Use `mockUc.AssertNumberOfCalls(t, "MethodName", N)` to verify call counts
- Test empty results → expect HTTP 204 with no Content-Type header via `NewResponseNoContent()`
- Create a new `echo.New()` for each sub-test to avoid state leakage

### 4.4 Usecase Testing Pattern (mock Repository)

```go
func TestUserUsecase_Save(t *testing.T) {
    t.Run("failure: returns conflict when email is duplicated", func(t *testing.T) {
        mockRepo := mocks.NewMockRepository(t)
        uc := user.NewUserUsecase(mockRepo)

        newUser := buildUser(uuid.MustParse("dddd..."), "taken@example.com", constants.RoleAdmin)

        // Mock call 1: FindByID returns existing user → email conflict
        mockRepo.EXPECT().
            FindByID(mock.Anything, newUser.Email.ValueOrZero()).
            Return(existingUser, nil).Times(1)

        // Execute — Save will NOT call mockRepo.Save because it fails at step 1
        err := uc.Save(context.Background(), newUser)

        require.Error(t, err)
        assert.Equal(t, string(constants.ApplicationCodeConflictUserEmailWasDuplicate), err.Error())
        
        // Verify: FindByID was called once, Save was never called
        mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
        mockRepo.AssertNumberOfCalls(t, "Save", 0)  // crucial!
    })
}
```

**Usecase test rules:**
- Assert error message matches the expected ApplicationCode string exactly
- Verify call ordering and counts — important for multi-step operations (FindByID → Save)
- Use `mock.Anything` for context, but assert specific values in `.Run()` callbacks
- Always check both success and failure branches

### 4.5 Multi-call Mock Ordering

When a usecase method calls the repository multiple times (e.g., Delete: FindByID then Delete), mock expectations are ordered by definition order:

```go
// Delete flow: FindByID → (check) → Delete
mockRepo.EXPECT().FindByID(mock.Anything, id.String()).Return(user, nil).Times(1)
mockRepo.EXPECT().Delete(mock.Anything, id).Return(nil).Times(1)
```

**Rules for multi-call:**
- Define expectations in the exact order they will be called
- Use `.Run(func(...))` to assert parameters on each call
- Verify both expected calls and non-called methods (`AssertNumberOfCalls(t, "Save", 0)`)

### 4.6 Helper Functions Pattern

Extract reusable helpers to keep tests DRY:

```go
// Build test data consistently
func buildUser(id uuid.UUID, email string, role constants.Role) *models.User {
    now := utilModels.NewTimestampFromNow()
    return &models.User{
        ID: id, Name: zeroString("Test User"), Email: zeroString(email),
        Role: utilModels.NewEnumScan[constants.Role](string(role)),
        CreatedAt: now.ToPointer(), UpdatedAt: now.ToPointer(),
    }
}

// Assert every field on domain models
func assertFullUser(t *testing.T, expected, actual *models.User) {
    assert.Equal(t, expected.ID, actual.ID)
    assert.Equal(t, expected.Name, actual.Name)
    // ... assert all fields including Timestamp pointers and Permissions slice
}

func assertFullPaginator(t *testing.T, expected, actual models.Paginator[[]*models.User]) {
    assert.Equal(t, expected.TotalDocument, actual.TotalDocument)
    assert.Len(t, actual.Documents, len(expected.Documents))
    for i := range expected.Documents {
        assertFullUser(t, expected.Documents[i], actual.Documents[i])
    }
}
```

### 4.7 Test Package Naming

All test files use `_test` suffix package name (e.g., `package web_test`, `package user_test`). This ensures tests are external to the production code and only interact through exported APIs — matching real-world usage patterns.

---

## 5. Migrations Rules

### 5.1 Dual-Track Migration Pattern

The migration runner uses **two independent tracks** with separate tracking tables:

| Track | Directory | Tracking Table | Direction | Purpose |
|-------|-----------|----------------|-----------|---------|
| Schema migrations | `database/postgres/schema_migrations/` | `schema_migrations` (per-schema table) | Up + Down | CREATE/DROP/ALTER tables and columns |
| Seed data (master) | `database/postgres/master/` | `seed_migrations` (per-schema table) | Up only | INSERT reference data, defaults, lookups |

**Why dual track?**
- Schema migrations must run **before** seed data (you can't INSERT into a table that doesn't exist yet)
- Seed data uses `ON CONFLICT DO NOTHING` — it's idempotent and safe to rerun
- Schema migrations are fully reversible (.down.sql for each .up.sql)
- Seed data is NOT reversed (no .down.sql files in master/)

### 5.2 Migration File Naming Convention

| Pattern | Example | Description |
|---------|---------|-------------|
| `000001_extension.up.sql` | ✅ | Schema migration — numbered, reversible |
| `000001_extension.down.sql` | ✅ | Down migration paired with .up.sql |
| `000001_permission.up.sql` | ✅ | Seed data — numbered within master/ directory only |
| (no down file) | ✅ | Master migrations never have .down.sql |

### 5.3 Migration Rules

**Schema migrations (.up.sql):**
- Must be reversible with a corresponding .down.sql
- Use `IF NOT EXISTS` for CREATE statements to handle reruns gracefully
- Apply schema changes in dependency order (extensions → tables → indexes)
- Never ALTER a column to drop NOT NULL if existing data violates the constraint

**Seed data (.up.sql):**
- Use `INSERT ... ON CONFLICT DO NOTHING` for idempotency
- Reference enum values that match OpenAPI spec definitions
- Only seed essential reference data (permissions, roles, default configurations)
- Do NOT seed business data — only lookup/reference tables

### 5.4 Migration Runner Configuration

```go
// migrations/migrate.go
type Config struct {
    schema string       // PostgreSQL search_path schema name
    path   string       // absolute path to migration files directory
    client *psql.Client // database connection (closed after migration)
}

func NewConfig(databaseURL, schemaName, migrationsDirPath string) (Config, error)
func (config Config) Migrate() error  // runs schema_migrations then master
```

**Usage in main.go:**
```go
relPath := filepath.Join(".", "migrations", "database", "postgres")
absPath, _ := filepath.Abs(relPath)
cfg, err := migrations.NewConfig(config.DatabaseURL, "public", absPath)
cfg.Migrate()  // panics on failure (STOP_ON_MIGRATION_FAILED=true by default)
```

### 5.5 Migration Rules Checklist

- [ ] Schema migration files use `_up.sql` / `_down.sql` naming convention
- [ ] Seed data uses only `.up.sql` in `master/` directory
- [ ] Seed data uses `ON CONFLICT DO NOTHING` for safe reruns
- [ ] Migrations are ordered by dependency (extensions first, then tables)
- [ ] Search path is set per-migration: `SET search_path TO "public"`
- [ ] Each track uses its own migrations table: `schema_migrations` vs `seed_migrations`

---

## 6. Key Dependencies Reference

| Package | Purpose | Usage Location |
|---------|---------|----------------|
| `github.com/labstack/echo/v4` | HTTP framework, routing, middleware | main.go, route/, handlers.go |
| `github.com/oapi-codegen/runtime` | OpenAPI spec compliance types (Email, UUID) | server.gen.go |
| `github.com/BlackMocca/utils/connectors/psql` | Pooled PostgreSQL client (pgx + sqlx) | psql_repository.go, main.go |
| `github.com/BlackMocca/utils/fn` | `ConvertStruct`, `CopyJSON` for type conversion | handlers.go |
| `github.com/BlackMocca/utils/models` | `Timestamp`, `EnumScan[T]`, `JsonScan[T]` types | models/user.go, usecase_test.go |
| `github.com/golang-migrate/migrate/v4` | Database migrations (dual-track) | migrations/migrate.go |
| `github.com/google/uuid` | UUID generation/parsing | handlers.go, psql_repository.go |
| `github.com/guregu/null/v6` | Nullable types (`zero.String`) for DB columns | models/user.go |
| `github.com/stretchr/testify` | Mocking and assertions | *_test.go files |
| `github.com/vektra/mockery` (mockery) | Auto-generates mock structs from interfaces | .mockery.yml |

---

## 7. Quick Reference Commands

```bash
# Generate code from OpenAPI spec
redocly bundle spec/openapi/openapi.yaml -o spec/openapi/openapi_bundle.yaml
go generate ./...

# Generate mocks for test dependencies
make gen-interface     # or: mockery

# Run unit tests with coverage
make test-unit         # outputs cover.out + coverage.html

# Build and run
go build -o example . && ./example

# Docker (dev environment)
docker compose up --build
```
