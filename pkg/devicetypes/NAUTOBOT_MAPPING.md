# CANI Inventory ↔ Nautobot 3.x Mapping Guide

This document is the authoritative developer reference for how CANI's `Inventory` struct and its six `Cani*Type` collections map to Nautobot 3.x API objects. It covers the current export implementation, relationship model, and improvement recommendations.

---

## 1. Overview

The `Inventory` struct (`pkg/devicetypes/inventory.go`) holds six UUID-keyed maps:

```go
type Inventory struct {
    Locations map[uuid.UUID]*CaniLocationType
    Racks     map[uuid.UUID]*CaniRackType
    Devices   map[uuid.UUID]*CaniDeviceType
    Modules   map[uuid.UUID]*CaniModuleType
    Cables    map[uuid.UUID]*CaniCableType
    Frus      map[uuid.UUID]*CaniFruType
}
```

Each `Cani*Type` implements the `CaniType` interface (`Validate()`, `GetID()`, `GetSlug()`, `GetStatus()`).

**Current export coverage:**

| Inventory Map | Nautobot Object(s) | Coverage | Export Path |
|---|---|---|---|
| `Locations` | `Location` + `LocationType` | ~70% | `loadLocations()` in `load_locations.go` |
| `Racks` | `Rack` | ~50% | `createRackFromCaniRack()` in `load.go` |
| `Devices` | `Device` + `DeviceType` | ~80% | `MapToWritableDeviceRequest()` in `mapper.go` |
| `Modules` | `Module` + `ModuleType` + `ModuleBay` | ~60% | `loadModules()` in `load_modules.go` |
| `Cables` | `Cable` | ~75% | `createCaniCableType()` in `load.go` |
| `Frus` | `InventoryItem` | ~60% | `loadFrus()` in `load_frus.go` |

---

## 2. Type-by-Type Field Mapping

Status legend:

- **Mapped** — field is exported to Nautobot
- **Partial** — exported with caveats (see notes)
- **Not Mapped** — field exists on both sides but is not wired in the export
- **Not Exported** — entire type has no export path
- **Cani-Internal** — no Nautobot equivalent; exists for internal hierarchy or validation

### 2.1 `CaniLocationType` → Nautobot `Location`

Source: `pkg/devicetypes/cani_location_types.go`

Locations are **exported as first-class objects** in Phase 0. `loadLocations()` in `load_locations.go` performs a topological sort (BFS from roots) to ensure parents are created before children. The `LocationType` field on each `CaniLocationType` is resolved as a Nautobot `LocationType` FK (defaults to `"Site"` if empty). Created locations are cached for downstream rack/device FK resolution.

| Cani Field | Go Type | Nautobot Field | Status | Notes |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | — | Cani-Internal | Primary key |
| `Name` | `string` | `Location.Name` | **Mapped** | |
| `LocationType` | `string` | `Location.LocationType` (FK) | **Mapped** | Resolved by name; defaults to `"Site"` if empty |
| `Parent` | `uuid.UUID` | `Location.Parent` (FK) | **Mapped** | Topological sort ensures parent exists first |
| `Children` | `[]uuid.UUID` | — | Cani-Internal | Rebuilt from `Parent` at load time |
| `Racks` | `[]uuid.UUID` | — | Cani-Internal | Rebuilt from `CaniRackType.Location` |
| `Status` | `string` | `Location.Status` (FK) | **Mapped** | Falls back to provider default |
| `Facility` | `string` | `Location.Facility` | **Mapped** | Mapped when non-empty |
| `Description` | `string` | `Location.Description` | **Mapped** | Mapped when non-empty |
| `PhysicalAddress` | `string` | `Location.PhysicalAddress` | **Mapped** | Mapped when non-empty |
| `ShippingAddress` | `string` | `Location.ShippingAddress` | **Mapped** | Mapped when non-empty |
| `Latitude` | `string` | `Location.Latitude` | **Mapped** | Parsed to float64 |
| `Longitude` | `string` | `Location.Longitude` | **Mapped** | Parsed to float64 |
| `ContactName` | `string` | `Location.ContactName` | **Mapped** | Mapped when non-empty |
| `ContactPhone` | `string` | `Location.ContactPhone` | **Mapped** | Mapped when non-empty |
| `ContactEmail` | `string` | `Location.ContactEmail` | **Mapped** | Mapped when non-empty |
| `TimeZone` | `string` | `Location.TimeZone` | **Mapped** | Mapped when non-empty |
| `Asn` | `*int64` | `Location.Asn` | **Mapped** | Mapped when non-nil |
| `Comments` | `string` | `Location.Comments` | **Mapped** | Mapped when non-empty |
| `Tenant` | `string` | `Location.Tenant` (FK) | Not Mapped | |
| `Tags` | `[]string` | `Location.Tags` | Not Mapped | |
| `CustomFields` | `map[string]any` | `Location.CustomFields` | **Mapped** | Mapped when non-empty |

---

### 2.2 `CaniRackType` → Nautobot `Rack`

Source: `pkg/devicetypes/cani_rack_types.go`
Mapper: `createRackFromCaniRack()` in `pkg/provider/nautobot/load.go`

Nautobot does **not** have a separate `RackType` model — it uses an enum on the `Rack` object. CANI's rack YAML library provides template fields (`Manufacturer`, `Model`, `PartNumber`, `DeviceBays`, `ModuleBays`) that pre-populate instance fields but do not create a separate Nautobot object.

| Cani Field | Go Type | Nautobot Field | Status | Notes |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | — | Cani-Internal | Primary key |
| `Name` | `string` | `Rack.Name` | **Mapped** | |
| `Slug` | `string` | — | Cani-Internal | Library lookup key |
| `PartNumber` | `string` | — | Cani-Internal | Template field from YAML library |
| `Manufacturer` | `string` | — | Cani-Internal | Template field |
| `Model` | `string` | — | Cani-Internal | Template field |
| `Description` | `string` | — | Cani-Internal | Template field; no Nautobot rack description |
| `HardwareType` | `string` | — | Cani-Internal | Classification |
| `UHeight` | `int` | `Rack.UHeight` | **Mapped** | Defaults to 48 if 0 |
| `OuterWidth` | `int` | `Rack.OuterWidth` | **Mapped** | Mapped when > 0 |
| `OuterDepth` | `int` | `Rack.OuterDepth` | **Mapped** | Mapped when > 0 |
| `OuterUnit` | `string` | `Rack.OuterUnit` | Not Mapped | mm or in |
| `Width` | `string` | `Rack.Width` | Not Mapped | Nautobot WidthEnum (10/19/21/23 in) |
| `Weight` | `float64` | — | Cani-Internal | No Nautobot equivalent |
| `WeightUnit` | `string` | — | Cani-Internal | No Nautobot equivalent |
| `DeviceBays` | `[]DeviceBaySpec` | — | Cani-Internal | Template: defines available bays |
| `ModuleBays` | `[]ModuleBaySpec` | — | Cani-Internal | Template: defines available bays |
| `Location` | `uuid.UUID` | `Rack.Location` (FK) | Partial | Resolved by default location name, not from `CaniLocationType` |
| `Status` | `string` | `Rack.Status` (FK) | **Mapped** | Falls back to provider default |
| `Role` | `string` | `Rack.Role` (FK) | Not Mapped | |
| `RackType` | `string` | `Rack.Type` | Not Mapped | Nautobot enum: `2-post-frame`, `4-post-cabinet`, etc. |
| `Serial` | `string` | `Rack.Serial` | Not Mapped | |
| `AssetTag` | `string` | `Rack.AssetTag` | Not Mapped | |
| `FacilityId` | `string` | `Rack.FacilityId` | Not Mapped | |
| `DescUnits` | `bool` | `Rack.DescUnits` | Not Mapped | Descending unit numbering |
| `Comments` | `string` | `Rack.Comments` | **Mapped** | Mapped when non-empty |
| `Devices` | `[]uuid.UUID` | — | Cani-Internal | Rebuilt from `CaniDeviceType.Rack` |
| `OccupiedSlots` | `map[int]map[string]uuid.UUID` | — | Cani-Internal | Rebuilt from device RackPosition + Face |
| `Tenant` | `string` | `Rack.Tenant` (FK) | Not Mapped | |
| `Tags` | `[]string` | `Rack.Tags` | Not Mapped | |
| `CustomFields` | `map[string]any` | `Rack.CustomFields` | Not Mapped | |
| `ProviderMetadata` | `map[string]any` | — | Partial | `u_height` extracted; rest not mapped |
| `Source` | `string` | — | Cani-Internal | |

---

### 2.3 `CaniDeviceType` → Nautobot `Device` + `DeviceType`

Source: `pkg/devicetypes/cani_device_types.go`
Mapper: `MapToWritableDeviceRequest()`, `MapToPatchRequest()` in `pkg/provider/nautobot/mapper.go`
DeviceType auto-creation: `CreateDeviceTypeFromLocal()` in `pkg/provider/nautobot/lookup.go`

This type conflates **template fields** (from the YAML library, used to auto-create a Nautobot `DeviceType`) and **instance fields** (from the user's inventory, used to create a Nautobot `Device`). See §3 for the template/instance split.

#### Device instance fields

| Cani Field | Go Type | Nautobot Field | Status | Notes |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `BulkWritableDeviceRequest.Id` | **Mapped** | Used for bulk creates |
| `Name` | `string` | `Device.Name` | **Mapped** | |
| `Slug` | `string` | — | **Mapped** (indirectly) | Resolves `Device.DeviceType` FK via cache lookup |
| `Serial` | `string` | `Device.Serial` | **Mapped** | |
| `AssetTag` | `string` | `Device.AssetTag` | **Mapped** | |
| `Status` | `string` | `Device.Status` (FK) | **Mapped** | Falls back to provider default, then "Active" |
| `Role` | `string` | `Device.Role` (FK) | **Mapped** | Explicit `Role` field checked first, then `ProviderMetadata["role"]` fallback |
| `Parent` | `uuid.UUID` | `Device.Rack` (FK) or `Device.ParentBay` | **Mapped** | Overloaded: mapper checks Racks first, then Devices |
| `RackPosition` | `int` | `Device.Position` | **Mapped** | Only when parent resolves to a rack |
| `Face` | `string` | `Device.Face` | **Mapped** | Resolved via `resolveFace()`: `"front"`, `"rear"`, or nil if empty |
| `Rack` | `uuid.UUID` | `Device.Rack` (FK) | Cani-Internal | Explicit FK; rebuilt from `Parent` at load time |
| `Location` | `uuid.UUID` | `Device.Location` (FK) | Partial | Resolved from `ProviderMetadata["location"]` or default |
| `ParentDevice` | `uuid.UUID` | `Device.ParentBay` | Cani-Internal | Explicit FK; rebuilt from `Parent` at load time |
| `Comments` | `string` | `Device.Comments` | **Mapped** | Mapped when non-empty |
| `Description` | `string` | — | Cani-Internal | No Nautobot device description |
| `Vendor` | `string` | — | Cani-Internal | Cani-specific (vendor ≠ manufacturer) |
| `Type` | `Type` | — | Cani-Internal | Classification enum |
| `HardwareType` | `string` | — | Cani-Internal | Used by `ClassifyForNautobot()` |
| `Children` | `[]uuid.UUID` | — | Cani-Internal | Rebuilt from `Parent` at load time |
| `Weight` / `WeightUnit` | `float64` / `string` | — | Cani-Internal | No Nautobot device weight |
| `Identifications` | `[]Identification` | — | Cani-Internal | Alternate manufacturer/model IDs |
| `Platform` | `string` | `Device.Platform` (FK) | Not Mapped | |
| `Tenant` | `string` | `Device.Tenant` (FK) | Not Mapped | |
| `Tags` | `[]string` | `Device.Tags` | Not Mapped | |
| `CustomFields` | `map[string]any` | `Device.CustomFields` | Partial | `ProviderMetadata` exported as CustomFields |
| `ProviderMetadata` | `map[string]any` | `Device.CustomFields` | **Mapped** | Full map sent; `location` and `role` keys also used for FK resolution |
| `Source` | `string` | — | Cani-Internal | |

#### DeviceType template fields (auto-creation)

When `create_device_types` is enabled, `CreateDeviceTypeFromLocal()` looks up the CANI embedded library by slug and creates a Nautobot `DeviceType` + `Manufacturer`:

| Cani Field | Nautobot DeviceType Field | Status |
|---|---|---|
| `Model` | `DeviceType.Model` | **Mapped** |
| `Manufacturer` | `DeviceType.Manufacturer` (FK, auto-created) | **Mapped** |
| `PartNumber` | `DeviceType.PartNumber` | **Mapped** |
| `UHeight` | `DeviceType.UHeight` | **Mapped** |
| `IsFullDepth` | `DeviceType.IsFullDepth` | **Mapped** |
| `SubdeviceRole` | `DeviceType.SubdeviceRole` | **Mapped** | `"parent"` or `"child"` for chassis/blade relationships |
| `Interfaces` | Creates `InterfaceTemplate` objects | Not Mapped | Templates not created; interfaces created on device instance |
| `ConsolePorts` | Creates `ConsolePortTemplate` objects | Not Mapped | |
| `PowerPorts` | Creates `PowerPortTemplate` objects | Not Mapped | |
| `ModuleBays` | Creates `ModuleBayTemplate` objects | Not Mapped | |
| `DeviceBays` | Creates `DeviceBayTemplate` objects | Not Mapped | |

---

### 2.4 `CaniModuleType` → Nautobot `Module` + `ModuleType`

Source: `pkg/devicetypes/cani_module_types.go`
Exported in Phase 4 by `loadModules()` in `load_modules.go`. For each module, the pipeline: (1) gets or creates a `ModuleType` from YAML library data, (2) gets or creates a `ModuleBay` on the parent device, (3) creates the `Module` with the resolved FKs.

| Cani Field | Go Type | Nautobot Field | Status | Notes |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | — | Cani-Internal | Primary key |
| `Name` | `string` | `Module.Name` | **Mapped** | |
| `Slug` | `string` | — | **Mapped** (indirectly) | Used to resolve `ModuleType` from library |
| `PartNumber` | `string` | `ModuleType.PartNumber` | **Mapped** | Template → `ModuleType` |
| `Manufacturer` | `string` | `ModuleType.Manufacturer` (FK) | **Mapped** | Template → `ModuleType`, FK resolved by name |
| `Model` | `string` | `ModuleType.Model` | **Mapped** | Template → `ModuleType` |
| `Description` | `string` | — | Cani-Internal | No Nautobot Module description |
| `HardwareType` | `string` | — | Cani-Internal | Cani classification |
| `Weight` / `WeightUnit` | `float64` / `string` | — | Cani-Internal | No Nautobot equivalent |
| `Comments` | `string` | `ModuleType.Comments` | **Mapped** | Template → `ModuleType.Comments` |
| `Interfaces` | `[]InterfaceSpec` | Creates `Interface` objects (with Module FK) | Not Mapped | Template; interface creation not yet wired |
| `ParentDevice` | `uuid.UUID` | `Module.Device` (via parent device's module bay) | **Mapped** | Resolved to Nautobot device ID via cache |
| `ModuleBayName` | `string` | `Module.ParentModuleBay` (FK) | **Mapped** | Gets or creates ModuleBay on parent device |
| `Serial` | `string` | `Module.Serial` | **Mapped** | Mapped when non-empty |
| `AssetTag` | `string` | `Module.AssetTag` | **Mapped** | Mapped when non-empty |
| `Status` | `string` | `Module.Status` (FK) | **Mapped** | Falls back to provider default |
| `Role` | `string` | `Module.Role` (FK) | **Mapped** | Resolved by name when non-empty |
| `Location` | `uuid.UUID` | `Module.Location` (FK) | **Mapped** | Resolved via cache when non-nil |
| `Tenant` | `string` | `Module.Tenant` (FK) | Not Mapped | |
| `Tags` | `[]string` | `Module.Tags` | Not Mapped | |
| `CustomFields` | `map[string]any` | `Module.CustomFields` | Not Mapped | |
| `Source` | `string` | — | Cani-Internal | |

---

### 2.5 `CaniCableType` → Nautobot `Cable`

Source: `pkg/devicetypes/cani_cable_types.go`
Mapper: `createCaniCableType()` in `pkg/provider/nautobot/load.go`

Cables use a library mechanism identical to devices: a slug/part-number resolves to a YAML definition that pre-fills specs (category, connector, length, color). No separate Nautobot cable "type" object is created.

| Cani Field | Go Type | Nautobot Field | Status | Notes |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | — | Cani-Internal | Primary key |
| `Slug` | `string` | — | Partial | Pattern-matched to derive `Cable.Type` (see below) |
| `Label` | `string` | `Cable.Label` | **Mapped** | |
| `PartNumber` | `string` | — | Cani-Internal | Library/template field |
| `Manufacturer` | `string` | — | Cani-Internal | Library/template field |
| `Model` | `string` | — | Cani-Internal | Library/template field |
| `Description` | `string` | — | Cani-Internal | Library/template field |
| `HardwareType` | `string` | — | Cani-Internal | Classification |
| `CableCategory` | `string` | — | **Mapped** (indirectly) | Used by `resolveCableType()` to derive `Cable.Type` when `CableType` is empty |
| `ConnectorType` | `string` | — | **Mapped** (indirectly) | Used by `resolveCableType()` as third-priority heuristic for `Cable.Type` |
| `CableType` | `string` | `Cable.Type` | **Mapped** | Primary source for `Cable.Type` via `resolveCableType()` |
| `Length` | `*float64` | `Cable.Length` | **Mapped** | Truncated to `int` at export |
| `LengthUnit` | `string` | `Cable.LengthUnit` | **Mapped** | m, cm, ft, in |
| `Weight` / `WeightUnit` | `float64` / `string` | — | Cani-Internal | No Nautobot equivalent |
| `Color` | `string` | `Cable.Color` | **Mapped** | RGB hex string, mapped when non-empty |
| `Status` | `string` | `Cable.Status` (FK) | **Mapped** | "connected"→"Connected", "planned"→"Planned" |
| `TerminationA` | `uuid.UUID` | `Cable.TerminationAId` (FK) | **Mapped** | Resolved through device→interface chain |
| `TerminationAType` | `string` | `Cable.TerminationAType` | Partial | Hardcoded to `"dcim.interface"` |
| `TerminationB` | `uuid.UUID` | `Cable.TerminationBId` (FK) | **Mapped** | Same chain |
| `TerminationBType` | `string` | `Cable.TerminationBType` | Partial | Hardcoded to `"dcim.interface"` |
| `TerminationADevice` | `uuid.UUID` | — | Cani-Internal | Convenience: resolved to interface at export |
| `TerminationBDevice` | `uuid.UUID` | — | Cani-Internal | Convenience: resolved to interface at export |
| `TerminationAPort` | `string` | — | Cani-Internal | Port name; used for fuzzy interface lookup |
| `TerminationBPort` | `string` | — | Cani-Internal | Same |
| `Tags` | `[]string` | `Cable.Tags` | Not Mapped | |
| `CustomFields` | `map[string]any` | `Cable.CustomFields` | Not Mapped | |
| `Source` | `string` | — | Cani-Internal | |

**Cable type derivation** (`resolveCableType()` in `load.go`):

| Priority | Source | Behavior |
|---|---|---|
| 1 | `cable.CableType` | Direct lookup in `cableTypeMap` (full Nautobot enum coverage) |
| 2 | `cable.CableCategory` | Lookup in `cableTypeMap` (same table, case-insensitive) |
| 3 | `cable.ConnectorType` | Heuristic via `connectorToCableType` (e.g. `"rj45"` → `cat6`) |
| 4 | `cable.Slug` | Legacy fallback: substring matching (`"cat"` → `cat5e`, etc.) |

---

### 2.6 `CaniFruType` → Nautobot `InventoryItem`

Source: `pkg/devicetypes/cani_fru_types.go`
Exported in Phase 5 by `loadFrus()` in `load_frus.go`. Uses a topological sort (BFS from roots) to ensure parent FRUs are created before nested children. Each FRU maps to a single Nautobot `InventoryItem`.

| Cani Field | Go Type | Nautobot Field | Status | Notes |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | — | Cani-Internal | Primary key |
| `Name` | `string` | `InventoryItem.Name` | **Mapped** | |
| `Slug` | `string` | — | Cani-Internal | Library lookup key |
| `PartNumber` | `string` | `InventoryItem.PartId` | **Mapped** | Maps to Nautobot `part_id` |
| `Manufacturer` | `string` | `InventoryItem.Manufacturer` (FK) | **Mapped** | FK resolved by name |
| `Model` | `string` | — | Cani-Internal | Library/template field |
| `Description` | `string` | `InventoryItem.Description` | **Mapped** | Mapped when non-empty |
| `HardwareType` | `string` | — | Cani-Internal | Cani classification |
| `Weight` / `WeightUnit` | `float64` / `string` | — | Cani-Internal | No Nautobot equivalent |
| `Label` | `string` | `InventoryItem.Label` | **Mapped** | Mapped when non-empty |
| `Serial` | `string` | `InventoryItem.Serial` | **Mapped** | Mapped when non-empty |
| `AssetTag` | `string` | `InventoryItem.AssetTag` | **Mapped** | Mapped when non-empty |
| `Role` | `string` | — | Cani-Internal | Cani-specific; no Nautobot InventoryItem role |
| `Status` | `string` | — | Cani-Internal | No Nautobot InventoryItem status |
| `Device` | `uuid.UUID` | `InventoryItem.Device` (FK) | **Mapped** | Resolved to Nautobot device ID via cache |
| `Parent` | `uuid.UUID` | `InventoryItem.Parent` (FK) | **Mapped** | Resolved to parent InventoryItem ID (topological sort ensures order) |
| `Discovered` | `bool` | `InventoryItem.Discovered` | **Mapped** | |
| `Tags` | `[]string` | `InventoryItem.Tags` | Not Mapped | |
| `CustomFields` | `map[string]any` | `InventoryItem.CustomFields` | Not Mapped | |
| `Source` | `string` | — | Cani-Internal | |

---

## 3. Template vs. Instance Fields

Racks, Devices, and Modules carry both **template fields** (from the YAML library, used to create Nautobot type objects) and **instance fields** (from the user's inventory, used to create Nautobot instances). Understanding the split is critical for the mapper.

### Device: template → `DeviceType`, instance → `Device`

| Template Fields (→ DeviceType) | Instance Fields (→ Device) |
|---|---|
| `Model`, `Manufacturer`, `PartNumber` | `Name`, `Serial`, `AssetTag` |
| `UHeight`, `IsFullDepth`, `SubdeviceRole` | `Status`, `Role`, `Platform` |
| `Interfaces`, `ConsolePorts`, `PowerPorts` | `Rack`, `Location`, `RackPosition`, `Face` |
| `ModuleBays`, `DeviceBays` | `Parent`, `Children`, `ParentDevice` |
| `Identifications` | `Tenant`, `Tags`, `CustomFields` |

### Module: template → `ModuleType`, instance → `Module`

| Template Fields (→ ModuleType) | Instance Fields (→ Module) |
|---|---|
| `Model`, `Manufacturer`, `PartNumber` | `Name`, `Serial`, `AssetTag` |
| `Interfaces` | `ParentDevice`, `ModuleBayName` |
| | `Status`, `Role`, `Location` |

### Rack: no separate Nautobot type object

| Template Fields (informational only) | Instance Fields (→ Rack) |
|---|---|
| `Manufacturer`, `Model`, `PartNumber` | `Name`, `Location`, `Status`, `Role` |
| `DeviceBays`, `ModuleBays` | `UHeight`, `OuterWidth`, `OuterDepth` |
| | `RackType`, `Serial`, `AssetTag` |

### Cable & FRU: no Nautobot type object

YAML library definitions populate instance fields directly. The slug resolves to a definition that pre-fills `CableCategory`, `ConnectorType`, `Length`, `Color` (cables) or `Manufacturer`, `Model`, `Description` (FRUs). No separate type object is created in Nautobot.

---

## 4. Relationship Model

All relationships are expressed as single-direction FKs on the child object — matching Nautobot's model. Reverse pointers (`Children`, `Racks`, `Devices`, `OccupiedSlots`) are **rebuilt at load time** by `VerifyParentChildRelationships()` in `inventory_relationships.go`.

### 4.1 Relationship Table

| Relationship | Child Field (source of truth) | Nautobot FK | Cani Reverse (rebuilt) |
|---|---|---|---|
| Location → Location | `CaniLocationType.Parent` | `Location.Parent` | `CaniLocationType.Children` |
| Rack → Location | `CaniRackType.Location` | `Rack.Location` | `CaniLocationType.Racks` |
| Device → Rack | `CaniDeviceType.Parent` (when target is a rack) | `Device.Rack` + `Device.Position` + `Device.Face` | `CaniRackType.Devices` + `.OccupiedSlots` |
| Device → Device | `CaniDeviceType.Parent` (when target is a device) | `Device.ParentBay` | `CaniDeviceType.Children` |
| Module → Device | `CaniModuleType.ParentDevice` + `.ModuleBayName` | `Module.Device` + `Module.ModuleBay` | — |
| FRU → Device | `CaniFruType.Device` | `InventoryItem.Device` | — |
| FRU → FRU | `CaniFruType.Parent` | `InventoryItem.Parent` | — |
| Cable → Interface | `CaniCableType.TerminationA/B` + `TerminationA/BDevice` + port name | `Cable.TerminationAId/BId` + content types | — |

### 4.2 The Overloaded `Parent` Field

`CaniDeviceType.Parent` is overloaded — it can point to either a rack UUID or another device UUID. The mapper resolves this with a sequential check:

1. Look up `Parent` in `inventory.Racks` → if found, map to `Device.Rack`
2. Look up `Parent` in `inventory.Devices` → if found, map to `Device.ParentBay`
3. If neither, report error

The `rebuildDeviceRelationships()` function in `inventory_relationships.go` performs the same logic at load time, setting the explicit FK fields:
- `device.Rack = device.Parent` (when parent is a rack)
- `device.ParentDevice = device.Parent` (when parent is a device)
- `device.Location` is inherited from the rack or parent device

### 4.3 Relationship Verification

`VerifyParentChildRelationships()` runs five phases:

1. **`rebuildLocationRelationships()`** — Clears and rebuilds `Location.Children` from `Parent`
2. **`rebuildRackRelationships()`** — Clears and rebuilds `Location.Racks` from `Rack.Location`
3. **`rebuildDeviceRelationships()`** — Clears and rebuilds `Rack.Devices`, `Device.Children`, and explicit FK fields from `Device.Parent`
4. **`validateModuleRelationships()`** — Verifies `Module.ParentDevice` exists in `Devices`
5. **`validateFruRelationships()`** — Verifies `FRU.Device` and `FRU.Parent` exist
6. **`validateCableRelationships()`** — Verifies cable termination devices and interfaces exist
7. **`detectCircularLocationRefs()`** — Cycle detection on location parent chains
8. **`detectCircularDeviceRefs()`** — Cycle detection on device parent chains

### 4.4 Hierarchy Diagram

```
CaniLocationType (site/building/floor/room)
  └─ CaniLocationType.Parent → parent Location
  └─ owns → CaniRackType (via Rack.Location)
       └─ contains → CaniDeviceType (via Device.Parent → rack)
            ├─ children → CaniDeviceType (via Device.Parent → device)  [chassis→blade]
            ├─ modules → CaniModuleType (via Module.ParentDevice)
            ├─ frus → CaniFruType (via FRU.Device)
            │    └─ nested → CaniFruType (via FRU.Parent)
            └─ interfaces → InterfaceSpec[] (embedded, not normalized)
                 └─ connected by → CaniCableType (via Cable.TerminationA/B)
```

---

## 5. Export Pipeline

The seven-phase ETL pipeline is orchestrated by `Load()` in `pkg/provider/nautobot/load.go`:

### Phase 0: Locations

Implemented in `loadLocations()` in `load_locations.go`. Iterates `inventory.Locations` in topological order (BFS from roots, parents before children). For each location:
- Resolves `LocationType` by name (defaults to `"Site"` if empty), auto-creates if `create_locations` enabled
- Resolves `Parent` FK from previously created locations
- Resolves `Status` (`location.Status` → provider default → `"Active"`)
- Maps all optional fields: `Facility`, `Description`, `PhysicalAddress`, `ShippingAddress`, `ContactName`, `ContactPhone`, `ContactEmail`, `TimeZone`, `Latitude`, `Longitude`, `Asn`, `Comments`, `CustomFields`
- Creates via `DcimLocationsCreate` and caches result for downstream rack/device FK resolution

### Phase 1: Racks

Iterates `inventory.Racks`. For each rack:
- Resolves location by name (default or `"Default"` auto-created)
- Resolves status (`rack.Status` → default → `"Active"`)
- Creates via `DcimRacksCreate` with `Name`, `Location`, `Status`, `UHeight`, `OuterWidth`, `OuterDepth`, `Comments`
- Also checks `inventory.Devices` for legacy rack-type devices (fallback)

### Phase 2: Devices

Iterates `inventory.Devices` where `ClassifyForNautobot(device.HardwareType) == CategoryDevice`. For each:
- Checks if device already exists by name
- If exists and `--merge`: updates via `DcimDevicesPartialUpdate`
- If exists and no `--merge`: skips with conflict info
- If new: creates via `DcimDevicesCreate` using `MapToWritableDeviceRequest()`
- Maps `Comments` and `Face` (via `resolveFace()`) in both create and patch paths

### Phase 3: Interfaces

For each device created/found in Phase 2:
- Gets interface specs from `device.Interfaces` (from YAML library) or falls back to hardcoded defaults based on `HardwareType`
- Creates each interface via `DcimInterfacesCreate` with device FK, name, type, status
- Caches created interface IDs for cable creation

### Phase 4: Modules

Implemented in `loadModules()` in `load_modules.go`. Iterates `inventory.Modules`. For each module:
- Resolves parent device Nautobot ID from cache
- Gets or creates `ModuleType` (from YAML library: `Model`, `Manufacturer`, `PartNumber`, `Comments`)
- Gets or creates `ModuleBay` on the parent device (by `ModuleBayName`)
- Creates `Module` with `ModuleType` FK, `ParentModuleBay` FK, `Status`, and optional `Serial`, `AssetTag`, `Role`, `Location`
- Creates via `DcimModulesCreate`

### Phase 5: FRUs

Implemented in `loadFrus()` in `load_frus.go`. Iterates `inventory.Frus` in topological order (BFS from roots). For each FRU:
- Resolves parent device Nautobot ID from cache
- Resolves `Manufacturer` FK by name
- Maps `PartNumber` → `PartId`, `Description`, `Label`, `Serial`, `AssetTag`, `Discovered`
- Resolves `Parent` FK for nested FRUs from previously created InventoryItems
- Creates via `DcimInventoryItemsCreate`

### Phase 6: Cables

Iterates `inventory.Cables`. For each cable:
- Resolves both termination devices from `inventory.Devices`
- Looks up Nautobot device IDs from Phase 2 results
- Finds interface IDs via fuzzy name matching through the cache
- Checks for existing cables to avoid duplicates
- Maps `Color` (RGB hex) when non-empty
- Resolves `Type` via `resolveCableType()`: `CableType` → `CableCategory` → `ConnectorType` → slug fallback
- Creates via `DcimCablesCreate` with termination FKs, label, type, length, color, status

---

## 6. Classification Routing

`ClassifyForNautobot()` in `registry.go` determines how each hardware type is handled during export:

| CANI `Type` Constants | Category | Export Behavior |
|---|---|---|
| `rack`, `cabinet` | `CategoryRack` | Phase 1 → Nautobot `Rack` |
| `chassis`, `blade`, `node`, `nodecard`, `switch`, `mgmt-switch`, `hsn-switch`, `cabinet-pdu`, `cdu` | `CategoryDevice` | Phase 2 → Nautobot `Device` |
| `nic`, `gpu`, `cpu`, `memory`, `power-supply` | `CategoryModule` | Phase 4 → Nautobot `Module` + `ModuleType` + `ModuleBay` |
| `cable` | `CategoryCable` | Phase 6 → Nautobot `Cable` |
| `fru` | `CategoryFru` | Phase 5 → Nautobot `InventoryItem` |

---

## 7. Improvement Suggestions

### High Impact — IMPLEMENTED

1. **~~Export `CaniLocationType` properly~~** — ✅ Implemented in `load_locations.go`. Topological sort walks locations root-first. `LocationType` resolved by name (defaults to `"Site"`). Parent FKs wired. All optional fields mapped.

2. **~~Export `CaniModuleType`~~** — ✅ Implemented in `load_modules.go`. Creates `ModuleType` + `ModuleBay` + `Module`. Template fields → `ModuleType`, instance fields → `Module`. Parent device bay resolved dynamically.

3. **~~Export `CaniFruType`~~** — ✅ Implemented in `load_frus.go`. Topological sort for nested FRUs. Maps `PartNumber` → `PartId`, `Manufacturer` → FK, `Device` → FK, `Parent` → FK, `Description`, `Label`.

4. **~~Map free-win fields~~** — ✅ All wired:
   - `CaniRackType.OuterWidth/OuterDepth` → `Rack.OuterWidth/OuterDepth` (in `createRackFromCaniRack()`)
   - `CaniRackType.Comments` → `Rack.Comments` (in `createRackFromCaniRack()`)
   - `CaniCableType.Color` → `Cable.Color` (in `createCaniCableType()`)
   - `CaniDeviceType.Comments` → `Device.Comments` (in `MapToWritableDeviceRequest()`, `MapToNautobotDevice()`, `MapToPatchRequest()`)
   - `CaniDeviceType.Face` → `Device.Face` via `resolveFace()` (in `MapToWritableDeviceRequest()`, `MapToPatchRequest()`)
   - `CaniFruType.Description` → `InventoryItem.Description` (in `createFruFromCani()`)
   - `CaniFruType.Label` → `InventoryItem.Label` (in `createFruFromCani()`)

### Medium Impact — IMPLEMENTED

5. **~~Promote `Role` from `ProviderMetadata` to first-class field~~** — ✅ Implemented in `resolveRole()` in `mapper.go`. Now checks `device.Role` first, falls back to `ProviderMetadata["role"]` for backwards compatibility, then to provider default, then `"Generic"`.

6. **~~Add `SubdeviceRole` to DeviceType auto-creation~~** — ✅ Implemented in `CreateDeviceTypeFromLocal()` in `lookup.go`. Maps `"parent"` → `nautobotapi.Parent` and `"child"` → `nautobotapi.Child` via `ParentChildStatus.FromSubdeviceRoleEnum()`.

7. **~~Improve cable type mapping~~** — ✅ Implemented via `resolveCableType()` in `load.go`. Four-tier resolution: (1) explicit `CableType` field → direct enum lookup, (2) `CableCategory` → enum lookup, (3) `ConnectorType` → heuristic mapping, (4) legacy slug-based fallback. Full `CableTypeChoices` coverage via `cableTypeMap` and `connectorToCableType` tables.

8. **~~Use actual `Face` value~~** — ✅ Implemented via `resolveFace()` in `mapper.go`. Maps `"front"` → `FaceEnumFront`, `"rear"` → `FaceEnumRear`, empty → nil (Nautobot default).

### Lower Priority

9. **~~Normalize interfaces into `Inventory.Interfaces`~~** — ✅ Implemented. Added `Interfaces map[uuid.UUID]*CaniInterface` to `Inventory` struct in `inventory.go`. `rebuildInterfaceRelationships()` in `inventory_relationships.go` populates the map from device and module `Interfaces` slices during `VerifyParentChildRelationships()`. `GetInterfaceByID()` in `inventory_queries.go` rewritten from O(n×m) scan to O(1) map lookup. Added `GetInterfacesByDevice()` helper. `CaniInterface` YAML tags fixed from PascalCase to snake_case.

10. **YAML/JSON tag style — NetBox devicetype-library parity (do NOT "standardize" to snake_case)** — The devicetype YAML schema mirrors the upstream NetBox devicetype-library, which is the source of truth. Component-collection keys are **kebab-case** and MUST stay that way: `console-ports`, `power-ports`, `module-bays`, `device-bays`, `allowed-children`, `hardware-type`. 50+ on-disk library files (`device-types/`, `rack-types/`, `module-types/`, …) and the `yaml:"…"` struct tags use these exact kebab keys, so renaming them to snake_case breaks loading the entire hardware library. NetBox scalar fields and entry sub-fields keep their upstream snake_case names (`part_number`, `u_height`, `is_full_depth`, `subdevice_role`, `weight_unit`, `mgmt_only`, `maximum_draw`); JSON tags are camelCase. This split is intentional NetBox parity, not a tag-convention violation.

11. **~~Clean stale netbox references~~** — ✅ Implemented. Removed 3 stale Makefile targets (`netbox-dt-schema`, `netbox-mt-schema`, `nbschema`) that referenced the deleted `pkg/netbox/` directory. No stale Go import references remained. AGENTS.md mentions of "NetBox" refer to the schema origin and are valid documentation.

---

## 8. Key File Index

| Purpose | Path |
|---|---|
| Inventory struct | `pkg/devicetypes/inventory.go` |
| Inventory queries | `pkg/devicetypes/inventory_queries.go` |
| CaniType interface | `pkg/devicetypes/cani_type.go` |
| CaniLocationType | `pkg/devicetypes/cani_location_types.go` |
| CaniRackType | `pkg/devicetypes/cani_rack_types.go` |
| CaniDeviceType | `pkg/devicetypes/cani_device_types.go` |
| CaniModuleType | `pkg/devicetypes/cani_module_types.go` |
| CaniCableType | `pkg/devicetypes/cani_cable_types.go` |
| CaniFruType | `pkg/devicetypes/cani_fru_types.go` |
| Component specs | `pkg/devicetypes/component_specs.go` |
| Type enums & classification | `pkg/devicetypes/registry.go` |
| Relationship rebuild logic | `pkg/devicetypes/inventory_relationships.go` |
| Lookup functions | `pkg/devicetypes/all.go` |
| Nautobot mapper | `pkg/provider/nautobot/mapper.go` |
| Nautobot export pipeline | `pkg/provider/nautobot/load.go` |
| Nautobot location export | `pkg/provider/nautobot/load_locations.go` |
| Nautobot module export | `pkg/provider/nautobot/load_modules.go` |
| Nautobot FRU export | `pkg/provider/nautobot/load_frus.go` |
| Nautobot lookup cache | `pkg/provider/nautobot/lookup.go` |
| Generated Nautobot API client | `pkg/nautobot/nautobot_api.go` |
