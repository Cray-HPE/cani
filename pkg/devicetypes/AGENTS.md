# Devicetypes Package

## Purpose

The devicetypes package is a **foundational piece** of CANI. It defines the data structures for hardware inventory and provides the embedded hardware type library. Every device in a CANI inventory must have a corresponding YAML definition based on NetBox's devicetype-library schema. The package is provider-agnostic — its types model generic hardware hierarchy that can be exported to Nautobot, NetBox, or other systems.

## Layout

| File | Purpose |
|------|---------|
| `cani_type.go` | `CaniType` interface — shared contract for all six inventory types |
| `cani_device_types.go` | `CaniDeviceType` struct — devices (chassis, blades, switches, nodes, PDUs, CDUs) |
| `cani_rack_types.go` | `CaniRackType` struct — racks and cabinets |
| `cani_module_types.go` | `CaniModuleType` struct — modules (NICs, GPUs, CPUs, memory, PSUs) |
| `cani_cable_types.go` | `CaniCableType` struct — cables (DAC, AOC, fiber, Cat) |
| `cani_fru_types.go` | `CaniFruType` struct — field-replaceable units / inventory items |
| `cani_location_types.go` | `CaniLocationType` struct — locations (site, building, floor, room) |
| `inventory.go` | `Inventory` struct — six UUID-keyed maps for all types |
| `inventory_crud.go` | CRUD operations on the inventory |
| `inventory_queries.go` | Query helpers (`DevicesByType`, `Exists`, `FindName`) |
| `inventory_relationships.go` | Rebuilds and validates parent/child relationships at load time |
| `component_specs.go` | Spec types: `InterfaceSpec`, `ConsolePortSpec`, `PowerPortSpec`, `ModuleBaySpec`, `DeviceBaySpec`, `Identification` |
| `registry.go` | `Type` and `Category` enums, `ClassifyForNautobot()`, global registries |
| `all.go` | Lookup functions: `GetBySlug`, `GetByPartNumber`, `GetByManufacturerModel`, `Register*` |
| `constructors.go` | Factory functions for creating typed inventory entries |
| `loader.go` | Loads embedded YAML definitions at init |
| `embed.go` | `embed.FS` directives for YAML directories |
| `gitloader.go` | Loads definitions from a git repository |

## Embedded YAML Directories

| Directory | Contents |
|-----------|----------|
| `device-types/` | Device YAML definitions by manufacturer (HPE, Intel, etc.) |
| `module-types/` | Module YAML definitions (NICs, memory, CPUs, etc.) |
| `rack-types/` | Rack YAML definitions |
| `cable-types/` | Cable type definitions |
| `inventory-types/` | FRU / inventory-item YAML definitions |

## Key Types

| Type | Purpose |
|------|---------|
| `CaniType` | Interface: `Validate()`, `GetID()`, `GetSlug()`, `GetStatus()` — implemented by all six types |
| `Inventory` | Holds six maps: `Locations`, `Racks`, `Devices`, `Modules`, `Cables`, `Frus` |
| `CaniLocationType` | Location hierarchy node (site → building → floor → room) |
| `CaniRackType` | Rack instance + template fields from YAML library |
| `CaniDeviceType` | Device instance + DeviceType template fields from YAML library |
| `CaniModuleType` | Module instance + ModuleType template fields |
| `CaniCableType` | Cable instance with termination references |
| `CaniFruType` | Field-replaceable unit / inventory item |
| `InterfaceSpec` | Port definition on a device or module (name, type, label) |
| `Type` | Hardware classification enum (`rack`, `chassis`, `blade`, `node`, `switch`, `nic`, `cable`, etc.) |
| `Category` | Higher-level grouping (`device`, `module`, `cable`, `rack`, `fru`) |

## Relationship Model

All relationships use single-direction FKs on the child object. Reverse pointers are rebuilt at load time by `VerifyParentChildRelationships()`.

```
Location
  └─ Parent → parent Location
  └─ owns → Rack (via Rack.Location)
       └─ contains → Device (via Device.Parent → rack)
            ├─ children → Device (via Device.Parent → device)  [chassis → blade]
            ├─ modules → Module (via Module.ParentDevice)
            ├─ frus → FRU (via FRU.Device)
            │    └─ nested → FRU (via FRU.Parent)
            └─ interfaces → InterfaceSpec[] (embedded)
                 └─ connected by → Cable (via Cable.TerminationA/B)
```

Key relationships:
- `CaniDeviceType.Parent` is overloaded: can point to a rack UUID or a parent device UUID. The rebuild logic resolves which.
- `CaniDeviceType.Rack`, `.Location`, `.ParentDevice` are explicit FK fields set during rebuild for unambiguous export.
- `CaniRackType.Devices` and `.OccupiedSlots` are derived from device references, not stored.

## Key Functions

**Lookups** (in `all.go`):
- `All()` / `GetBySlug(slug)` / `GetByPartNumber(pn)` / `GetByManufacturerModel(mfr, model)` — Device types
- `AllModules()` / `GetModuleBySlug(slug)` / `GetModuleByPartNumber(pn)` / `GetModuleByManufacturerModel(mfr, model)` — Module types
- `AllRackTypes()` / `GetRackTypeBySlug(slug)` / `GetRackTypeByPartNumber(pn)` — Rack types
- `AllCables()` / `GetCableTypeBySlug(slug)` / `GetCableTypeByPartNumber(pn)` — Cable types
- `AllFruTypes()` / `GetFruTypeBySlug(slug)` / `GetFruTypeByPartNumber(pn)` — FRU types

**Registration** (in `all.go`):
- `RegisterDeviceType(dt)` / `RegisterModuleType(mt)` / `RegisterRackType(rt)` / `RegisterCableType(ct)` / `RegisterFruType(ft)`

**Classification** (in `registry.go`):
- `ClassifyForNautobot(hardwareType)` — Routes a `Type` to a `Category` for export dispatch
- `ListCaniDeviceTypes(types...)` — Filter devices by hardware type
- `ListAllAvailableTypes()` — Flat list of every registered type across all registries

## Nautobot Mapping Summary

The full field-by-field mapping lives in `NAUTOBOT_MAPPING.md`. Quick reference:

| Cani Type | Nautobot Object(s) | Export Coverage | Key Gap |
|---|---|---|---|
| `CaniLocationType` | `Location` + `LocationType` | ~10% | No hierarchy walk; hardcoded "Site" type |
| `CaniRackType` | `Rack` | ~40% | Missing OuterWidth/Depth, Role, Type, Serial |
| `CaniDeviceType` | `Device` + `DeviceType` | ~70% | Role from ProviderMetadata; Comments not mapped |
| `CaniModuleType` | `Module` + `ModuleType` + `ModuleBay` | 0% | Entire export path missing |
| `CaniCableType` | `Cable` | ~60% | Color not mapped; type derived by slug heuristic |
| `CaniFruType` | `InventoryItem` | 0% | Entire export path missing |

Key export files: `pkg/provider/nautobot/mapper.go`, `pkg/provider/nautobot/load.go`, `pkg/provider/nautobot/lookup.go`.

## Adding a New Device Type

1. Create a YAML file under `device-types/<Manufacturer>/<slug>.yaml`
2. Follow the NetBox devicetype-library schema
3. Include required fields: `manufacturer`, `model`, `slug`
4. Include `hardware-type` to classify for export routing
5. Rebuild to embed the new definition
