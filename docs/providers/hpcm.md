# HPCM Provider

The HPCM provider imports hardware inventory from HPE Performance Cluster Manager node JSON exports and `cm.config` files. This enables migration of existing HPCM-managed clusters into `cani`'s unified inventory.

## Import

### From A Node JSON File

```shell
# Import from an HPCM node JSON export
cani alpha import hpcm --node-json-file ./hpcm-nodes.json
```

### From A cm.config File

```shell
# Import from an HPCM cm.config file
cani alpha import hpcm --cm-config ./cm.config
```

### From Both Sources

When both flags are provided, nodes are merged and deduplicated by name:

```shell
cani alpha import hpcm --node-json-file ./hpcm-nodes.json --cm-config ./cm.config
```

### From Stdin

If `--node-json-file` is omitted, node JSON can be piped from stdin:

```shell
cat hpcm-nodes.json | cani alpha import hpcm
```

## Export

> Export is not yet implemented for the HPCM provider.

## Classification

Devices imported from HPCM are classified based on their node type (compute, service, switch, storage, etc.). Some devices may need manual classification:

```shell
# Import from HPCM
cani alpha import hpcm --node-json-file ./hpcm-nodes.json

# Auto-classify the imported devices
cani alpha classify --auto
```

See [Classifying Devices](../getting_started/classify.md) for more detail.

## Configuration

Provider-specific options in `~/.cani/cani.yml`:

```yaml
providers:
  hpcm:
    import:
      node_json_file: ""      # Path to HPCM node JSON file
      cm_config_file: ""      # Path to HPCM cm.config file
```
