# Example Provider

The Example provider is a reference implementation for bootstrapping new providers. It imports from CSV or YAML files and exports a visual hierarchy.

## Import

```shell
# Import from a CSV file
cani alpha import example --source ./inventory.csv

# Import from a YAML file
cani alpha import example --source ./inventory.yaml
```

## Export

```shell
# Export a visual hierarchy of the inventory
cani alpha export example
```

## Usage

The Example provider is primarily intended as:

- A reference for developing new providers
- A way to bootstrap an inventory from flat files (CSV/YAML)
- A debugging tool for viewing the inventory hierarchy

## Configuration

Provider-specific options in `~/.cani/cani.yml`:

```yaml
providers:
  example:
    import:
      # Import-specific options
```
