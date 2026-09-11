# Devicetypes Package

## Purpose

The devicetypes package is a **foundational piece** of CANI. It defines the data structures for hardware inventory and provides the embedded hardware type library. Every device in a CANI inventory must have a corresponding YAML definition based on NetBox's devicetype-library schema. The package is provider-agnostic — its types model generic hardware hierarchy that can be exported to Nautobot, NetBox, or other systems.

## Layout

| File | Purpose |
|------|---------|
| `cani_type.go` | `CaniType` interface — shared contract for all DCIM and IPAM inventory types |
| `cani_device_types.go` | `CaniDeviceType` struct — devices (chassis, blades, switches, nodes, PDUs, CDUs) |
| `cani_rack_types.go` | `CaniRackType` struct — racks and cabinets |
| `cani_module_types.go` | `CaniModuleType` struct — modules (NICs, GPUs, CPUs, memory, PSUs) |
| `cani_cable_types.go` | `CaniCableType` struct — cables (DAC, AOC, fiber, Cat) |
| `cani_fru_types.go` | `CaniFruType` struct — field-replaceable units / inventory items |
| `cani_location_types.go` | `CaniLocationType` struct — locations (site, building, floor, room) |
| `inventory.go` | `Inventory` struct — DCIM, interface, IPAM, and metadata collections + `TransformResult` dedup helper |
| `inventory_crud.go` | Shared add/remove operations and relationship verification |
| `inventory_merge_devices.go` | Device merge and provider-identity lookup |
| `inventory_merge_locations.go` | Location and rack natural-key merges |
| `inventory_merge_components.go` | Module and FRU natural-key merges |
| `inventory_merge_cables.go` | Cable identity and endpoint-based merge |
| `inventory_merge_ipam.go` | VLAN, prefix, IP address, and VRF merges |
| `transform_merge.go` | Atomic full-result merge transaction |
| `transform_remap.go` | Canonical UUID remapping and IPAM parent derivation |
| `inventory_add_remove.go` | Single-item add/remove with validation and relationship rebuild |
| `inventory_queries.go` | Location, rack, component, and interface query helpers |
| `inventory_device_queries.go` | Device lookup, connectable lookup, module-bay occupancy, and parent-kind checks |
| `inventory_validate.go` | Full-inventory referential validation |
| `inventory_relationships.go` | Rebuilds and validates parent/child relationships at load time |
| `inventory_index.go` | O(1) provider-key lookup index for device dedup during merge |
| `inventory_orphans.go` | Orphan detection for devices and racks without parents |
| `component_specs.go` | Spec types: `InterfaceSpec`, `ConsolePortSpec`, `PowerPortSpec`, `ModuleBaySpec`, `DeviceBaySpec`, `Identification` |
| `registry.go` | `Type` and `Category` enums, `ClassifyForNautobot()`, global registries |
| `all.go` | Lookup functions: `GetBySlug`, `GetByPartNumber`, `GetByManufacturerModel`, `Register*` |
| `lookup_any.go` | `LookupAny()` — cross-registry slug/PN search (rack→device→module→cable) |
| `lookups.go` | Scored lookup: `Lookup`, `LookupScored`, `LookupModule`, `FuzzyMatchAll`, `scoreFields` |
| `classify.go` | `SuggestTypes()` — query decomposition and hardware-type fallback for unclassified devices |
| `classify_interactive.go` | Interactive classification prompt (`PromptForDeviceType`, `searchSlugs`) |
| `constructors.go` | Factory functions for creating typed inventory entries |
| `suggest_parent.go` | `SuggestParents()` — scored parent suggestions for orphan devices/racks |
| `reparent_interactive.go` | Interactive parent assignment prompt (`PromptForParent`) |
| `resolve_plan.go` | `ResolvePlan` — serializable parent assignment plan (write, read, apply) |
| `provider_metadata.go` | Provider metadata helpers (`GetProviderMeta`, `SetProviderMeta`, `FlattenProviderMetadata`) |
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
| `CaniType` | Interface: `Validate()`, `GetID()`, `GetSlug()`, `GetStatus()` — implemented by all ten DCIM/IPAM object types |
| `Inventory` | Holds schema/provider context; DCIM, interface, and IPAM maps; metadata definitions; and a transient provider-key index |
| `CaniLocationType` | Location hierarchy node (site → building → floor → room) |
| `CaniRackType` | Rack instance + template fields from YAML library |
| `CaniDeviceType` | Device instance + DeviceType template fields from YAML library |
| `CaniModuleType` | Module instance + ModuleType template fields |
| `CaniCableType` | Cable instance with termination references |
| `CaniFruType` | Field-replaceable unit / inventory item |
| `InterfaceSpec` | Port definition on a device or module (name, type, label) |
| `Type` | Hardware classification enum (`rack`, `chassis`, `blade`, `node`, `switch`, `nic`, `cable`, etc.) |
| `Category` | Higher-level grouping (`device`, `module`, `cable`, `rack`, `fru`) |

## Tag Conventions

The `Cani*Type` YAML/JSON tags mirror the upstream **NetBox devicetype-library** schema (the source of truth), not an internal house style:

- **Component-collection keys are kebab-case** and must stay that way: `console-ports`, `power-ports`, `module-bays`, `device-bays`, `allowed-children`, `hardware-type`. The embedded library (`device-types/`, `rack-types/`, `module-types/`, …) and the loader depend on these exact keys — renaming them to snake_case breaks loading every bundled type.
- **Scalar fields keep NetBox's snake_case** names: `part_number`, `u_height`, `is_full_depth`, `subdevice_role`, `weight_unit`, and entry sub-fields like `mgmt_only`, `maximum_draw`.
- **JSON tags are camelCase** (`partNumber`, `uHeight`).

Do not "standardize" the kebab-case keys to snake_case; it is intentional NetBox parity.

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
- `LookupAny(key)` — Cross-registry search: rack → device → module → cable by slug or PN

**Scored Lookups** (in `lookups.go`):
- `Lookup(query)` / `LookupScored(query)` — Exact + fuzzy device lookup with confidence score
- `LookupModule(query)` / `LookupModuleScored(query)` — Same for modules
- `FuzzyMatchAll(query, max)` — Multi-result fuzzy search across all device types
- `ScoreTierLabel(score)` — Human-readable label for a score ("exact match", "slug match", etc.)

**Registration** (in `all.go`):
- `RegisterDeviceType(dt)` / `RegisterModuleType(mt)` / `RegisterRackType(rt)` / `RegisterCableType(ct)` / `RegisterFruType(ft)`

**Classification** (in `registry.go` and `classify.go`):
- `ClassifyForNautobot(hardwareType)` — Routes a `Type` to a `Category` for export dispatch
- `ListCaniDeviceTypes(types...)` — Filter devices by hardware type
- `ListAllAvailableTypes()` — Flat list of every registered type across all registries
- `SuggestTypes(device, maxResults)` — Query decomposition + fuzzy + fallback for unclassified devices

**Inventory CRUD** (in `inventory_crud.go` and `inventory_add_remove.go`):
- `EnsureLocation()` — Guarantee at least one location exists
- `AssignRacksToLocation(locID)` — Link orphan racks to a location
- `AddDevices(batch)` / `MergeDevices(incoming)` / `MergeDevicesStrict(incoming, strict)` — Batch device operations
- `MergeTransformResult(result)` — atomically clone, merge, remap, validate, and commit a complete provider transform result
- `MergeRacks(incoming)` / `MergeLocations(incoming)` / `MergeModules(incoming)` / `MergeFrus(incoming)` / `MergeCables(incoming)` — Merge by UUID → external identity → natural key → insert
- `MergeVLANs(incoming)` / `MergePrefixes(incoming)` / `MergeIPAddresses(incoming)` / `MergeVRFs(incoming)` — IPAM natural-key merges returning canonical UUID remaps
- `RemoveDevice(id)` — Cascading delete (unlinks parent, removes cables/modules/children)
- `AddLocation(loc)` / `AddRack(rack)` / `AddModule(mod)` / `AddCable(cable)` — Single-item insert with validation
- `RemoveLocation(id)` / `RemoveRack(id)` / `RemoveModule(id)` / `RemoveCable(id)` — Single-item delete with constraints

**Queries** (in `inventory_queries.go`):
- `FindLocationByName(name)` / `FindRackByName(name)` / `FindModuleByName(name)` / `FindFruByName(name)` / `FindCableByLabel(label)`
- `LocationExists(name)` / `RackExists(name)` / `ModuleExists(name)` / `FruExists(name)`
- `GetDevicesInRack(rackID)` / `GetCablesForDevice(devID)` / `GetModulesForDevice(devID)`
- `GetInterfaceByID(ifaceID)` / `GetInterfacesByDevice(devID)`
- `Validate()` — Full referential integrity check across all maps

**Provider Metadata** (in `provider_metadata.go`):
- `GetProviderMeta(key)` / `SetProviderMeta(provider, key, value)` / `GetProviderSubMap(provider)`
- `FlattenProviderMetadata()` — Merge all provider sub-maps into a flat map
- `FindDeviceByProviderKey(provider, key, value)` — O(1) index lookup for dedup

**Orphan Management** (in `suggest_parent.go`, `reparent_interactive.go`, `resolve_plan.go`):
- `SuggestParents(inv, orphan)` — Scored parent suggestions for devices/racks
- `PromptForParent(inv, orphan, opts)` — Interactive parent assignment
- `WritePlan(path, plan)` / `ReadPlan(path)` / `ApplyPlan(inv, plan)` — Serializable resolve plans
- `PlaceDeviceInRack(dev, devID, rack, startU, face)` — Auto-find slot and set position

## Nautobot Export Coverage

The full field-by-field mapping lives in `NAUTOBOT_MAPPING.md`. Quick reference:

| Cani Type | Nautobot Object(s) | Approx. Test Coverage | Notes |
|---|---|---|---|
| `CaniLocationType` | `Location` + `LocationType` | ~79.2% | Topological sort; custom fields and provider metadata mapped; Tags/Tenant not mapped |
| `CaniRackType` | `Rack` | ~79.2% | Physical fields, type, serial, asset tag, facility ID, comments, location, and create-time tags mapped |
| `CaniDeviceType` | `Device` + `DeviceType` | ~79.2% | Role and location hierarchy mapped; create-time tags mapped; Platform/Tenant not mapped |
| `CaniModuleType` | `Module` + `ModuleType` + `ModuleBay` | ~79.2% | ModuleType + ModuleBay auto-created; module interfaces created on parent device |
| `CaniCableType` | `Cable` | ~79.2% | Color mapped; Type via multi-priority resolution; Tags not mapped |
| `CaniFruType` | `InventoryItem` | ~79.2% | Topological nesting, Manufacturer FK, Tags, and CustomFields mapped |

Coverage is Go statement coverage for the shared `pkg/provider/nautobot/export` package, reported by the `go test -cover ./...` phase of `make test`. It is not per-object coverage or field-mapping completeness.

Key export files: `pkg/provider/nautobot/export/mapper.go`, `pkg/provider/nautobot/export/load.go`, `pkg/provider/nautobot/export/lookup.go`, `pkg/provider/nautobot/export/load_locations.go`, `pkg/provider/nautobot/export/load_modules.go`, `pkg/provider/nautobot/export/load_frus.go`.

## Testing

**Run tests:**
```bash
go test ./pkg/devicetypes/... -v
```

**Test files (28):**

| Test File | Covers | Key Patterns |
|-----------|--------|--------------|
| `all_test.go` | Registry lookups, registration | Uses `resetRegistries()` helper to isolate; `seedDevice()` for fixtures |
| `cani_type_test.go` | `CaniType` interface methods | Tests all six implementors |
| `cani_device_types_test.go` | `CaniDeviceType` methods | `MergeProperties`, `InstantiateInterfaces`, getters |
| `cani_rack_types_test.go` | `CaniRackType` methods | Slot placement, `CanFitDevice`, `FindNextAvailableSlot` |
| `cani_module_types_test.go` | `CaniModuleType` methods | Interface instantiation, getters |
| `cani_cable_types_test.go` | `CaniCableType` methods | `NewCable`, `SetTerminations`, `ValidateCable` |
| `cani_fru_types_test.go` | `CaniFruType` methods | Validate, getters |
| `cani_location_types_test.go` | `CaniLocationType` methods | `AddRack`, `AddChild`, validate |
| `inventory_test.go` | `NewInventory`, `EnsureUniqueDeviceNames` | Map initialization, name dedup suffixes |
| `inventory_crud_test.go` | CRUD operations | `EnsureLocation`, `MergeDevices`, `MergeRacks`, `MergeLocations`, `RemoveDevice`, `providerIdentityCompatible` |
| `inventory_add_remove_test.go` | Single-item add/remove | `AddLocation`/`AddRack`/`AddModule`/`AddCable`, `RemoveLocation` constraints |
| `inventory_queries_test.go` | Find/Exists/Validate queries | Referential integrity: `validateModuleRefs`, `validateCableRefs`, `validateFruRefs` |
| `inventory_relationships_test.go` | Relationship rebuild | Location/rack/device rebuild, cycle detection, cable/FRU validation |
| `inventory_index_test.go` | Provider key index | `RebuildProviderKeyIndex`, `indexDeviceMetadata`, `toIndexValue` |
| `inventory_orphans_test.go` | Orphan detection | `OrphanDevices`, `OrphanRacks` |
| `inventory_orphan_reparent_test.go` | Orphan + reparent flow | `VerifyPopulatesOrphans`, `ReparentDeviceViaParentField` |
| `inventory_export_test.go` | Export edge cases | Fixture-based tests for classification, FRU cycles, cable truncation |
| `classify_test.go` | Classification helpers | `SuggestTypes`, `normalizeHardwareType`, `hardwareTypeFallback`, `collectQueries` |
| `classify_interactive_test.go` | Interactive helpers | `searchSlugs`, `colorFuncs` |
| `constructors_test.go` | Factory functions | All `New*FromSlug`/`New*FromPartNumber` constructors |
| `lookups_test.go` | Scored lookups | `Lookup`, `LookupScored`, `LookupModule`, `LookupModuleScored`, `scoreFields`, `FuzzyMatchAll`, `tokenizeCamelNum` |
| `lookup_any_test.go` | Cross-registry lookup | `LookupAny` by slug, PN, empty key, no match |
| `registry_test.go` | Type/Category enums | `ClassifyForNautobot`, `ListCaniDeviceTypes`, `ListAllAvailableTypes` |
| `component_specs_test.go` | Spec types | JSON round-trip, unmarshal, interface constants |
| `provider_metadata_test.go` | Provider metadata | `GetProviderMeta`, `SetProviderMeta`, `FlattenProviderMetadata`, `FindDeviceByProviderKey` |
| `suggest_parent_test.go` | Parent suggestions | `SuggestParents`, `nameSimilarity`, `SearchParentCandidates` |
| `reparent_interactive_test.go` | Interactive reparent | `PromptForParent` with simulated input |
| `resolve_plan_test.go` | Resolve plans | `WritePlan`/`ReadPlan` round-trip, `ApplyPlan`, `PlaceDeviceInRack` |

**Test patterns:**
- Flat test functions (no subtests), each named `Test<Function><Scenario>`
- `NewInventory()` for fresh inventory; `defer func() { delete(registry, slug) }()` for registry cleanup
- `t.Errorf` for soft failures, `t.Fatalf` for critical assertions
- `resetRegistries()` in `all_test.go` clears global maps — tests in other files that depend on embedded data must register their own fixtures

## Adding a New Device Type

1. Create a YAML file under `device-types/<Manufacturer>/<slug>.yaml`
2. Follow the NetBox devicetype-library schema
3. Include required fields: `manufacturer`, `model`, `slug`
4. Include `hardware-type` to classify for export routing
5. Rebuild to embed the new definition
