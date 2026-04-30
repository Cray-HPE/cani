# Testing Patterns

**Analysis Date:** 2026-04-30

## Test Framework

**Runner:**
- Go standard `testing` package (no third-party test frameworks)
- No testify, gomega, or other assertion libraries

**Assertion Library:**
- Standard library only: `t.Fatal()`, `t.Fatalf()`, `t.Error()`, `t.Errorf()`

**Run Commands:**
```bash
make utest              # Run unit tests (go test -cover ./...)
make ftest              # Run functional tests (shell-based)
make itest              # Run integration tests (shell scripts via spec/)
make etest              # Run edge-case tests (disabled by default, slow)
make test               # Run all: utest + ftest + itest
go test ./...           # Direct Go test invocation
go test -cover ./...    # With coverage
go test -run TestName   # Single test
```

## Test File Organization

**Location:**
- Co-located with source files (same package, same directory)
- File naming: `<source_file>_test.go` or `<feature>_test.go`

**Naming:**
- `inventory_queries_test.go` tests functions in `inventory_queries.go`
- `constructors_test.go` tests functions in `constructors.go`

**Structure:**
```
pkg/devicetypes/
├── inventory.go
├── inventory_queries.go
├── inventory_queries_test.go      # 990 lines
├── inventory_relationships.go
├── inventory_relationships_test.go # 854 lines
├── constructors.go
├── constructors_test.go            # 494 lines
└── ...

internal/util/nameexpand/
├── expand.go
├── expand_test.go
├── sequence.go
├── sequence_test.go
└── ...
```

## Test Structure

**Suite Organization:**
```go
// Table-driven test (primary pattern)
func TestExpand_NumericRanges(t *testing.T) {
    tests := []struct {
        name    string
        pattern string
        want    []string
    }{
        {
            name:    "simple numeric",
            pattern: "x370{1..4}",
            want:    []string{"x3701", "x3702", "x3703", "x3704"},
        },
        // ... more cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Expand(tt.pattern)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Expand(%q) = %v, want %v", tt.pattern, got, tt.want)
            }
        })
    }
}

// Individual test functions (common for simple cases)
func TestFindLocationByNameFound(t *testing.T) {
    inv := NewInventory()
    id := uuid.New()
    inv.Locations[id] = &CaniLocationType{ID: id, Name: "site-alpha"}

    got := inv.FindLocationByName("site-alpha")
    if got == nil {
        t.Fatal("expected location, got nil")
    }
}
```

**Patterns:**
- Table-driven tests with `t.Run()` for parameterized scenarios
- Individual function tests for simple happy-path/failure pairs
- Test names follow `Test<FunctionName><Scenario>` pattern: `TestFindRackByNameFound`, `TestFindRackByNameNotFound`
- Test coverage documentation tables at file top (ASCII format)

**Setup pattern:**
- Direct struct construction in test body (no shared setup functions for most tests)
- `NewInventory()` constructor for creating test fixtures inline
- `t.Helper()` for shared assertion/setup helpers

**Teardown pattern:**
- `defer` for cleanup: `defer func() { Datastore = nil }()`
- `t.Cleanup()` for resource cleanup in import tests

## Mocking

**Framework:** No mocking framework — hand-rolled test doubles

**Patterns:**
```go
// Setup global state, restore after test
func TestSetDeviceStoreJSON(t *testing.T) {
    original := config.Cfg
    config.Cfg = &config.Config{
        Path:      "/tmp/cani-test/config.yaml",
        Datastore: "inventory.json",
    }
    defer func() {
        config.Cfg = original
        Datastore = nil
    }()

    root := &cobra.Command{}
    root.PersistentFlags().String("datastore", "json", "datastore type")

    if err := SetDeviceStore(root, nil); err != nil {
        t.Fatalf("SetDeviceStore() returned unexpected error: %v", err)
    }
}
```

**What to Mock:**
- Global singletons (`config.Cfg`) — save/restore via defer
- External services (Nautobot API) — skip test if unavailable with `t.Skip()`
- Cobra commands — create minimal `*cobra.Command` with required flags

**What NOT to Mock:**
- Core domain logic (`devicetypes.Inventory` methods)
- Pure functions (name expansion, transforms)
- Data structures — construct real instances

## Fixtures and Factories

**Test Data:**
```go
// Inline factory helper (common pattern)
func makeRack(name string, uHeight int) *devicetypes.CaniRackType {
    return &devicetypes.CaniRackType{
        ID:            uuid.New(),
        Name:          name,
        UHeight:       uHeight,
        OccupiedSlots: make(map[int]map[string]uuid.UUID),
    }
}

// Using NewInventory constructor
inv := devicetypes.NewInventory()
nodeID := uuid.New()
inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
    ID:   nodeID,
    Name: "node-fallback",
    Type: devicetypes.Type("node"),
}
```

**Location:**
- `testdata/fixtures/` — JSON/YAML/CSV fixtures for import/export tests
  - `testdata/fixtures/ochami/` — Ochami provider fixtures
  - `testdata/fixtures/redfish/` — Redfish provider fixtures
  - `testdata/fixtures/nautobot/` — Nautobot docker-compose and config
  - `testdata/fixtures/example/` — CSV example fixtures

**Loading fixtures:**
```go
func fixtureDir() string {
    _, thisFile, _, _ := runtime.Caller(0)
    return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "testdata", "fixtures", "cani")
}
```

## Coverage

**Requirements:** No enforced threshold; `-cover` flag used during unit tests

**View Coverage:**
```bash
go test -cover ./...                    # Summary per package
go test -coverprofile=coverage.out ./... # Generate profile
go tool cover -html=coverage.out        # HTML report
```

## Test Types

**Unit Tests (`make utest`):**
- Pure Go tests via `go test -cover ./...`
- 93 test files, ~23,000 lines total
- Heaviest coverage in `pkg/devicetypes/` (core domain)
- Standard `testing` package, no external dependencies required

**Integration Tests (`make itest`):**
- Shell-based scripts in `spec/integration/`
- Uses [ShellSpec](https://shellspec.info/)-style `Describe`/`It`/`When call`/`The` BDD syntax
- Requires running services (SLS simulator, etc.)
- Invokes compiled binary: `bin/cani alpha --config "$CANI_CONF" ...`
- Validates exit codes and stdout/stderr content

**Functional Tests (`make ftest`):**
- Shell-based (currently minimal/placeholder in Makefile)

**Edge-Case Tests (`make etest`):**
- Shell scripts in `spec/edge/`
- Disabled by default (slow, CSM EOL)
- Tests boundary conditions like subnet/VLAN limits

**Go Integration Tests (build-tag guarded):**
- `pkg/provider/nautobot/export/export_integration_test.go`
- Skip pattern: `skipUnlessNautobot(t)` — skips when external service unreachable
- Uses `sync.Once` for expensive one-time setup (loading device type library)

## Common Patterns

**Async Testing:**
```go
// Not applicable — tests are synchronous
```

**Error Testing:**
```go
// Check error returned
func TestSetDeviceStoreUnsupported(t *testing.T) {
    defer func() { Datastore = nil }()

    root := &cobra.Command{}
    root.PersistentFlags().String("datastore", "unsupported", "datastore type")

    err := SetDeviceStore(root, nil)
    if err == nil {
        t.Error("SetDeviceStore() expected error for unsupported type, got nil")
    }
}

// Check specific error content
if !strings.Contains(out, want) {
    t.Errorf("output missing %q\ngot:\n%s", want, out)
}
```

**Test Naming Convention:**
- Happy path: `Test<Function>Found`, `Test<Function>HappyPath`, `Test<Function>True`
- Failure path: `Test<Function>NotFound`, `Test<Function>Unknown`, `Test<Function>False`
- Always paired: every tested function has both happy and failure test

**Skip Pattern for External Dependencies:**
```go
func skipUnlessNautobot(t *testing.T) {
    t.Helper()
    req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, testNautobotURL+"/status/", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Skipf("Nautobot not reachable: %v", err)
    }
    resp.Body.Close()
}
```

## CI Pipeline

**Jenkins (`Jenkinsfile.github`):**
- Uses `csm-shared-library` shared library
- Go version extracted from `go.mod`
- Builds on `metal-gcp-builder` agent
- Matrix build (multiple platforms)
- Runs `go test -cover ./...` as part of build

---

*Testing analysis: 2026-04-30*
