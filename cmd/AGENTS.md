# CANI overview

`cani` is a portable inventory of hardware devices.

## Two Paths

### Import

A common path allows users to import data from any source they can write code for.

### Manual 

A less common path allows users to run ad-hoc commands to add devices, building up an inventory from scratch.

## Extract, Transform, Load

`cani` works by maintaing the inventory in a portable format based on Netbox's [DeviceType-Library](https://github.com/netbox-community/devicetype-library/tree/master/device-types).

This provides a simple Extract, Transform, and Load (ETL) workflow:

### Extract

Raw data is extracted from any arbitrary source.

This code is written by each provider.  

An example could be as simple as reading a CSV file or gathering information from an API.

### Transform

The raw data is mapped to `cani`-specific fields:

```csv
ProductNumber,ProductDescription,Qty,ConfigGroup
P9K58A,HPE 48U 800mmx1200mm G2 Enterprise Shock Rack,2,0100
```

```json
{
  "Racks": {
    "32183b78-77ea-4a2c-961b-78bb3cfc88ee": {
      "id": "32183b78-77ea-4a2c-961b-78bb3cfc88ee",
      "name": "HPE 48U 800mmx1200mm G2 Enterp-001",
      "rackTypeSlug": "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack",
      "partNumber": "P9K58A",
      "location": "00000000-0000-0000-0000-000000000000",
      "uHeight": 48,
      "devices": [],
      "providerMetadata": {
        "example": {
          "ConfigGroup": "0100",
          "Source": "my_custom_inventory.csv"
        }
      }
    }
  }
}
```

### Load

After mapping the raw data to `cani` fields, loading the new devices into the inventory is trivial because `cani` already knows how to manage these device types.

## Provider Workflow

The majority of the `cani` code can remain untouched.  Developers need to focus on provider-specific logic, which is written in their own sub-package.

Developers can define how data is imported, what fields map to what, and they do not need to worry about how CANI does the work.


## Command Hierarchy

CLI built with the standard-library-only `internal/cli` framework (no Cobra/Viper). Entry point: `main.go` (`init()` calls `cmd.Init()`, `main()` calls `cmd.Execute()`) → `cmd/init.go:Init()` assembles the tree via `newRootCommand()` in `cmd/root.go`.

```
cani (root)
└── alpha/                       # Unstable/WIP wrapper
    ├── import                   # ETL from providers
    ├── add/                     # Add inventory items
    │   ├── location <name>
    │   ├── rack <slug-or-part>
    │   ├── device <slug-or-part>
    │   ├── module <slug-or-part>
    │   ├── cable <slug-or-part>
    │   └── connections <file.yaml>
    │       └── generate <pattern>
    ├── remove/                  # Remove inventory items
    │   ├── location <uuid-or-name>
    │   ├── rack <uuid-or-name>
    │   ├── device <uuid-or-name>
    │   ├── module <uuid-or-name>
    │   └── cable <uuid-or-name>
    ├── update/                  # Modify inventory items
    │   ├── location <uuid-or-name>
    │   ├── rack <uuid-or-name>
    │   ├── device <uuid-or-name>
    │   ├── module <uuid-or-name>
    │   └── cable <uuid-or-name>
    ├── show/                    # Display inventory
    │   ├── location
    │   ├── rack
    │   ├── device
    │   ├── module
    │   └── cable
    ├── export/                  # Export to providers
    └── serve/                   # Future: API server
```

---

## CLI Contract

Commands are organized as **verb → noun**. The five nouns correspond to `Inventory` struct fields:

| Noun | Inventory Map | Underlying Types |
|------|---------------|------------------|
| `location` | `Locations` | site, building, floor, room |
| `rack` | `Racks` | rack, cabinet |
| `device` | `Devices` | blade, node, node-card, chassis, switch, mgmt-switch, hsn-switch, cabinet-pdu, cdu |
| `module` | `Modules` | module, nic, gpu, cpu, memory, power-supply |
| `cable` | `Cables` | cable |

### Identification Rules

| Verb | Positional Arg | Resolution |
|------|----------------|------------|
| `add` | slug or part number (except `location` which takes a name) | Registry lookup via `lookupBySlugOrPart()` — tries slug first, then part number |
| `remove` | UUID or name | `resolve.<Noun>()` — tries `uuid.Parse()` first, then case-insensitive name match |
| `update` | UUID or name | Same as remove |

### `add` Subcommands

Persistent flags inherited by all `add` subcommands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--auto` | `-a` | bool | false | Automatically recommend values |
| `--accept` | `-y` | bool | false | Accept recommended values |
| `--list-supported-types` | `-L` | bool | false | List hardware types for the noun |
| `--qty` | `-q` | int | 1 | Quantity to add |
| `--parent` | `-p` | string | nil UUID | Parent item UUID |

#### `add location <name>`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | `"site"` | Location type (site, building, floor, room) |
| `--parent` | string | `""` | Parent location UUID or name |

#### `add rack <slug-or-part-number>`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--location` | string | `""` | Parent location UUID or name |

#### `add device <slug-or-part-number>`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--rack` | string | `""` | Parent rack UUID or name |
| `--position` | int | 0 | Rack U position |
| `--face` | string | `""` | Rack face (front, rear) |

#### `add module <slug-or-part-number>`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--device` | string | `""` | Parent device UUID or name |
| `--bay` | string | `""` | Module bay name |

#### `add cable <slug-or-part-number>`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--a-device` | string | `""` | Termination A device UUID or name |
| `--a-port` | string | `""` | Termination A port name |
| `--b-device` | string | `""` | Termination B device UUID or name |
| `--b-port` | string | `""` | Termination B port name |
| `--label` | string | `""` | Cable label |
| `--color` | string | `""` | Cable color |

### `remove` Subcommands

Each takes `<uuid-or-name>` as a positional arg. Persistent flag:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--force` | `-f` | bool | false | Skip confirmation |

- `remove location` — Errors if the location still has child racks.
- `remove rack` — Unlinks devices before deleting.
- `remove device` — Cascades to children (modules, cables).
- `remove module` — Removes the module.
- `remove cable` — Removes the cable.

### `update` Subcommands

Each takes `<uuid-or-name>` as a positional arg. Supports typed flags **and** generic `--set key=value` (repeatable).

| Noun | Typed Flags | `--set` Keys |
|------|-------------|--------------|
| `location` | `--name`, `--status`, `--type`, `--description`, `--facility`, `--address` | name, status, location_type, description, facility, physical_address |
| `rack` | `--name`, `--status`, `--role`, `--description`, `--u-height` | name, status, role, description, u_height |
| `device` | `--name`, `--status`, `--role`, `--description`, `--position`, `--face` | name, status, role, description, rack_position, face, serial, asset_tag |
| `module` | `--name`, `--status`, `--role`, `--description`, `--bay` | name, status, role, description, module_bay_name, serial, asset_tag |
| `cable` | `--label`, `--status`, `--color`, `--description` | label, status, color, description |

### `show` Subcommands

Each noun subcommand is a stub (not yet implemented). The parent `show` command supports:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--sort` | `-s` | string | `"name"` | Sort by field |
| `--format` | `-o` | string | `"json"` | Output format |
| `--visual` | `-v` | bool | false | ASCII rack visualization |
| `--rack-view` | — | bool | false | Compact rack view |
| `--show-routing` | — | bool | false | Cable routing visualization |

---

## Key Patterns

### Factory Functions
Each subpackage exports a `NewCommand() *cli.Command` function (e.g., `add.NewCommand()`). These are assembled in `cmd/init.go:Init()`.

### Noun Command Files
Each noun has its own file per verb (e.g., `add/device.go` exports `newDeviceCommand()`—lowercase/unexported). One file per noun, one function per noun.

### Slug/Part Number Validation
`add/validate_noun.go` maps each `Noun` to the hardware types it accepts (`nounTypeMap`) and validates the positional arg against slug and part-number registries.

### UUID-or-Name Resolution
`internal/util/resolve/` exports `Location()`, `Rack()`, `Device()`, `Module()`, `Cable()`. Each tries `uuid.Parse()` first, then case-insensitive name search. Errors on zero or multiple matches.

### Generic `--set` Flags
`update/set.go` exports `parseSetFlags()` which splits `key=value` pairs. Each noun's update command applies known keys via a `switch` statement, ignoring unknowns with a warning.

### Provider Extension
Providers implement `NewProviderCmd()` to decorate commands. Registered via blank imports in `main.go`, discovered at runtime from `provider.GetProviders()`.

## Key Files

| File | Purpose |
|------|---------|
| `root.go` | Root command (`newRootCommand`), config wiring, `--config`/`--debug` flags |
| `init.go` | Assembles command tree, integrates providers |
| `import.go` | ETL pipeline (Extract → Transform → Load) |
| `add/validate_noun.go` | Slug/part-number validation and `--list-supported-types` |
| `add/table.go` | Type-listing table printer |
| `update/set.go` | `--set key=value` parser |
| `internal/util/resolve/resolve.go` | UUID-or-name resolution helper |
| `pkg/devicetypes/inventory_add_remove.go` | Inventory CRUD for location, rack, module, cable |
| `pkg/devicetypes/inventory_crud.go` | Inventory CRUD for devices |

---

## Provider Interface

Providers live in `pkg/provider/<name>/` and implement interfaces from `internal/provider/interface.go`.

### Required Interface: `Provider`

All providers must implement:

```go
type Provider interface {
    // Transform converts imported data into CANI's format (ETL "Transform" step).
    // ctx carries cancellation/deadlines from the command layer.
    Transform(ctx context.Context, existing devicetypes.Inventory) (*devicetypes.TransformResult, error)

    // NewProviderCmd customizes CLI commands for this provider
    NewProviderCmd(base *cli.Command) (*cli.Command, error)

    // Slug returns the provider's identifier (e.g., "example", "nautobot")
    Slug() string
}
```

### Optional Interfaces

| Interface | Purpose |
|-----------|---------|
| `Importer` | Extract data from external systems (`Import()`) |
| `Exporter` | Load data to external systems (`Export()`) |
| `HasOptions` | Expose default config options |
| `HasImportOptions` | Import-specific config and CLI flag binding |
| `HasExportOptions` | Export-specific config and CLI flag binding |

### Provider File Structure

```
pkg/provider/<name>/
├── init.go        # Register provider in init(), implement NewProviderCmd()
├── provider.go    # Provider struct and Slug()
├── options.go     # Options, ImportOptions, ExportOptions structs
├── import.go      # Import() method (if Importer)
├── export.go      # Export() method (if Exporter)
├── transform.go   # Transform() method
├── commands/      # CLI command customizations
│   └── commands.go
├── import/        # Import logic
├── export/        # Export logic
└── transform/     # Transform logic
```

### Registration

Providers self-register in `init()`:

```go
func init() {
    instance = New()
    provider.Register("example", instance)
}
```

Enable via blank import in `main.go`:

```go
import _ "github.com/Cray-HPE/cani/pkg/provider/example"
```

### Options Pattern

Config structs use YAML tags with `head_comment` for documentation:

```go
type Options struct {
    URL   string `yaml:"url" head_comment:"Base URL for the API"`
    Token string `yaml:"token" head_comment:"API authentication token"`
}
```

Implement `HasOptions`, `HasImportOptions`, or `HasExportOptions` for auto-population in config files with preserved comments.

### NewProviderCmd Pattern

Switch on command name to customize each subcommand:

```go
func (p *Example) NewProviderCmd(base *cli.Command) (*cli.Command, error) {
    switch base.Name() {
    case "import":
        return commands.NewImportCommand(base)
    case "export":
        return commands.NewExportCommand(base)
    default:
        return base, nil
    }
}
