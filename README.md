<p align="center">
  <img src="https://user-images.githubusercontent.com/3843505/235496554-806630e3-a818-4e04-8d46-6a024994d08f.png"" width="150" height="150" alt="cani">
  <br>
  <strong>cani: Cani's Automated Nomicon Inventory</strong>
</p>

> Can I manage an inventory of an entire datacenter? From subfloor to top-of-rack, **yes** you can.

# `cani` Converges Disparate Hardware Inventory Systems

You can inventory hardware with this utility.  The tool itself generates a portable inventory format that can serve as the source of truth itself.  It can also be used to transition from one inventory system to another, using `cani`'s portable format as an intermediate inventory.

## Portable Inventory Format

`cani` uses a simple key/value approach where each piece of hardware as a unique identifier.  Any metadata or relationships to other hardware components is self-contained to that entry in the datastore.

```json
"f5d90c28-f480-4f76-a329-d645ee7af350": {
  "ID": "f5d90c28-f480-4f76-a329-d645ee7af350",
  "Type": "Chassis",
  "Vendor": "HPE",
  "Model": "CrayEX liquid-cooled cabinet chassis",
  "Status": "staged",
  "Parent": "4c7e02e7-5068-4dc2-b727-089a6b11eb66",
  "Children": ["49aa81e0-413e-4142-b56f-8c78cf4daa0c"],
  "LocationPath": [
    {"HardwareType": "Cabinet","Ordinal": 0},
    {"HardwareType": "Chassis","Ordinal": 1}
  ],
  "LocationOrdinal": 1
}
```

## External Inventory Providers

`cani` has built-in support for CSM and can communicate with SLS and HSM.  Support for other inventory providers is still under development.
