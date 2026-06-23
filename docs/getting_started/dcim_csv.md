
# DCIM CSV

The `example` provider's **DCIM CSV** is a single multi-section CSV that
describes a whole inventory — roles, statuses, locations, racks, devices,
modules, interface metadata, and cabling — in one file. It is the reference
format for bootstrapping inventories and exercising the full ETL pipeline.

```shell
cani alpha import example --csv ./dcim.csv
```

A file is treated as a DCIM CSV (rather than a flat BOM CSV) when its header
contains a `Section` column. Detection lives in `IsDcimCSV` and parsing in
`ParseDcimCSV` (`pkg/provider/example/import/dcim_csv.go`); transformation in
`TransformDcim` (`pkg/provider/example/transform/dcim.go`).

## Pipeline

```
import (ParseDcimCSV)  ->  transform (TransformDcim)  ->  merge  ->  datastore
        buckets by Section      passes -> TransformResult        cmd/import
```

| Phase | Code | Responsibility |
|-------|------|----------------|
| Extract | `import/dcim_csv.go` `ParseDcimCSV` | Group rows into per-section buckets |
| Transform | `transform/dcim.go` `TransformDcim` | Roles/Statuses → Locations → Racks → Devices → Modules → Interfaces → Connections |
| Load | `cmd/import` | Merge the `TransformResult` and save |

## Header

The header **names** the columns; the parser maps by column name
(order-independent, common aliases accepted), so a file only needs the columns
its sections use. The full set of recognized columns is:

```
Section,PartNumber,Name,Qty,Rack,Position,Face,Role,Status,Serial,Device,Bay,ADevice,APort,BDevice,BPort,Color,Length,LengthUnit,Location,ContentTypes,MacAddress,Description,LocationType
```

## Sections

Each data row starts with a **Section** value that selects the record type:

| Section | Key columns | Notes |
|---|---|---|
| `_defaults` | `Status` (or any field) | Optional single row; values fill blanks for all rows. Per-section `<section>_defaults` rows also work |
| `role` | `Name`, `ContentTypes`, `Color`, `Description` | Define roles before devices reference them. `ContentTypes` is comma-separated (e.g. `dcim.device,dcim.rack`) |
| `status` | `Name`, `ContentTypes`, `Color`, `Description` | First-class status catalog (e.g. `Active`, `Planned`) |
| `location` | `Name`, `LocationType`, `Location`, `ContentTypes` | `LocationType` = type slug (`dc`, `level`, `section`); `Role` is accepted as a legacy alias. `Location` = parent location name. Define parents before children |
| `rack` | `PartNumber`, `Name`, `Qty`, `Status`, `Location` | `PartNumber` = rack-type slug. `Location` = location name to assign the rack to |
| `device` | `PartNumber`, `Name`, `Qty`, `Rack`, `Position`, `Face`, `Role`, `Status`, `Serial` | `Face` = `front` or `rear`. `Position` = U number (omit for zero-U like PDUs) |
| `module` | `PartNumber`, `Qty`, `Device`, `Bay`, `Serial` | `Device` = parent device name, `Bay` = module bay name from the device type (e.g. `GPU0`, `PCIe5`) |
| `interface` | `Device`, `Name`, `MacAddress` | Annotates an existing device/module interface with a MAC address |
| `connection` | `PartNumber`, `ADevice`, `APort`, `BDevice`, `BPort`, `Color`, `Length`, `LengthUnit` | `PartNumber` = cable-type slug. Ports support brace expansion: `HSN {0..3}` |

## Rules

- **Comments**: lines starting with `#` are ignored.
- **Order**: roles/statuses → locations → racks → devices → modules → interfaces → connections (each section can reference items from earlier sections, and hardware imported in an earlier run).
- **Content types**: bare model names are normalized to Nautobot form — `device` → `dcim.device`, `prefix` → `ipam.prefix`; already-qualified values pass through.
- **Defaults**: the `_defaults` row fills blanks globally; `<section>_defaults` (e.g. `device_defaults`) overrides for one section.
- **Qty**: defaults to 1 if omitted. For `Qty > 1`, names get `-1`, `-2`, … suffixes.
- **Empty columns**: leave unused columns empty; rows are matched by header name, not position.
- **Brace expansion**: connection ports like `{0..3}` expand to `0,1,2,3`. Both A and B sides expand in lockstep (zip).

## Minimal Example

```csv
Section,PartNumber,Name,Qty,Rack,Position,Face,Role,Status,Serial,Device,Bay,ADevice,APort,BDevice,BPort,Color,Length,LengthUnit,Location,ContentTypes,MacAddress
_defaults,,,,,,,,Active,,,,,,,,,,,,,
role,,ComputeNode,,,,,,,,,,,,,,,,,,,dcim.device,
status,,Active,,,,,,,,,,,,,,,,,,,dcim.device,
location,,MyDC,,,,,dc,,,,,,,,,,,,,,
location,,Zone-A,,,,,section,,,,,,,,,,,,MyDC,"dcim.rack,dcim.device",
rack,hpe-48u-800mmx1200mm-g2-enterprise-shock-rack,rack1,1,,,,,Available,,,,,,,,,,,Zone-A,,
device,hpe-xd670,node1,1,rack1,34,front,ComputeNode,Active,,,,,,,,,,,,,
module,nvidia-h100-sxm-gpu,,1,,,,,,,node1,GPU0,,,,,,,,,,
interface,,iLO,,,,,,,,node1,,,,,,,,,,,aa:bb:cc:dd:ee:01
connection,hpe-3m-cat6-stp,,,,,,,,,,,node1,iLO,switch1,1,blue,3,m,,,
```

## Finding Valid PartNumbers

`PartNumber` values are slugs from the device-type library under
`pkg/devicetypes/`. Browse the YAML files there for available rack types,
device types, module types, and cable types.
