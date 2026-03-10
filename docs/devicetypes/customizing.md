# Customizing Device Types

`cani` ships with built-in device types, but custom types can be added for hardware that is not in the library.

## Creating A Device Type

Create a YAML file that follows the [device type schema](./devicetype.md). Place it in a `device-types/` subdirectory (or `rack-types/`, `module-types/`, etc. depending on the category).

### Device Example

```yaml
manufacturer: HPE
model: Custom Compute Blade
hardware-type: Device
slug: my-custom-device
u_height: 1
is_full_depth: true

interfaces:
  - name: mgmt0
    type: 1000base-t
  - name: hsn0
    type: 200gbase-x-qsfp56

module-bays:
  - name: GPU 0
    allowed:
      slug: [my-custom-gpu-module]
    default:
      slug: my-custom-gpu-module
    ordinal: 0
```

### Rack Example

```yaml
manufacturer: HPE
model: Custom 48U Rack
hardware-type: Rack
slug: my-custom-rack
u_height: 48

device-bays:
  - name: Bay 0
    allowed:
      slug: [my-custom-device]
    default:
      slug: my-custom-device
    ordinal: 0
```

## Required Fields

| Field | Description |
|-------|-------------|
| `manufacturer` | Vendor name |
| `model` | Model identifier |
| `slug` | Unique key (must not collide with existing slugs) |
| `hardware-type` | One of `Device`, `Rack`, `Module`, `Cable` |

## Verifying Custom Types

After adding a custom type, use the `-L` flag to confirm it was loaded:

```shell
# Verify the custom device type appears in the list
cani alpha add device -L

# Verify the custom rack type appears
cani alpha add rack -L
```

Custom types appear alongside built-in types and can be used with any `add` command:

```shell
cani alpha add device my-custom-device --auto --accept
cani alpha add rack my-custom-rack --auto --accept
```

See [Extending](extending.md) for how to configure where `cani` loads types from.
