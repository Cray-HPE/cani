# Ochami Provider

The Ochami provider imports hardware inventory from OpenCHAMI hardware database exports.

## Import

```shell
# Import from an Ochami JSON export file
cani alpha import ochami --source ./ochami-export.json
```

## Export

> Export is not yet implemented for the Ochami provider.

## Configuration

Provider-specific options in `~/.cani/cani.yml`:

```yaml
providers:
  ochami:
    import:
      # Import-specific options
```
