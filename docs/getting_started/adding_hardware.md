# Adding Hardware

Systems change over time and adding hardware is often needed to expand the available resources. The `add` command supports locations, racks, devices, modules, and cables.

## Add A Location

```shell
# Add a location (data center, room, row, etc.)
cani alpha add location my-datacenter
```

## Add A Rack

```shell
# List available rack types
cani alpha add rack --list-supported-types

# Add a rack, accepting recommended values
cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --auto --accept
```

## Add A Device

```shell
# List available device types
cani alpha add device --list-supported-types

# Add a compute blade into a rack
cani alpha add device hpe-crayex-ex420-compute-blade --auto --accept
```

## Add A Module

```shell
# Add a module into a device
cani alpha add module hpe-crayex-ex420-gpu-module --auto --accept
```

## Show The Inventory

Recent additions can be viewed with the `show` command:

```shell
# Show all racks
cani alpha show rack

# Show all devices
cani alpha show device
```
