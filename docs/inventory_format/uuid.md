# UUID

Every item in the inventory — whether a location, rack, device, module, cable, FRU, or interface — is assigned a unique UUID (v4) when it is created. This UUID is the primary identifier the system uses for all read and write operations.

## How UUIDs Are Used

- **Map keys** — Each hardware map in the `Inventory` is keyed by UUID, giving O(1) lookups.
- **Parent/child references** — Relationships between items (e.g. a device's `parent`, a rack's `devices` list) are stored as UUIDs, not names.
- **Provider deduplication** — When importing data, the system checks whether an incoming item already exists by looking up provider-specific keys (via the provider-key index) before assigning a new UUID. This prevents duplicates across repeated imports.
- **Export targeting** — During export, `externalIDs` maps a provider name to the remote UUID so the system knows which remote record to update.

## Generation

UUIDs are generated with `uuid.New()` (Google's UUID v4 implementation). Once assigned, a UUID never changes for the lifetime of that inventory item.

## Example

```json
{
  "f7448392-1e1c-45d0-9c59-be7dfc44c15c": {
    "id": "f7448392-1e1c-45d0-9c59-be7dfc44c15c",
    "name": "nid000001",
    "type": "Device",
    "status": "active"
  }
}
```

The key `f7448392-1e1c-45d0-9c59-be7dfc44c15c` and the `id` field always match. The key is used for map lookups; the `id` field makes the identifier accessible when the item is serialized on its own.
