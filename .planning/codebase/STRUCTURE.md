# Codebase Structure

**Analysis Date:** 2026-04-30

## Directory Layout

```
cani/
├── main.go                 # Entry point — blank imports providers, calls cmd.Init()/Execute()
├── go.mod                  # Go module: github.com/Cray-HPE/cani (Go 1.25)
├── go.sum                  # Dependency checksums
├── Makefile                # Build/test automation
├── cmd/                    # CLI commands (Cobra)
│   ├── init.go             # Builds command tree, wires providers
│   ├── root.go             # Root command, Viper setup, setupDomain
│   ├── version.go          # Version info
│   ├── add/                # `add` verb — rack, device, module, cable, location, metadata
│   ├── alpha/              # Parent command grouping alpha-stage features
│   ├── classify/           # `classify` verb — assign device types interactively
│   ├── export/             # `export` verb — export inventory to external systems
│   ├── import/             # `import` verb — ETL pipeline (extract/transform/load)
│   ├── init/               # `init` verb — initialize provider session
│   ├── remove/             # `remove` verb — remove items from inventory
│   ├── serve/              # `serve` verb — API server (not yet implemented)
│   ├── show/               # `show` verb — display inventory (table/tree/json/rack)
│   └── update/             # `update` verb — modify existing items
├── internal/               # Private packages (not importable externally)
│   ├── config/             # YAML config management (singleton)
│   ├── core/               # App constants, taxonomy helpers
│   ├── provider/           # Provider interface + registry
│   └── util/               # Internal utilities
├── pkg/                    # Public packages
│   ├── canu/               # CANU integration utilities
│   ├── datastores/         # Inventory persistence (DeviceStore interface, JSON impl)
│   ├── devicetypes/        # Core domain: type definitions, inventory model, CRUD, loaders
│   │   ├── device-types/   # Embedded YAML device definitions (by vendor)
│   │   ├── module-types/   # Embedded YAML module definitions
│   │   ├── rack-types/     # Embedded YAML rack definitions
│   │   ├── cable-types/    # Embedded YAML cable definitions
│   │   ├── location-types/ # Embedded YAML location definitions
│   │   └── connections/    # Connection/topology logic
│   ├── nautobot/           # Nautobot API client (generated)
│   ├── provider/           # Provider implementations
│   │   ├── csm/            # Cray System Management provider
│   │   ├── example/        # Example/template provider
│   │   ├── hpcm/           # HPC Manager provider
│   │   ├── nautobot/       # Nautobot DCIM provider
│   │   ├── ochami/         # OpenCHAMI provider
│   │   └── redfish/        # Redfish/BMC provider
│   ├── utils/              # Shared utility functions
│   ├── visual/             # Terminal output formatting (tables, trees, rack views, ETL UX)
│   └── xname/              # XName (location path) utilities
├── testdata/               # Test fixtures and sample data
├── spec/                   # OpenAPI/spec files
├── docs/                   # Documentation (mkdocs)
├── tools/                  # Build/dev tooling
├── bin/                    # Built binaries
└── vendor/                 # Vendored Go dependencies
```

## Directory Purposes

**`cmd/`:**
- Purpose: All CLI command definitions (one subdirectory per verb)
- Contains: Cobra command constructors, flag definitions, RunE handlers
- Key files: `cmd/init.go` (tree assembly), `cmd/root.go` (entry logic)

**`internal/provider/`:**
- Purpose: Define the provider contract and global registry
- Contains: `interface.go` (Provider, Exporter, Importer, DeviceStager, etc.), `registry.go` (Register/GetProvider)
- Key files: `internal/provider/interface.go`

**`internal/config/`:**
- Purpose: Application configuration (YAML file, Viper integration)
- Contains: Config struct, Load/Save, migration
- Key files: `internal/config/config.go`

**`pkg/devicetypes/`:**
- Purpose: Central domain package — hardware type library and inventory model
- Contains: Type structs, inventory CRUD, YAML loaders, registries, classification, parent suggestion
- Key files: `inventory.go`, `registry.go`, `loader.go`, `constructors.go`, `classify.go`

**`pkg/datastores/`:**
- Purpose: Persistence abstraction layer
- Contains: `DeviceStore` interface, JSON file implementation, schema migration
- Key files: `datastore.go`, `json.go`, `migrate.go`

**`pkg/provider/`:**
- Purpose: Concrete provider implementations for different HPC/infrastructure platforms
- Contains: One subdirectory per provider, each with init.go (registration), provider.go, import/, export/, transform/
- Key files: `pkg/provider/csm/init.go`, `pkg/provider/hpcm/provider.go`

**`pkg/visual/`:**
- Purpose: All terminal output formatting and user interaction
- Contains: Table rendering, tree building, rack visualization, ETL step UX, color utilities
- Key files: `table.go`, `tree.go`, `rack.go`, `etl.go`

## Key File Locations

**Entry Points:**
- `main.go`: Binary entry point
- `cmd/init.go`: Command tree construction
- `cmd/root.go`: Root command definition and domain setup

**Configuration:**
- `internal/config/config.go`: Config struct and persistence
- `~/.cani/cani.yml`: Runtime config file (user home)

**Core Logic:**
- `pkg/devicetypes/inventory.go`: Inventory struct definition
- `pkg/devicetypes/inventory_crud.go`: Create/Read/Update/Delete operations
- `pkg/devicetypes/inventory_add_remove.go`: Add/Remove with parent assignment
- `pkg/devicetypes/registry.go`: Hardware type registries (slug→definition)
- `pkg/devicetypes/loader.go`: YAML type loading (embedded + filesystem + git)
- `pkg/devicetypes/classify.go`: Device type classification logic

**Provider Interface:**
- `internal/provider/interface.go`: Provider, Exporter, Importer, DeviceStager interfaces
- `internal/provider/registry.go`: Global provider map

**Persistence:**
- `pkg/datastores/datastore.go`: DeviceStore interface
- `pkg/datastores/json.go`: JSON file store implementation
- `~/.cani/canidb.json`: Runtime inventory file (user home)

**Testing:**
- `testdata/`: Test fixtures
- `*_test.go` files co-located with source

## Naming Conventions

**Files:**
- `snake_case.go`: All Go source files
- `*_test.go`: Test files (co-located with implementation)
- Verb-based command files: `add.go`, `show.go`, `import.go`
- Type-prefixed domain files: `cani_device_types.go`, `cani_rack_types.go`
- Concern-split inventory files: `inventory_crud.go`, `inventory_queries.go`, `inventory_relationships.go`

**Directories:**
- `cmd/<verb>/`: One directory per CLI verb
- `pkg/provider/<name>/`: One directory per provider
- `pkg/devicetypes/<category>-types/`: YAML definition directories by hardware category

**Go Packages:**
- Package names match directory names (lowercase, single word)
- Exception: `cmd/import` uses package name `imprt` to avoid keyword collision

**Types and Functions:**
- Exported types: `CamelCase` (e.g., `CaniDeviceType`, `Inventory`, `Provider`)
- Unexported: `camelCase`
- Constructors: `New<Type>()` pattern (e.g., `NewCommand()`, `NewJSONStore()`)
- Interface methods: verb-based (e.g., `Transform()`, `Import()`, `Export()`)

## Where to Add New Code

**New CLI Command (verb):**
- Create directory: `cmd/<verb>/`
- Create `<verb>.go` with `NewCommand() *cobra.Command`
- Register in `cmd/init.go` under `alphaCmd.AddCommand()`

**New Provider:**
- Create directory: `pkg/provider/<name>/`
- Implement `internal/provider.Provider` interface
- Add `init.go` with `func init() { provider.Register("<name>", instance) }`
- Add blank import in `main.go`: `_ "github.com/Cray-HPE/cani/pkg/provider/<name>"`

**New Hardware Type Definition:**
- Add YAML file to appropriate subdirectory in `pkg/devicetypes/<category>-types/<Vendor>/`
- Types are auto-discovered by the loader at startup

**New Inventory Operation:**
- Add method to `pkg/devicetypes/inventory_*.go` (pick appropriate concern file)
- Add tests in corresponding `*_test.go`

**New Datastore Backend:**
- Implement `DeviceStore` interface in `pkg/datastores/`
- Add case to `SetDeviceStore()` in `pkg/datastores/datastore.go`

**New Visual Output Format:**
- Add to `pkg/visual/` following existing patterns (e.g., `table.go`, `tree.go`)

**Utilities:**
- Internal helpers: `internal/util/`
- Public/shared helpers: `pkg/utils/`

## Special Directories

**`vendor/`:**
- Purpose: Vendored Go module dependencies
- Generated: Yes (via `go mod vendor`)
- Committed: Yes

**`bin/`:**
- Purpose: Compiled binary output
- Generated: Yes (via `make build`)
- Committed: No (gitignored)

**`testdata/`:**
- Purpose: Test fixtures (sample inventories, YAML configs, import data)
- Generated: No
- Committed: Yes

**`pkg/devicetypes/device-types/` (and sibling `-types/` dirs):**
- Purpose: Embedded YAML hardware type definitions shipped with the binary
- Generated: No (manually authored)
- Committed: Yes
- Note: Loaded via Go embed at compile time (`pkg/devicetypes/embed.go`)

**`docs/`:**
- Purpose: MkDocs documentation source
- Generated: No
- Committed: Yes

---

*Structure analysis: 2026-04-30*
