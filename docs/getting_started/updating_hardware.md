# Updating Hardware

Some providers and certain types of hardware require additional metadata. The `update` command modifies existing inventory items.

## Update A Device

After importing or adding devices, they may need metadata like roles or aliases:

```shell
# Update a device's role and alias
cani alpha update device --role Compute --alias nid00001

# Update a device's NID
cani alpha update device --nid 1
```

## Update With Classification

Devices imported from Redfish or other sources may not have a device-type slug. Use `classify` first, then update:

```shell
# Import from Redfish
cani alpha import redfish --root ./redfish-roots.json

# Classify unrecognized devices
cani alpha classify --auto

# Update metadata
cani alpha update device --role Compute --alias nid00001 --nid 1
```

## Bulk Updates

The `update set` command applies changes to multiple items at once:

```shell
# Set the role for all devices matching a filter
cani alpha update set --role Compute
```

