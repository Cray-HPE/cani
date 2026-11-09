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

### Custom fields

Every object family that carries custom fields in the inventory (locations,
racks, devices, modules, cables, inventory items, VLANs, prefixes, IP
addresses, VRFs) sends them on create and on `--merge`, through one projection:
explicit custom fields merged with flattened provider metadata. When both set
the same key the provider value is sent and the export logs the collision by
object and key. `false`, `0` and empty strings are sent as written, because
Nautobot fills an omitted key from the definition's default.

Import fetches the custom-field definitions and their choices into the
inventory metadata catalog, and export creates any definition the destination
lacks before it writes the first object (Phase 0a). Nautobot answers 2xx for
a value whose key has no definition for that object type (it logs a warning
on the server and keeps the value in the object's data, but the API never
exposes it), so after each write the export compares the keys it sent with
the keys the response carries and reports every key that did not arrive; the
summary counts them with the other objects that were not exported as
authored. The catalog does not model a
definition's validation limits and pattern, filter logic, grouping, advanced-UI
flag or scope filter, nor its exact choice weights (choices keep their order):
import names every definition that sets one, never the value, and a fresh
destination gets the definition without it.

Under `--merge`, a definition's label, description, required flag, weight and
newly added content types are reconciled. Its `type` and `key` are immutable in
Nautobot and are only reported when they differ, and a content type is never
removed, since Nautobot would delete the field's values on every object of
that type.

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
