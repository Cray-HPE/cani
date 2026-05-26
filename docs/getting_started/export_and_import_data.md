# Visual Output

The `show` command displays inventory data with sorting, filtering, and visual hierarchy options.

## Show Inventory

```shell
# Show all devices
cani alpha show device

# Show all racks
cani alpha show rack

# Show all locations
cani alpha show location

# Show cables
cani alpha show cables
```

## Filtering

```shell
# Show devices filtered by role
cani alpha show device --role Compute

# Show devices filtered by type
cani alpha show device --type blade
```

## Visual Hierarchy

The show command can render a tree view of the inventory:

```shell
# Show a tree view of all inventory
cani alpha show device --tree
```

