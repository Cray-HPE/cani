# Remove

The `remove` command deletes hardware from the local inventory. Like all CRUD operations, it works the same way regardless of which provider the data came from.

## Supported Types

```shell
cani alpha remove device [flags]
cani alpha remove rack [flags]
cani alpha remove location [flags]
cani alpha remove module [flags]
```

## Examples

```shell
# Remove a device interactively (select from a list)
cani alpha remove device

# Remove a device by name
cani alpha remove device --name nid000042
```

## Cascading Deletes

Removing a parent item (e.g. a rack) also removes its children (the devices in that rack). A confirmation prompt is shown before cascading.

## Provider Independence

Removed hardware is deleted from the local inventory immediately. On the next export, the provider determines how to reconcile the difference:

- Providers with full sync (e.g. CSM) will remove the corresponding records from the external system.
- Providers with merge-mode export will skip removed items.

This makes it straightforward to consolidate hardware across systems:

```shell
# Import from both CSM and Redfish
cani alpha import csm ...
cani alpha import redfish --root ./redfish-roots.json

# Remove decommissioned hardware
cani alpha remove device --name old-node-001

# Export the cleaned inventory to CSM
cani alpha export csm ...
```
