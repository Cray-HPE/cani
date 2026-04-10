# Metadata

## ObjectMeta

Every inventory type (device, rack, location, module, cable, FRU, interface) embeds the same `ObjectMeta` struct. This provides a uniform set of metadata fields across all hardware:

```go
type ObjectMeta struct {
    Status           string
    Role             string
    Tags             []string
    Tenant           string
    CustomFields     map[string]any
    ExternalIDs      map[string]uuid.UUID
    ProviderMetadata map[string]any
}
```

Because `ObjectMeta` is embedded in Go, these fields appear directly on each type. In JSON they are serialized as flat fields:

```json
{
  "id": "f7448392-...",
  "name": "nid000001",
  "status": "active",
  "role": "Compute",
  "tags": ["production", "gpu-node"],
  "tenant": "HPC-East",
  "providerMetadata": {
    "csm": { "xname": "x3000c0s3b0n0", "nid": 1 }
  }
}
```

## Inventory Metadata Catalog

The inventory carries a top-level `metadata` field that stores the global catalog of roles, statuses, and tags available for assignment:

```json
{
  "schemaVersion": "v1alpha2",
  "metadata": {
    "roles": [
      { "name": "ComputeNode", "contentTypes": ["dcim.device"], "weight": 1000 }
    ],
    "statuses": [
      { "name": "Active", "color": "green" }
    ],
    "tags": [
      { "name": "gpu-node", "description": "Nodes with GPUs" }
    ]
  }
}
```

Add entries with the CLI:

```shell
cani alpha add metadata role ComputeNode --content-types dcim.device
cani alpha add metadata status Active --color green
cani alpha add metadata tag gpu-node --description "Nodes with GPUs"
```

## ProviderMetadata

Devices and racks carry a `providerMetadata` field: a `map[string]any` keyed by provider name. Each provider stores its own arbitrary data under its key, so multiple providers can coexist without colliding.

```json
"providerMetadata": {
  "csm": {
    "xname": "x3000c0s3b0n0",
    "class": "Mountain",
    "role": "Compute",
    "nid": 1
  },
  "redfish": {
    "redfish_uuid": "abc-123",
    "bmc_fqdn": "bmc1.example.com"
  }
}
```

The top-level key (`"csm"`, `"redfish"`, etc.) identifies which provider wrote the data. The nested map can contain anything that provider needs.

### Provider-Specific Keys

#### CSM

| Key | Type | Description |
|-----|------|-------------|
| `xname` | string | Location-based component name (e.g. `x3000c0s3b0n0`) |
| `class` | string | Hardware class (`Mountain`, `River`, `Hill`) |
| `role` | string | Node role (`Compute`, `Management`, `Application`) |
| `subRole` | string | Node sub-role (`Worker`, `Master`, `Storage`) |
| `nid` | int | Node ID number |
| `aliases` | []string | Hostname aliases |
| `state` | string | Hardware state from HSM |
| `hmnVlan` | int | Hardware Management Network VLAN ID |

#### Redfish

| Key | Type | Description |
|-----|------|-------------|
| `redfish_uuid` | string | UUID reported by the Redfish service |
| `bmc_fqdn` | string | BMC fully-qualified domain name |

#### Example

| Key | Type | Description |
|-----|------|-------------|
| `Source` | string | Import source (e.g. the CSV file path) |
| `PartNumber` | string | Part number from the CSV record |
| `ConfigGroup` | string | Configuration group from the CSV record |

#### HPCM

The HPCM provider stores a free-form map under the `"hpcm"` key. Common fields include `node_uuid`, `location`, aliases, and other node attributes from the HPCM JSON export.

### Custom Providers

Any provider can write arbitrary keys under its own namespace. The only requirement is that the top-level key matches the provider name. This makes `providerMetadata` fully extensible — a custom provider can store whatever data it needs without a schema change.

```json
"providerMetadata": {
  "my-provider": {
    "custom_key": "custom_value",
    "nested": { "anything": true }
  }
}
```

## ExternalIDs

The `externalIDs` field maps provider names to the remote UUID that provider uses for this item. This is used during export to update the correct remote record.

```json
"externalIDs": {
  "nautobot": "9a8b7c6d-5e4f-3a2b-1c0d-000000000001"
}
```

## CustomFields

The `customFields` field is a general-purpose `map[string]any` for user- or provider-defined data that does not belong under a specific provider namespace.

## Tags

The `tags` field is a string slice for arbitrary labels (e.g. `["production", "gpu-node"]`).
