# Nautobot Provider

The Nautobot provider imports from and exports to Nautobot DCIM instances (Netbox-compatible). It uses the Nautobot REST API to sync devices, racks, and device types.

## Import

```shell
# Import from a Nautobot instance
cani alpha import nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN
```

## Export

```shell
# Export to a Nautobot instance
cani alpha export nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN

# Preview changes without making API calls
cani alpha export nautobot --dry-run

# Merge with existing devices instead of skipping conflicts
cani alpha export nautobot --merge
```

## Configuration

Provider-specific options in `~/.cani/cani.yml`:

```yaml
providers:
  nautobot:
    url: ""                     # Base URL of the Nautobot instance
    token: ""                   # API token for authentication
    default_location: ""        # Default location for devices
    default_role: ""            # Default role for devices
    default_status: ""          # Default status for devices
    import:
      # Import-specific options
    export:
      # Export-specific options
```
