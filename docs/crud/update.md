# Update

The `update` command modifies existing hardware in the local inventory. Updates are provider-agnostic — a device imported from Redfish is updated the same way as one imported from CSM.

## Supported Types

```shell
cani alpha update device [flags]
cani alpha update rack [flags]
cani alpha update location [flags]
```

## Examples

```shell
# Update a single device interactively
cani alpha update device

# Bulk-update devices matching a filter
cani alpha update set --status staged --where "name=nid*"
```

## Classify After Update

If a device's type or role changes, run `classify` to ensure it matches a known hardware type:

```shell
cani alpha classify --auto
```

## Provider Independence

Updates apply to the local inventory only. The provider handles reconciliation during export:

1. **Import** from one or more providers to populate the inventory.
2. **Update** devices locally using a consistent interface.
3. **Export** to the original provider or a different one — the export step translates the common inventory format into whatever the target system expects.

This means you can import from HPCM, update device roles locally, and export the result to Nautobot without any provider-specific update logic.
