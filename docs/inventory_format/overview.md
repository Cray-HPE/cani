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

    Prefixes    map[uuid.UUID]*CaniPrefix
    IPAddresses map[uuid.UUID]*CaniIPAddress
    VLANs       map[uuid.UUID]*CaniVLAN
    VRFs        map[uuid.UUID]*CaniVRF

    Metadata *InventoryMetadata // catalog of roles, statuses, tags
}
```

The current schema generation is `v1alpha10`. It gives every value one home: a
modeled attribute is a field, an operator-defined attribute is a custom field,
and `providerMetadata` holds one bucket per provider and nothing else (see
[Provider metadata on upgrade](#provider-metadata-on-upgrade)). `v1alpha9`
records the owning module of a module interface (`moduleId`) and admits spare
modules, which are stored at a location with no parent device. `v1alpha8`
added optional device parent-bay placement and unresolved module-interface
name/label templates. Older inventories are migrated through each generation
on load, with the original file backed up to `.canisave`. The `v1alpha7` to
`v1alpha8` and `v1alpha8` to `v1alpha9` steps only advance the version: they
preserve existing values without inferring bay assignments or missing
templates, and module ownership is rebuilt from the module specs on every
load. Earlier cani releases refuse a newer inventory rather than load and
rewrite one that may contain data they do not preserve.

### Provider metadata on upgrade

Through `v1alpha9`, a value typed with `--metadata` was stored under
`providerMetadata.nautobot`, some values were stored in `providerMetadata`
outside any provider's bucket, and the Nautobot export read both: as custom
fields, and as a device's role and location. Loading a `v1alpha9` inventory
moves each of those values to the one place it is read from afterwards:

| Value | Where it goes |
| --- | --- |
| `role` on a device | the device's `role` field, when that field is empty |
| `location` on a device | the device's `location`, found by the location's name |
| `rack_number` outside a bucket | `providerMetadata.hpcm.rack_number`, the provider that wrote it |
| any other key | the object's `customFields`, replacing a custom field of the same key |

A device with an empty `role` field also takes the role the export read from
another provider's bucket, which keeps its copy. Other providers' buckets are
not changed otherwise.

A value moves only when it is the one an export was already sending, so the
upgrade never makes a value live that was not. Where a key was held more than
once, the export used one copy: the `role` field over provider metadata, a
key outside the buckets over a bucket for `role` and `location`, and for every
other key the holder whose name sorts last. The other copies are dropped.

Everything dropped, replaced or left unplaced is logged once, by object and
key and never by value, on the load that upgrades the file:

| Log line | Meaning | What to do |
| --- | --- | --- |
| `custom field "K" held another value; the provider metadata value took precedence and replaced it` | The custom field now holds what the export sent. | Nothing, unless the replaced value was the right one; it is in the `.canisave` backup and can be set again with `--custom-field`. |
| `provider metadata "K" was dropped: another copy of that key took precedence` | A copy the export never used. | Nothing, unless it was the right one; set it with `--custom-field`, or with `--role` or `--location` for those two keys. |
| `provider metadata "role" was dropped: it was not the role export sent` | The `role` field was already set, or the value was not text. | Set the role with `update device --role` if the dropped one was intended. |
| `provider metadata "location" replaced the location the device recorded, as it did on export` | The device had no parent and both placements were set. | Re-import, or set the location with `update device --location`, if the recorded one was right. |
| `provider metadata "location" was not applied: ...` | A device in a rack can only record a location that holds its rack's, and a device installed in another device takes that device's. | Nothing; Nautobot did not keep such a placement either. |
| `provider metadata "location" names no location in the inventory` (or `more than one`) | The name was looked up in Nautobot on export, and cannot be turned into a placement here. It is kept in `providerMetadata.nautobot.location`. | Add or import the location, then `update device --location`. |
| `provider metadata "location" was dropped: it is not a location name, so export never read it` | The value was not text. | Nothing. |
| `provider metadata "u_height"` or `"rack_position" was dropped: export never sent it` | Typed on a rack that is stored as a device. | Set the height on the rack itself with `update rack --set u_height=N`. |

A location stored on a device that has a parent is kept only when it is its
rack's. Earlier generations re-derived it on every load; from `v1alpha10` a
device in a rack may record a location of its own, so a stale stored value
would otherwise be read as one.

`--metadata` itself is gone. `--custom-field key=value` sets a custom field on
`add` and `update`, `--role` sets a role and `--location` a device's location.
A custom field has to be defined for the kind of item it is set on
(`add metadata custom-field`, or a definition imported from Nautobot). A key
the upgrade carried over without a definition is kept and exported as before;
define it to change its value with `--custom-field`, or remove it with
`update <noun> --clear-custom-field <key>`.

### Placement validation on load

From `v1alpha9` on, every load validates module and device-bay placement. A
module bay the parent device does not declare (on a parent that declares bays),
two authored claims on one bay, or a spare stored at a location that does not
exist is an error; a module with no bay and no name, a spare that still names
a bay, or an installed module whose authored location differs from its
parent's is a warning. Earlier releases accepted any bay name, so a `v1alpha8`
inventory can carry an error the new rules refuse. Such an inventory still
loads, but the upgrade is deferred: the file stays at `v1alpha8`, the errors
are reported, and every save is refused until they are repaired. Repair an
undeclared bay with `update module --bay <declared bay>`, or re-import from the
source system, whose declared bays are recorded on the device so a bay the
library spells differently becomes declared. The upgrade is written on the
first load after the repair.

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
| VRF | `CaniVRF` | Virtual routing and forwarding instance used by IPAM objects |

## Relationships

Items reference each other by UUID:

- A **Location** has `Children` (child locations) and `Racks`.
- A **Rack** has a `Location` FK and a `Devices` list.
- A **Device** has `Parent`, `Children`, `Rack`, `Location`, and `Frus` references.
- A **Module** has a `ParentDevice` FK, or, as a spare, a `Location` FK and no
  parent device.
- A **Cable** has `TerminationA` and `TerminationB` endpoint UUIDs.
- A **FRU** has a `Device` or `Parent` FK.

### Cable Endpoint Integrity

Each side of an inventory cable is either absent or complete. An absent side
has an empty device/module UUID, port name, and interface UUID. A populated side
must contain all three fields:

- `terminationADevice` or `terminationBDevice` identifies a device or module.
- `terminationAPort` or `terminationBPort` names an interface on that object.
- `terminationA` or `terminationB` is the UUID of that exact interface.

Relationship rebuilding fills a missing interface UUID when the referenced
device or module contains the named port. Validation fails if the port cannot be
resolved, if fields are only partially populated, or if an explicit interface
UUID belongs to another endpoint. Module-owned interfaces may be referenced
through the module UUID or found through their parent device.

All command saves validate relationships before writing. Import fails the whole
operation before its Load phase when transformed inventory is invalid, and
export performs the same preflight before invoking a provider. Existing files
remain loadable so operators can inspect and repair invalid data, but migration
does not automatically rewrite a file while relationship errors remain.

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
