# AGENTS.md — HPCM Transform

## Parent-Child Hierarchy

The transform pipeline produces the following device hierarchy from HPCM node data:

```
Rack
 └── Chassis (Device)
      └── Module
           └── FRU
```

### Hierarchy Details

| Child    | Parent Field      | Points To  |
|----------|-------------------|------------|
| Chassis  | `dev.Parent`      | Rack ID    |
| Module   | `mod.ParentDevice`| Chassis ID |
| FRU      | `fru.Parent`      | Module ID (if from module) or Device ID |

> **Note:** Module FRUs point to the **module** (`mod.ID`) that owns them,
> establishing the full hierarchy: rack → chassis → module → FRU.
> Device FRUs point directly to the device.

### Geoloc Parent Resolution

When a module node has an `inventory.geoloc` key (or `aliases["cm-geo-name"]`)
containing an xname like `x9000c1s7b0n0`, the transform derives the parent
chassis xname (`x9000c1`) and uses it as a fallback for parent resolution
when `location.rack`/`location.chassis` are not available.

### Two-Pass Transform

1. **Pass 1 — Chassis devices only**: Iterates all nodes, creates chassis
   devices, assigns them to racks via `assignRack`, and registers their UUIDs
   in `chassisByLoc` so modules can reference them in Pass 2.
2. **Pass 2 — Everything else**: Iterates all non-chassis nodes, classifies
   each as location/device/module via `classifyNode`, builds modules with
   `buildModuleFromNode` (which looks up the parent chassis from
   `chassisByLoc`), and collects FRUs via `buildFrusFromInventory`.

### Rack Deduplication

`buildRack` deduplicates by rack number using `racksByNumber`. Multiple chassis
in the same physical rack (e.g. chassis 1 and chassis 3 both in rack 9000)
produce a single rack entry.

### Where HPCM's Own Data Goes

Everything HPCM records about an object that the inventory has no field for
goes in `ProviderMetadata["hpcm"]` (`metadata.go`): a node's UUID, aliases and
location on devices and modules, a module's geoloc, a rack's number, a
location's UUID, and every inventory field of an inventory item that is not
its serial, part number, model, manufacturer or name. Nothing is written to
`CustomFields`, which hold attributes an operator defines and which every
export sends, and nothing is written to `ProviderMetadata` outside the bucket.
