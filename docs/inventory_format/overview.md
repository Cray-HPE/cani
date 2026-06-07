# Inventory Format Overview

The inventory is a versioned datastore that holds every piece of hardware `cani` knows about. Each hardware category is stored as a separate map keyed by UUID.

## Top-Level Structure

In Go the inventory is the `Inventory` struct:

```go
type Inventory struct {
    SchemaVersion string
    Provider      string

    Locations  map[uuid.UUID]*CaniLocationType
    Racks      map[uuid.UUID]*CaniRackType
    Devices    map[uuid.UUID]*CaniDeviceType
    Modules    map[uuid.UUID]*CaniModuleType
    Cables     map[uuid.UUID]*CaniCableType
    Frus       map[uuid.UUID]*CaniFruType
    Interfaces map[uuid.UUID]*CaniInterface

    Metadata   *InventoryMetadata // catalog of roles, statuses, tags
}
```

The current schema version is `v1alpha2`.

## Hardware Types

| Type | Go Type | Description |
|------|---------|-------------|
| Location | `CaniLocationType` | Physical site, building, floor, or room |
| Rack | `CaniRackType` | Equipment rack with U-slot tracking |
| Device | `CaniDeviceType` | Server, switch, PDU, chassis, or blade |
| Module | `CaniModuleType` | Component installed in a device (GPU, NIC, PSU) |
| Cable | `CaniCableType` | Physical cable between two endpoints |
| FRU | `CaniFruType` | Field-replaceable unit (spare or replacement part) |
| Interface | `CaniInterface` | Network or console port on a device or module |

## Relationships

Items reference each other by UUID:

- A **Location** has `Children` (child locations) and `Racks`.
- A **Rack** has a `Location` FK and a `Devices` list.
- A **Device** has `Parent`, `Children`, `Rack`, `Location`, and `Frus` references.
- A **Module** has a `ParentDevice` FK.
- A **Cable** has `TerminationA` and `TerminationB` endpoint UUIDs.
- A **FRU** has a `Device` or `Parent` FK.

## Example: Device

```json
{
  "f7448392-1e1c-45d0-9c59-be7dfc44c15c": {
    "id": "f7448392-1e1c-45d0-9c59-be7dfc44c15c",
    "name": "nid000001",
    "type": "Device",
    "manufacturer": "HPE",
    "model": "EX420 Compute Blade",
    "status": "active",
    "role": "Compute",
    "parent": "7e3de0fa-e3d6-421b-9d25-c0192d2a5966",
    "rack": "a1b2c3d4-0000-0000-0000-000000000001",
    "rackPosition": 3,
    "children": [
      "00050177-a309-4fde-bf85-70452b228e24",
      "004ecb7f-50bb-4975-9973-b6c617d6cc82"
    ],
    "providerMetadata": {
      "csm": {
        "xname": "x3000c0s3b0n0",
        "class": "Mountain",
        "role": "Compute",
        "nid": 1
      }
    },
    "externalIDs": {
      "nautobot": "9a8b7c6d-5e4f-3a2b-1c0d-000000000001"
    }
  }
}
```

## Example: Rack

```json
{
  "a1b2c3d4-0000-0000-0000-000000000001": {
    "id": "a1b2c3d4-0000-0000-0000-000000000001",
    "name": "x3000",
    "uHeight": 42,
    "location": "b2c3d4e5-0000-0000-0000-000000000001",
    "status": "active",
    "devices": [
      "f7448392-1e1c-45d0-9c59-be7dfc44c15c"
    ]
  }
}
```

## Common Fields

All types embed `ObjectMeta`, providing a uniform set of metadata fields:

- `id` — unique UUID identifier
- `name` — human-readable name
- `status` — lifecycle state (`staged`, `active`, `planned`, etc.)
- `role` — functional role (e.g. `Compute`, `Spine Switch`)
- `tags` — arbitrary string labels
- `tenant` — tenant or ownership group
- `customFields` — free-form key/value map
- `externalIDs` — maps a provider name to that provider's remote UUID
- `providerMetadata` — provider-specific data (see [Metadata](metadata.md))
