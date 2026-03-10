# Classifying Devices

Devices imported from external sources (such as Redfish BMC data) may not have a recognized device-type slug. The `classify` command matches unclassified devices against the device-types library.

## Auto-Classify

```shell
# Automatically classify all unclassified devices
cani alpha classify --auto
```

Auto-classification uses manufacturer and model information from the imported data to find the best match in the device-types library.

## Interactive Classification

Without the `--auto` flag, each unclassified device is presented interactively with candidate matches ranked by similarity:

```shell
# Walk through each unclassified device
cani alpha classify
```

## Typical Workflow

Classification is usually needed after importing from Redfish or other raw sources:

```shell
# 1. Import from Redfish
cani alpha import redfish --root ./redfish-roots.json

# 2. Classify the imported devices
cani alpha classify --auto

# 3. Review the results
cani alpha show device
```

See [Device Type Data Model](../devicetypes/devicetype.md) for details on the device-types library schema.

## Strict Mode

By default, `cani` runs in **strict mode** (`--strict=true`). In strict mode, every device must have a resolved device-type slug before export. Unclassified devices will cause the export to fail.

This ensures inventory data is always complete and portable across providers.

### Bypassing Strict Mode

To allow exports with unclassified devices (e.g. during exploratory imports), disable strict mode:

```shell
# Disable strict mode for a single command
cani alpha export nautobot --strict=false

# Or set it globally
cani --strict=false alpha export csm ...
```

> Disabling strict mode is useful when iterating on imports from new sources where not all device types are mapped yet.
