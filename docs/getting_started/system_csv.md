# System CSV

A single CSV file that defines your entire system: roles, racks, devices, modules, and cable connections.

## Header

Row 1 is always the same 21-column header:

```
Section,PartNumber,Name,Qty,Rack,Position,Face,Role,Status,Serial,Device,Bay,ADevice,APort,BDevice,BPort,Color,Length,LengthUnit,Location,ContentTypes
```

## Sections

Each data row starts with a **Section** value that determines which columns matter:

| Section | Key columns | Notes |
|---|---|---|
| `_defaults` | `Status` (or any field) | Optional single row; values fill in blanks for all rows |
| `role` | `Name`, `ContentTypes` | Define roles before devices reference them. ContentTypes is comma-separated (e.g. `dcim.device,dcim.rack`) |
| `location` | `Name`, `Role`, `Location`, `ContentTypes` | `Role` = location type slug (`dc`, `level`, `section`). `Location` = parent location name. Define parents before children |
| `rack` | `PartNumber`, `Name`, `Qty`, `Status`, `Location` | `PartNumber` = rack-type slug. `Location` = location name to assign rack to |
| `device` | `PartNumber`, `Name`, `Qty`, `Rack`, `Position`, `Face`, `Role`, `Status`, `Serial` | `Face` = `front` or `rear`. `Position` = U number (omit for zero-U like PDUs) |
| `module` | `PartNumber`, `Qty`, `Device`, `Bay` | `Device` = parent device name, `Bay` = module bay name from device type (e.g. `GPU0`, `PCIe5`) |
| `connection` | `PartNumber`, `ADevice`, `APort`, `BDevice`, `BPort`, `Color`, `Length`, `LengthUnit` | `PartNumber` = cable-type slug. Ports support brace expansion: `HSN {0..3}` |

## Rules

- **Comments**: Lines starting with `#` are ignored
- **Order**: roles → locations → racks → devices → modules → connections (each section can reference items from earlier sections)
- **Defaults**: The `_defaults` row fills blanks globally. You can also use section-specific defaults like `device_defaults` or `connection_defaults`
- **Qty**: Defaults to 1 if omitted. For racks, Qty > 1 auto-generates `name-1`, `name-2`, etc.
- **Empty columns**: Leave unused columns empty — every row must have 21 comma-separated fields
- **Brace expansion**: Connection ports like `{0..3}` expand to `0,1,2,3`. Both A and B sides expand in lockstep (zip)

## Minimal Example

```csv
Section,PartNumber,Name,Qty,Rack,Position,Face,Role,Status,Serial,Device,Bay,ADevice,APort,BDevice,BPort,Color,Length,LengthUnit,Location,ContentTypes
_defaults,,,,,,,,Active,,,,,,,,,,,,
role,,ComputeNode,,,,,,,,,,,,,,,,,,dcim.device
location,,MyDC,,,,,dc,,,,,,,,,,,,,
location,,Floor1,,,,,level,,,,,,,,,,,,MyDC,
location,,Zone-A,,,,,section,,,,,,,,,,,,Floor1,"rack,device,module"
rack,hpe-48u-800mmx1200mm-g2-enterprise-shock-rack,rack1,1,,,,,Available,,,,,,,,,,,Zone-A,
device,hpe-xd670,node1,1,rack1,34,front,ComputeNode,Active,,,,,,,,,,,,
module,nvidia-h100-sxm-gpu,,1,,,,,,,node1,GPU0,,,,,,,,,
connection,hpe-3m-cat6-stp,,,,,,,,,,,node1,iLO,switch1,1,blue,3,m,,
```

## Finding Valid PartNumbers

`PartNumber` values are slugs from the device-type library under `pkg/devicetypes/`. Browse the YAML files there for available rack types, device types, module types, and cable types.
