# Add

The `add` command creates new hardware in the local inventory. All add operations work the same way regardless of which provider the data was imported from. This makes it possible to import from one system, add new hardware locally, and export the result to any provider.

## Supported Types

```shell
cani alpha add location <name>
cani alpha add rack <device-type-slug> [flags]
cani alpha add device <device-type-slug> [flags]
cani alpha add module <device-type-slug> [flags]
```

## Common Flags

| Flag | Description |
|------|-------------|
| `--list-supported-types` | List available hardware type slugs |
| `--auto` | Automatically determine placement (parent, position) |
| `--accept` | Accept recommended values without confirmation |

## Examples

```shell
# Add a rack
cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --auto --accept

# Add a compute blade
cani alpha add device hpe-crayex-ex420-compute-blade --auto --accept

# Add a module into a device
cani alpha add module hpe-crayex-ex420-gpu-module --auto --accept
```

## Provider Independence

Added hardware exists only in the local inventory until it is exported. Because the inventory format is provider-agnostic, hardware added after importing from CSM can be exported to Nautobot (or any other provider) without modification.

```shell
# Import existing infrastructure from CSM
cani alpha import csm ...

# Add new hardware locally
cani alpha add device hpe-crayex-ex420-compute-blade --auto --accept

# Export the combined inventory to Nautobot
cani alpha export nautobot --url http://localhost:8081 --token $TOKEN
```
