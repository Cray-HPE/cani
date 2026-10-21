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

### Device-bay moves

With `--merge`, export reconciles bay assignments for inventory devices that
resolve successfully in Nautobot. It clears obsolete assignments before device
updates and installs the desired assignments afterward. This supports moving a
child between bays or parents, detaching it into a rack, swapping children, and
replacing an occupant that is moving out.

Destination conflicts are checked before any bay is cleared. Occupants outside
the resolved inventory, including devices removed entirely from local inventory,
are never automatically detached. Without `--merge`, existing occupied bays
remain conflicts.

Bay lookup errors are returned rather than treated as missing bays. A bay is
created only after a successful, well-formed lookup confirms no matches; failed
HTTP requests, malformed responses, and inconsistent empty pages never trigger
creation.

Creation responses must include a nonzero bay UUID, a nonzero parent UUID, and a
bay name. Invalid responses are reported without sending an installation PATCH.
Failed creations and installations do not increment successful-install counts.

`--dry-run --merge` reports planned removals and installations without sending
write requests. If discovery or removal fails, export stops before device
updates or installations; successful removals before a later failure are not
rolled back. Re-export after correcting the failure to finish reconciliation.

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
