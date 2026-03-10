# Device Type Data Model

The `cani` device-types library follows the [Nautobot device-type](https://github.com/nautobot/nautobot-app-device-type-library) schema.
Each YAML definition maps directly to the Nautobot data model, making inventory data portable between `cani` and Nautobot.

## Type Categories

Device types are organized into subdirectories that map to inventory types:

| Directory | Inventory Type | Go Type |
|-----------|---------------|---------|
| `device-types/` | Device | `CaniDeviceType` |
| `rack-types/` | Rack | `CaniRackType` |
| `module-types/` | Module | `CaniModuleType` |
| `cable-types/` | Cable | `CaniCableType` |
| `inventory-types/` | FRU | `CaniFruType` |

## Nautobot Data Model Reference

| Resource | Nautobot Documentation |
|----------|------------------------|
| Locations | [Location Model](https://docs.nautobot.com/projects/core/en/stable/user-guide/core-data-model/dcim/location/) |
| Racks | [Rack Model](https://docs.nautobot.com/projects/core/en/stable/user-guide/core-data-model/dcim/rack/) |
| Devices | [Device Model](https://docs.nautobot.com/projects/core/en/stable/user-guide/core-data-model/dcim/device/) |
| Device Types | [Device Type Model](https://docs.nautobot.com/projects/core/en/stable/user-guide/core-data-model/dcim/devicetype/) |
| Modules | [Module Model](https://docs.nautobot.com/projects/core/en/stable/user-guide/core-data-model/dcim/module/) |
| Cables | [Cable Model](https://docs.nautobot.com/projects/core/en/stable/user-guide/core-data-model/dcim/cable/) |
| FRUs | [Inventory Item Model](https://docs.nautobot.com/projects/core/en/stable/user-guide/core-data-model/dcim/inventoryitem/) |

## YAML Schema

Each device type definition includes:

| Field | Description |
|-------|-------------|
| `manufacturer` | Vendor name (e.g. `HPE`, `Cray`) |
| `model` | Model identifier (e.g. `ProLiant DL380 Gen10`) |
| `slug` | URL-safe unique key (e.g. `hpe-proliant-dl380-gen10`) |
| `hardware-type` | Classification (`Device`, `Rack`, `Module`, `Cable`) |
| `u_height` | Rack units consumed (devices and racks) |
| `interfaces` | Network interface definitions |
| `console-ports` | Serial/console port definitions |
| `power-ports` | Power input definitions |
| `module-bays` | Slots for installable modules |
| `device-bays` | Slots for child devices |

## Listing Available Types

Use the `-L` flag on any `add` subcommand to see what device types are currently loaded:

```shell
# List all device types
cani alpha add device -L

# List all rack types
cani alpha add rack -L

# List all module types
cani alpha add module -L
```

See [Customizing](customizing.md) for creating new definitions and [Extending](extending.md) for adding external type sources.
