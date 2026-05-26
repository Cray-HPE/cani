# Show

The `show` command reads hardware from the local inventory. It works identically regardless of which provider(s) the data was originally imported from.

## Supported Types

```shell
cani alpha show location
cani alpha show rack
cani alpha show device
cani alpha show module
cani alpha show cable
```

## Filtering

Results can be narrowed with flags:

```shell
# Show devices with a specific status
cani alpha show device --status staged

# Show devices matching a name pattern
cani alpha show device --name "nid*"
```

## Provider Independence

Because the inventory is normalized into a common format during import, `show` displays a consistent view whether the data came from CSM, Nautobot, Redfish, HPCM, or any combination:

```shell
# Import from multiple providers
cani alpha import csm ...
cani alpha import redfish --root ./redfish-roots.json

# Show a unified view of all devices
cani alpha show device
```

All devices share the same fields (`name`, `status`, `type`, `manufacturer`, `model`, etc.) regardless of their origin. Provider-specific details are preserved in `providerMetadata` but do not affect the common display.
