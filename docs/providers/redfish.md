# Redfish Provider

The Redfish provider imports hardware inventory from Redfish ServiceRoot JSON data. This is commonly used for BMC/iLO discovery from any Redfish-compliant endpoint.

## Import

### From A File

```shell
# Import from a Redfish ServiceRoot JSON file
cani alpha import redfish --root ./redfish-roots.json
```

### From Stdin

```shell
# Pipe Redfish data from another tool
curl -s https://bmc01.example.com/redfish/v1 | cani alpha import redfish --root -
```

## Export

> Export is not yet implemented for the Redfish provider.

## Classification

Devices imported from Redfish may not have a device-type slug assigned. Use the `classify` command to resolve them:

```shell
# Import from Redfish
cani alpha import redfish --root ./redfish-roots.json

# Auto-classify the imported devices
cani alpha classify --auto
```

See [Classifying Devices](../getting_started/classify.md) for more detail.

## Configuration

Provider-specific options in `~/.cani/cani.yml`:

```yaml
providers:
  redfish:
    import:
      root: ""                  # Path to Redfish ServiceRoot JSON file
```
