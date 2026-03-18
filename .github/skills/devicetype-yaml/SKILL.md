---
name: devicetype-yaml
description: "Create or edit devicetype YAML files for the cani hardware library. Use when: adding new hardware (servers, switches, GPUs, cables, racks, PSUs, transceivers), editing device-types, module-types, rack-types, or cable-types YAML, reviewing devicetype YAML for correctness."
argument-hint: "Hardware category and model (e.g., 'HPE ProLiant DL380 server' or 'NVIDIA H100 GPU module')"
---

# Devicetype YAML Authoring

Create and edit hardware definition YAML files under `pkg/devicetypes/`.

## When to Use

- Adding a new server, switch, chassis, node, or other device
- Adding a new module (GPU, NIC, transceiver, PSU, memory)
- Adding a new rack or cabinet definition
- Adding a new cable type
- Reviewing or fixing existing devicetype YAML

## Directory Layout

| Category | Directory | Go Type |
|----------|-----------|---------|
| Devices | `pkg/devicetypes/device-types/<Manufacturer>/` | `CaniDeviceType` |
| Modules | `pkg/devicetypes/module-types/<Manufacturer>/` | `CaniModuleType` |
| Racks | `pkg/devicetypes/rack-types/<Manufacturer>/` | `CaniRackType` |
| Cables | `pkg/devicetypes/cable-types/` | `CaniCableType` |

## Procedure

### 1. Determine the Category

Ask the user or infer from context which category the hardware belongs to:
- **device-types** — servers, blades, switches, chassis, controllers, nodes
- **module-types** — GPUs, NICs, transceivers, PSUs, memory DIMMs
- **rack-types** — cabinets, racks (contain device-bays for child devices)
- **cable-types** — DACs, AOCs, fiber, copper cables

### 2. Gather Required Information

Every type needs at minimum:
- **manufacturer** — company name (matches directory name)
- **model** — full product name
- **slug** — lowercase-kebab-case unique identifier (convention: `manufacturer-model-variant`)
- **hardware-type** — category tag (see [field reference](./references/fields.md))

### 3. Choose the Filename

Convention: `<Model-Name-With-Dashes>.yaml`
- Use the model name converted to title-case dashes
- Examples: `ProLiant-DL380-Gen11-8SFF.yaml`, `NVIDIA-H100-SXM-GPU.yaml`

### 4. Build the YAML

Start with the identity block, then add category-specific sections. Use the templates and field reference below.

### 5. Validate

- **slug must not be empty** — loader silently skips entries with no slug
- **slug must be unique** — first slug wins; duplicates are ignored
- **Racks require `u_height >= 1`**
- Run `go build ./pkg/devicetypes/...` to catch struct/tag mismatches
- Check existing files in the same directory for conventions

## License Header

Every YAML file must begin with this MIT license comment block:

```yaml
#
# MIT License
#
# (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
```

Place a `---` separator after the license block, before the YAML content.

## Templates

### Device Type

```yaml
manufacturer: <Manufacturer>
model: <Full Model Name>
slug: <manufacturer-model-variant>
part_number: <part-number>
description: "<one-line description>"
hardware-type: <blade|hsn-switch|mgmt-switch|pdu|cec|bmc|...>
u_height: <int>
is_full_depth: <true|false>
weight: <float>
weight_unit: <kg|lb>

console-ports:
  - name: <port-name>
    type: <de-9|rj-45|usb-a|other>

power-ports:
  - name: <port-name>
    type: <iec-60320-c14|...>
    maximum_draw: <watts>
    allocated_draw: <watts>

module-bays:
  - name: <bay-name>
    position: <position-label>

device-bays:
  - name: <bay-name>
    allowed:
      slug: [<allowed-child-slug>, ...]
    default:
      slug: <default-child-slug>
    ordinal: <int>

interfaces:
  - name: <interface-name>
    type: <interface-type>
    label: <optional-label>
    mgmt_only: <true>  # only for management ports

provider_defaults:
  csm:
    Class: <River|Hill|Mountain>
    Ordinal: <int>
```

### Module Type

```yaml
manufacturer: <Manufacturer>
model: <Full Model Name>
slug: <manufacturer-model-variant>
part_number: <part-number>
description: "<one-line or multi-line description>"
hardware-type: <gpu|nic|transceiver|psu|memory|...>
weight: <float>
weight_unit: <kg|lb>

interfaces:
  - name: <interface-name>
    type: <interface-type>
```

### Rack Type

```yaml
manufacturer: <Manufacturer>
model: <Full Model Name>
slug: <manufacturer-model-variant>
hardware-type: <Cabinet|rack>

device-bays:
  - name: <bay-name>
    allowed:
      slug: [<allowed-child-slug>, ...]
    default:
      slug: <default-child-slug>
    ordinal: <int>

provider_defaults:
  csm:
    Class: <River|Hill|Mountain>
    Ordinal: <int>
    StartingHmnVlan: <int>
    EndingHmnVlan: <int>
```

### Cable Type

```yaml
manufacturer: <Manufacturer>
model: <Full Model Name>
slug: <manufacturer-model-variant>
part_number: <part-number>
description: "<description>"
hardware-type: cable
cable_category: <dac|aoc|fiber|copper>
connector_type: <qsfp28|qsfp56|qsfpdd|osfp|sfp28|...>
length: <number>
length_unit: <m|ft>
```

## Common Interface Types

| Type | Use Case |
|------|----------|
| `1000base-t` | 1G Ethernet (RJ45) |
| `10gbase-t` | 10G Ethernet (RJ45) |
| `10gbase-x-sfpp` | 10G SFP+ |
| `25gbase-x-sfp28` | 25G SFP28 |
| `40gbase-x-qsfpp` | 40G QSFP+ |
| `100gbase-x-qsfp28` | 100G QSFP28 |
| `200gbase-x-qsfp56` | 200G QSFP56 |
| `400gbase-x-qsfpdd` | 400G QSFP-DD |
| `400gbase-x-osfp` | 400G OSFP |
| `virtual` | Virtual/logical interface |
| `lag` | Link aggregation group |

## Common Hardware Types

| Value | Category |
|-------|----------|
| `blade` | Server / compute node |
| `hsn-switch` | High-speed network switch |
| `mgmt-switch` | Management network switch |
| `pdu` | Power distribution unit |
| `cec` | Cabinet environmental controller |
| `bmc` | Baseboard management controller |
| `gpu` | GPU module |
| `nic` | Network interface card |
| `transceiver` | Optical/copper transceiver |
| `psu` | Power supply unit |
| `cable` | Cable |
| `rack` / `Cabinet` | Rack or cabinet |

## Slug Convention

Format: `<manufacturer>-<model>-<variant>` in lowercase kebab-case.

Examples:
- `hpe-proliant-dl380-gen11-8sff`
- `nvidia-h100-sxm-gpu`
- `hpe-100gb-qsfp28-qsfp28-3m-dac`
- `hpe-ex2500-1-liquid-cooled-chassis`

## Quality Checks

1. Slug is non-empty, unique, and follows kebab-case convention
2. Manufacturer matches the directory name
3. All interface types use valid constants from [Common Interface Types](#common-interface-types)
4. `mgmt_only: true` is set on management-only ports (iLO, BMC, mgmt0)
5. Description is present and accurate
6. File includes the MIT license header comment block
7. YAML uses consistent indentation (2 spaces)
