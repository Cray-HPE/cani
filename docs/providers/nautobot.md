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
addresses, VRFs) sends them on create and on `--merge` through one
projection, and `--merge` compares with the same set. An object's custom
fields are sent as they are: `false`, `0` and empty strings are sent as
written, because Nautobot fills an omitted key from the definition's default.

Provider metadata is not exported. It belongs to the provider that wrote it
(`providerMetadata.csm`, `providerMetadata.redfish`, and so on), and a role,
location or height held there sets nothing in Nautobot: a device's role comes
from its role, its location from where the inventory places it, and the
height of a rack from its height. Give Nautobot a value with the field's own
flag or with `--custom-field`.

#### Sending provider metadata as custom fields

To send a provider's value to Nautobot, list its key in the export options.
An entry is `provider.key`, sent as the custom field `key`, or
`provider.key=custom_field`, sent under another name:

```yaml
providers:
  nautobot:
    export:
      provider_metadata_fields:
        - csm.xname        # providerMetadata.csm.xname as custom field "xname"
        - csm.nid=node_id  # providerMetadata.csm.nid as custom field "node_id"
```

`--provider-metadata-field` (comma-separated or repeatable) replaces the
configured list for one export:

```shell
cani alpha export nautobot --provider-metadata-field csm.xname,csm.nid=node_id
```

- The provider is the text before the first dot, and only that provider's
  bucket is read. An object whose bucket does not hold the key sends nothing
  for it.
- The list defines nothing. The custom field needs a definition for each
  object type that holds the key, in the inventory
  (`cani alpha add metadata custom-field`, which the export creates in
  Nautobot) or already in Nautobot. Without one Nautobot drops the value, and
  the export reports the object and the key each time; with `--merge` it also
  updates the object again on every export.
- A listed key belongs to its provider. Its value is sent in place of a
  custom field of the same name, and while the two differ the export names
  the object and the key and does not send the object's own value. An edit
  made to such a field in Nautobot is overwritten by the next export with
  `--merge`, and one made with `--custom-field` has no effect; change the
  value where the provider reads it.
- An entry that names no provider or no key, or two keys sent as the same
  custom field, stops the export before it connects to Nautobot.

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

### IP address assignments and primary IPs

Import reads the `ipam/ip-address-to-interface` join table (a Nautobot
address does not list its own interfaces) and links every address to the
device and module ports it is assigned to, one address on several ports
included; an assignment to a virtual-machine interface is reported and left
out. A device's primary IPv4 and IPv6 travel with it, and Nautobot accepts a
primary address only when it is assigned to one of that device's interfaces,
so the two arrive together on a fresh destination.

On a re-import, an imported device or module that matches an existing record
keeps that record's ports, and the assignments and cables follow the ports by
name, so an inventory authored with the CLI can take an import. An imported
address replaces the local one, assignments included. An assignment to a port
the existing record has no port of that name for (a bond, a loopback or a
port the template names differently) is dropped from the import and listed
with the other unresolved records; the rest of the import merges, and
Nautobot keeps the assignment, since export never deletes one.

Export creates the assignments an address lacks and leaves the ones Nautobot
already holds alone: for an address that already exists it reads the current
rows first, so a repeated export never posts a duplicate, and the summary
counts assignments created and skipped. The join row has no change-log entry
of its own (Nautobot records its creation as an update of the address and of
the interface), so a repeated export is verified by counting its rows as
well as the ledger. A row's `is_source`, `is_destination`, `is_default`,
`is_preferred`, `is_primary`, `is_secondary` and `is_standby` flags are not
modelled: import keeps the assignment and names every flag the row sets.

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
