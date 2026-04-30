# Architecture

**Analysis Date:** 2026-04-30

## System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                     CLI Layer (Cobra)                        │
│  `cmd/` — root, add, remove, show, import, export, etc.     │
├──────────────────┬──────────────────┬───────────────────────┤
│  Provider Plugin │   Device Types   │    Visualization      │
│  `pkg/provider/` │  `pkg/devicetypes/`│   `pkg/visual/`     │
└────────┬─────────┴────────┬─────────┴──────────┬────────────┘
         │                  │                     │
         ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Core Domain / Inventory Model                   │
│  `pkg/devicetypes/inventory.go` (Inventory struct)          │
└─────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Persistence Layer                        │
│  `pkg/datastores/` (DeviceStore interface — JSON impl)      │
└─────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Root Command | CLI entry, config loading, provider setup | `cmd/root.go` |
| Init (cmd) | Bootstrap cmd tree, register subcommands | `cmd/init.go` |
| Provider Registry | Register/retrieve provider plugins | `internal/provider/registry.go` |
| Provider Interface | Contract for import/export/transform | `internal/provider/interface.go` |
| DeviceTypes | Hardware type definitions, inventory model, CRUD | `pkg/devicetypes/` |
| Datastores | Inventory persistence (Load/Save) | `pkg/datastores/datastore.go` |
| Config | YAML config singleton, Viper binding | `internal/config/config.go` |
| Visual | Terminal output: tables, trees, rack diagrams, ETL UX | `pkg/visual/` |

## Pattern Overview

**Overall:** Plugin-based CLI with ETL pipeline for hardware inventory management

**Key Characteristics:**
- Provider plugins self-register via Go `init()` + blank imports in `main.go`
- ETL pipeline (Extract→Transform→Load) for import operations
- Shared Inventory model is the central data structure
- Cobra command tree with Viper config/env/flag precedence
- Hardware types loaded from embedded YAML + optional local/remote directories

## Layers

**CLI Layer (`cmd/`):**
- Purpose: Parse commands/flags, orchestrate operations, call domain logic
- Location: `cmd/`
- Contains: Cobra commands organized by verb (add, remove, show, import, export, update, classify, serve)
- Depends on: `internal/provider`, `internal/config`, `pkg/datastores`, `pkg/devicetypes`, `pkg/visual`
- Used by: End users via terminal

**Provider Layer (`pkg/provider/`):**
- Purpose: Vendor-specific import/export/transform logic (CSM, HPCM, Nautobot, Ochami, Redfish)
- Location: `pkg/provider/`
- Contains: Each provider implements `internal/provider.Provider` interface
- Depends on: `internal/provider` (for registration), `pkg/devicetypes` (for inventory model)
- Used by: CLI layer via registry lookup

**Domain Model (`pkg/devicetypes/`):**
- Purpose: Hardware type definitions, inventory CRUD, classification, parent resolution
- Location: `pkg/devicetypes/`
- Contains: Type structs (Device, Rack, Module, Cable, FRU, Location), Inventory methods, YAML loaders, registries
- Depends on: External YAML type definitions (embedded + filesystem)
- Used by: All layers

**Persistence Layer (`pkg/datastores/`):**
- Purpose: Load/Save the Inventory to a backing store
- Location: `pkg/datastores/`
- Contains: `DeviceStore` interface, JSON implementation
- Depends on: `pkg/devicetypes` (Inventory struct)
- Used by: CLI commands, ETL pipeline

**Internal Support (`internal/`):**
- Purpose: App constants, config management, provider interface definition
- Location: `internal/`
- Contains: `core/` (constants, taxonomy), `config/` (YAML config), `provider/` (interface + registry), `util/`
- Used by: `cmd/`, `pkg/provider/`

## Data Flow

### Import (ETL Pipeline)

1. User runs `cani alpha import <provider>` → `cmd/import/import.go`
2. **Extract**: `provider.Import(cmd, args, &inventory)` — provider fetches raw data from external source (files, APIs)
3. **Transform**: `provider.Transform(existingInventory)` → `*devicetypes.TransformResult` — converts raw data to CANI format
4. **Merge**: Result racks/devices/modules/cables/frus merged into `ctx.inventory`
5. **Load**: `datastores.Datastore.Save(ctx.inventory)` — persist to JSON file

### Add (Direct CRUD)

1. User runs `cani alpha add <slug>` → `cmd/add/add.go`
2. Slug resolved against registries (device, rack, module, cable) → `pkg/devicetypes/registry.go`
3. New instance constructed from type definition → `pkg/devicetypes/constructors.go`
4. Parent assignment (auto or manual) → `pkg/devicetypes/suggest_parent.go`
5. Inventory loaded, item added, inventory saved → `pkg/datastores/`

### Show (Query)

1. User runs `cani alpha show` → `cmd/show/show.go`
2. Inventory loaded from datastore → `pkg/datastores/`
3. Items formatted as table/tree/JSON/rack-view → `pkg/visual/`

### Provider Registration (startup)

1. `main.go` blank-imports all provider packages (`pkg/provider/csm`, etc.)
2. Each provider's `init()` calls `provider.Register("name", instance)` → `internal/provider/registry.go`
3. `cmd.Init()` builds command tree, decorates import/export with provider subcommands
4. `setupDomain` (PersistentPreRunE) loads config, device types, and configures providers

**State Management:**
- `devicetypes.Inventory` is the single source of truth (in-memory during execution)
- Persisted as JSON file at `~/.cani/canidb.json`
- Config persisted as YAML at `~/.cani/cani.yml`
- Global singleton: `config.Cfg`, `datastores.Datastore`

## Key Abstractions

**Provider Interface:**
- Purpose: Pluggable vendor adapters for import/export
- Examples: `pkg/provider/csm/`, `pkg/provider/hpcm/`, `pkg/provider/nautobot/`
- Pattern: Interface + registry + init() self-registration

**Inventory:**
- Purpose: Central data model holding all hardware items indexed by UUID
- Examples: `pkg/devicetypes/inventory.go`
- Pattern: Struct with map[UUID]*Type fields, methods for CRUD/merge/query

**Hardware Type Registries:**
- Purpose: Lookup tables mapping slugs/part-numbers to type definitions
- Examples: `pkg/devicetypes/registry.go`
- Pattern: Package-level maps populated by YAML loaders at startup

**DeviceStore:**
- Purpose: Abstraction over inventory persistence
- Examples: `pkg/datastores/json.go`
- Pattern: Interface with Load/Save methods; implementation selected by flag

## Entry Points

**`main.go`:**
- Location: `main.go`
- Triggers: Binary execution
- Responsibilities: Import providers via blank imports, call `cmd.Init()`, call `cmd.Execute()`

**`cmd.Init()`:**
- Location: `cmd/init.go`
- Triggers: Called from `main.init()`
- Responsibilities: Build entire Cobra command tree, attach provider subcommands to import/export

**`setupDomain`:**
- Location: `cmd/root.go:97`
- Triggers: PersistentPreRunE on root command (runs before any subcommand)
- Responsibilities: Load Viper config, load device types, configure providers

## Architectural Constraints

- **Threading:** Single-threaded CLI; no goroutines or concurrency in core paths
- **Global state:** `config.Cfg` singleton (`internal/config/config.go`), `datastores.Datastore` singleton (`pkg/datastores/datastore.go`), package-level type registries (`pkg/devicetypes/registry.go`)
- **Circular imports:** Broken via callback setters (e.g., `import_.SetProviderGetter()` in `pkg/provider/csm/init.go`)
- **Plugin loading:** Compile-time only via blank imports; no dynamic loading

## Anti-Patterns

### Global Mutable Singletons

**What happens:** `config.Cfg`, `datastores.Datastore`, and type registries are package-level vars mutated at runtime
**Why it's wrong:** Makes testing harder, creates hidden dependencies, prevents concurrent use
**Do this instead:** Pass dependencies explicitly through function parameters or use a context/container struct

### Import Cycle Workarounds

**What happens:** `pkg/provider/csm/init.go` uses `SetProviderGetter` callbacks to break import cycles between transform and provider packages
**Why it's wrong:** Adds indirection, makes data flow harder to trace
**Do this instead:** Restructure so transform functions accept data parameters rather than reaching back to the provider

## Error Handling

**Strategy:** Return `error` up the call stack; Cobra handles top-level display

**Patterns:**
- `fmt.Errorf("context: %w", err)` wrapping throughout
- Sentinel errors in `pkg/datastores/` (e.g., `ErrHardwareNotFound`)
- Provider errors surfaced to user via ETL pipeline

## Cross-Cutting Concerns

**Logging:** `log` stdlib package with `[prefix]` pattern (e.g., `[datastores]`); debug output gated by `config.Cfg.Debug`
**Validation:** Schema version checks in datastores; slug resolution in add commands; `--strict` flag controls whether unclassified devices are errors
**Authentication:** Provider-specific (e.g., CSM/Nautobot use API URLs/tokens from config)
**Configuration Precedence:** CLI flags > environment variables (CANI_ prefix) > config file > defaults (Viper)

---

*Architecture analysis: 2026-04-30*
