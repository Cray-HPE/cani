# Use Custom Hardware Types With CANI

CANI offers built-in hardware types for each provider, but it is also possible to define custom hardware.

## Define A Custom Hardware Type

Ensure a `hardware-types` directory exists next to the `cani.yml` config file (CANI makes the directory by default).

Create a YAML file that conforms to the [harware-types schema](./devicetype.md).  A simple example is shown below that adds a custom cabinet type that supports a custom chassis type:

```yaml
---
manufacturer: HPE
model: EX2000
hardware-type: Cabinet
slug: my-custom-cabinet

device-bays:
  - name: Chassis 0
    allowed:
      slug: [my-custom-chassis]
    default:
      slug: my-custom-chassis
    ordinal: 0

provider_defaults:
  csm:
    Class: River
    Ordinal: 4321
    StartingHmnVlan: 1111
    EndingHmnVlan: 1769

---
manufacturer: HPE
model: Standard/EIA Chassis
hardware-type: Chassis
slug: my-custom-chassis

provider_defaults:
  csm:
    Class: River
    starting_cabinet: 4321
    StartingHmnVlan: 1111
    EndingHmnVlan: 1769
```
