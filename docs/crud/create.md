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

## Device Bays

Some hardware mounts inside another device instead of directly in a rack, for example two half-width switches in a 1U dual chassis. The parent device type defines `device-bays`; the child device type sets `subdevice_role: child`. Add the parent to the rack, then install each child into a bay with `--parent` and optionally `--bay`:

```shell
# 1U parent chassis in the rack
cani alpha add device nvidia-sn2100-dual-chassis --rack R0179 --position 15 --name sn2100-chassis-15

# Children go into the first free bay, or a named one
cani alpha add device nvidia-sn2100 --parent sn2100-chassis-15 --name sw-leaf-001
cani alpha add device nvidia-sn2100 --parent sn2100-chassis-15 --bay "Bay 2" --name sw-leaf-002

# Move an existing device into a bay
cani alpha update device sw-leaf-001 --parent sn2100-chassis-15 --bay "Bay 1"
```

`--parent` accepts a rack or device name or UUID. When the parent device has bays, the child must fit a free bay that allows its slug (`allowed` in the parent's device type); a child in a bay has no rack position or face of its own. The tree view (`cani alpha show rack <rack> -o tree`) shows the bay next to each child. On Nautobot export, each child is installed into the named bay on its parent, and the bay is created on the parent device if its device type has no matching bay template. On Nautobot import, devices installed in device bays come back as children of their parent device with the bay name.

Bay selection and local occupancy checks are case-insensitive, including existing
and pending assignments. Case differences do not make an occupied bay free.
Successful placements use the bay name declared by the device type; checking
occupancy does not rename existing assignments.

Rack placement strategies (`--rack '%{FILL}'` or `--rack '%{SPREAD}'`) cannot
be combined with a specific `--parent` or a named `--bay`. To install into a
device bay, use `--parent` and optionally `--bay` without a rack strategy.
Literal rack placement keeps its existing precedence: when both `--rack` and
`--parent` name items, `--rack` wins.

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
