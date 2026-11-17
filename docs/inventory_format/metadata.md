# Metadata

Every item in the inventory carries the same metadata fields: locations, racks, devices, modules, cables, FRUs, interfaces, VLANs, prefixes, IP addresses and VRFs. Each kind of value has one home.

| A value that is | Lives in | Written by | Sent on export |
| --- | --- | --- | --- |
| An attribute the inventory models, such as a status, a role, a tag, a device's location or a rack's height | The field of that name | The field's own flag (`--status`, `--role`, `--tag`, `--location`) and any provider's import | Yes. Each export maps the field. |
| An attribute you define | [`customFields`](#custom-fields), typed by a definition in the [catalog](#inventory-metadata-catalog) | `--custom-field`, and an import from a system that has custom fields | Yes, to a system that has custom fields. |
| What one provider records about the item for its own use | [`providerMetadata.<provider>`](#provider-metadata) | That provider only | Only by that provider's own export. Nautobot receives a key only when you list it. |
| The item's ID in another system | [`externalIDs.<provider>`](#external-ids) | That provider | No. It finds the remote record. |

A role, a location or a custom field kept in provider metadata is that provider's own record. No command treats it as the item's role, location or custom field, and no other system receives it unless you list it for export. To give another system a value, set the field or a custom field.

## ObjectMeta

In Go, every inventory type embeds the same `ObjectMeta` struct:

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

Because `ObjectMeta` is embedded, these fields appear directly on each type. In JSON they are serialized as flat fields:

```json
{
  "id": "f7448392-...",
  "name": "nid000001",
  "status": "Active",
  "role": "ComputeNode",
  "tags": ["gpu-node"],
  "tenant": "HPC-East",
  "customFields": { "cost_center": 1001 },
  "externalIDs": { "nautobot": "9a8b7c6d-5e4f-3a2b-1c0d-000000000001" },
  "providerMetadata": {
    "csm": { "xname": "x3000c0s3b0n0", "nid": 1 }
  }
}
```

## Inventory Metadata Catalog

The inventory carries a top-level `metadata` field: the catalog of roles, statuses and tags available for assignment, and of the custom fields that are defined.

```json
{
  "schemaVersion": "v1alpha10",
  "metadata": {
    "roles": [
      { "name": "ComputeNode", "contentTypes": ["dcim.device"], "weight": 1000 }
    ],
    "statuses": [
      { "name": "Active", "color": "green" }
    ],
    "tags": [
      { "name": "gpu-node", "description": "Nodes with GPUs" }
    ],
    "customFields": [
      { "key": "cost_center", "label": "Cost center", "type": "integer", "contentTypes": ["dcim.device", "dcim.rack"] }
    ]
  }
}
```

Add entries with the CLI:

```shell
cani alpha add metadata role ComputeNode --content-types dcim.device
cani alpha add metadata status Active --color green
cani alpha add metadata tag gpu-node --description "Nodes with GPUs"
cani alpha add metadata custom-field cost_center --label "Cost center" --type integer --content-types dcim.device,dcim.rack
```

## Custom Fields

A custom field is an attribute you define, such as a cost center, an owner or an install date. `customFields` holds the values. The catalog holds one definition per key, which names the kinds of item the field applies to and the type of its value: `text`, `integer`, `boolean`, `date`, `url`, `json`, `select`, `multi-select` or `markdown`.

Set a custom field when you add or update an item, and clear it on update:

```shell
cani alpha add device hpe-dl380-gen-11 --name nid000001 --rack r1 --position 1 --custom-field cost_center=1001
cani alpha update device nid000001 --custom-field cost_center=1002
cani alpha update device nid000001 --clear-custom-field cost_center
```

- The value is stored as the type the definition declares. `cost_center=1001` is stored as the number `1001`.
- Clearing a defined field stores it as `null` instead of removing the key, because a system that already holds the item keeps the value of a key an export leaves out. Clearing a key that has no definition removes it.
- A key that is not defined for that kind of item is refused, and so is a value of the wrong type:

  ```text
  Error: --custom-field: custom field "owner" is not defined in the inventory; define it with 'add metadata custom-field' or import its definition
  Error: --custom-field: custom field "cost_center": "east" is not an integer
  ```

- An import from Nautobot brings the definitions with the values, and an export to Nautobot creates a definition the destination lacks. See [Custom fields](../providers/nautobot.md#custom-fields) in the Nautobot provider.
- A provider does not keep its own records here. They go in its provider metadata.
- An inventory upgraded from an earlier release can hold a custom field that has no definition. It is kept and exported as before. See [Provider metadata on upgrade](overview.md#provider-metadata-on-upgrade).

### Stored Values Are Checked

The flags check a value as it is typed. A value that reached `customFields` another way, from an import, from an upgrade or from an edit of the datastore file, is checked each time the inventory is loaded or changed, against the same definitions. A command that changes the inventory warns about each problem once per field and kind of item, with how many items have it and one example:

```text
Warning: custom field "rack_pos" on dcim.device has no definition; define it with "cani alpha add metadata custom-field" or remove the value; 12 dcim.device object(s) hold it, e.g. device "nid000001" (001a1a2f-dc9e-424d-b04f-7f56a02655cc)
Warning: custom field "cost_center" on dcim.device holds a value that is not an integer; 1 dcim.device object(s) hold it, e.g. device "nid000002" (...)
```

The checks are: the key has a definition; the definition applies to that kind of item; the value is of the kind the type stores (a whole number for `integer`, `true` or `false` for `boolean`, `YYYY-MM-DD` for `date`, one of the choices for `select`, a list of them for `multi-select`, text for the rest; `json` takes anything); and a `required` field does not hold `null`. The value is kept and the command is not refused. To clear a warning, define the field or extend its definition to that kind of item, set the value again with `--custom-field`, or remove it with `--clear-custom-field`.

## Provider Metadata

`providerMetadata` holds one map per provider, under the provider's name. That map is the provider's bucket. It is what the provider records about the item for its own use and the inventory has no field for: the CSM xname, the HPCM node UUID, the Redfish service UUID.

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

A bucket belongs to its provider:

- **Only the provider writes it.** Its import fills the bucket. So do the flags and hooks the provider adds to commands, such as CSM's `--nid` and `--alias` on `update device`. No other flag writes provider metadata.
- **The provider reads it for its own export.** `cani alpha export csm` reads a device's xname, class, NID and aliases from `providerMetadata.csm`.
- **It is not sent to other systems.** The Nautobot export sends none of it, unless you list a key. See [Sending provider metadata as custom fields](../providers/nautobot.md#sending-provider-metadata-as-custom-fields).
- **It sets nothing else.** A `role` in a bucket is that provider's record of the role, and only that provider reads it. `show` displays the item's `role` field, and the Nautobot export sends that field. The same holds for a location and for a rack's height.
- **An import does not remove it.** Importing an item again, from the same system or another, updates the importing provider's bucket and leaves the others as they are.

One kind of key is read across buckets: an **identity key**, one a provider declares as naming a single physical device. CSM and OChaMi declare `xname`; Redfish declares `bmc_fqdn` and `bmc_hostname`. Declaring a key says it means the same thing in every bucket that holds it, so:

- An import merges a device into an existing device of the same name only when the identity keys the importing provider declares agree: for the first declared key both devices hold, in any bucket, the values are equal. A provider that declares none, such as HPCM or Nautobot, merges by name alone.
- `update orphans` scores a candidate parent higher when the orphan and the candidate record a declared key with a common prefix in the same provider's bucket, as a node's xname does with its cabinet's.
- The OChaMi export reads a node's `xname` from whichever bucket holds it, so a node CSM imported can be exported to OpenCHAMI. Its `ip`, `mac` and `boot_mac` come from the OChaMi bucket alone; no other provider records them.

### A Value Outside Every Bucket

Every entry of `providerMetadata` is a provider's bucket. A value stored beside the buckets belongs to no provider, and nothing reads it:

```json
"providerMetadata": {
  "cost_center": "1001",
  "csm": { "xname": "x3000c0s3b0n0" }
}
```

A command that changes the inventory warns about it, by item and key. It lists the first ten warnings and counts the rest; `--debug` lists them all.

```text
Warning: device "nid000001" (001a1a2f-dc9e-424d-b04f-7f56a02655cc): provider metadata "cost_center" is outside every provider's bucket, so nothing reads it; move it into a bucket or into customFields
```

The value is kept and the command is not refused. `cani` writes no such value itself, and loading an inventory from an earlier release moves the ones those releases wrote (see [Provider metadata on upgrade](overview.md#provider-metadata-on-upgrade)). One can come from a provider that writes it, or from an inventory file edited or written by hand.

To clear the warning, edit the datastore file, which `datastore` in the configuration names (`canidb.json` by default):

- If the value is a provider's record, move the key into that provider's bucket.
- If it is an attribute of your own, delete the key, define a custom field and set it with `--custom-field`.
- If nothing needs it, delete the key.

An entry that holds a map is a bucket. One named for a provider that is not built into this `cani` is reported too, once per name, since nothing reads it either: a bucket left by a build that had the provider, or a misspelt name.

```text
Warning: provider metadata bucket "acme" is held by 3 object(s), e.g. device "nid000001" (001a1a2f-dc9e-424d-b04f-7f56a02655cc), but no provider "acme" is built in, so nothing reads it
```

Move its keys into the right provider's bucket, or delete it.

### Provider-Specific Keys

#### CSM

| Key | Type | Description |
|-----|------|-------------|
| `xname` | string | Location-based component name (e.g. `x3000c0s3b0n0`) |
| `class` | string | Hardware class (`Mountain`, `River`, `Hill`) |
| `role` | string | Node role (`Compute`, `Management`, `Application`). Import also sets the device's `role` field from it. The CSM export sends this key, and the field only when the key is empty. |
| `subRole` | string | Node sub-role (`Worker`, `Master`, `Storage`) |
| `nid` | int | Node ID number |
| `aliases` | []string | Hostname aliases |
| `state` | string | Hardware state from HSM |
| `hmnVlan` | int | Hardware Management Network VLAN ID |
| `import_source` | string | Where the device was read from (`sls`) |

#### Redfish

| Key | Type | Description |
|-----|------|-------------|
| `redfish_uuid` | string | UUID reported by the Redfish service |
| `redfish_version` | string | Redfish version of the service |
| `vendor` | string | Vendor the service reports |
| `odata_type` | string | OData type of the service root |
| `bmc_type` | string | Manager type |
| `bmc_firmware` | string | Manager firmware version |
| `bmc_fqdn` | string | BMC fully-qualified domain name |
| `bmc_hostname` | string | BMC host name |
| `product_tag` | string | Product tag |
| `system_family` | string | System family |
| `import_source` | string | The ServiceRoot file the device was read from |

#### OChaMi

| Key | Type | Description |
|-----|------|-------------|
| `source` | string | The JSON file the record was read from |
| `xname` | string | Location-based component name |
| `nid` | int | Node ID number |
| `hostname` | string | Host name |
| `host_aliases` | []string | Host aliases |
| `ip` | string | IP address |
| `mac` | string | MAC address |
| `boot_mac` | string | MAC address the node boots from |
| `serial_number` | string | Serial number in the record |
| `parent_serial_number` | string | Serial number of the record's parent |
| `redfish_uri` | string | Redfish URI of the component |

#### Example

| Key | Type | Description |
|-----|------|-------------|
| `Source` | string | Import source (e.g. the CSV file path) |
| `PartNumber` | string | Part number from the CSV record |
| `ConfigGroup` | string | Configuration group from the CSV record |

#### HPCM

The HPCM provider keeps what HPCM records about an object, and the inventory has no field for, under the `"hpcm"` key. None of it is stored as a custom field.

| Key | On | Type | Description |
|-----|----|------|-------------|
| `hpcm_uuid` | devices, modules, locations | string | UUID of the HPCM node |
| `aliases` | devices, modules | map | The node's HPCM aliases |
| `location` | devices, modules | map | The node's HPCM location: `rack`, `chassis`, `tray`, `node`, `controller` |
| `geoloc` | modules | string | The geolocation xname the module's chassis was found by |
| `rack_number` | racks | int | The HPCM rack number |
| `import_source` | devices | string | The node file the device was read from |
| any other key | inventory items | string | An inventory field HPCM reported that the inventory has no field for, such as a disk's `size` |

### Custom Providers

A provider writes whatever it needs in its own bucket, without a schema change. The only requirement is that the top-level key is the provider's name, which registering the provider makes a known bucket. In Go, `SetProviderMeta(provider, key, value)` writes one key of a bucket and `GetProviderSubMap(provider)` reads the bucket. A provider whose bucket holds a key that names one physical device declares it on `TransformResult.IdentityKeys` and returns it from `IdentityKeys()`; see [Provider Metadata](#provider-metadata) for what that means.

```json
"providerMetadata": {
  "my-provider": {
    "custom_key": "custom_value",
    "nested": { "anything": true }
  }
}
```

## External IDs

The `externalIDs` field maps a provider's name to the UUID that provider's system uses for the item. An export uses it to update the right remote record.

```json
"externalIDs": {
  "nautobot": "9a8b7c6d-5e4f-3a2b-1c0d-000000000001"
}
```

## Tags

The `tags` field is a string slice for arbitrary labels (e.g. `["production", "gpu-node"]`).
