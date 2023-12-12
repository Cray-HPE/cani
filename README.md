<p align="center">
  <img src="docs/custom_theme/img/hpe_pri_wht_rev_rgb.png"" width="150" height="63" alt="HPE Logo">
  <br>
  <strong>Continuous And Never-ending Inventory</strong>
</p>

# Use Cases

- Migrating from one inventory system to another
- Validating exisiting inventory data
- Exporting to another inventory format
- Serving as a front-end for the UX, while the backend is replaced with something new
- Use as a new self-contained inventory system

# Add/Remove/Update Hardware Example

```shell
cani session init csm # start a session with a specific provider by importing data and converting it to CANI format
cani add cabinet hpe-ex4000 --auto --accept # add hardware (child hardware is added automatically: chassis, controllers, etc.)
cani add blade hpe-crayex-ex4252-compute-blade --auto --accept # add a blade to the cabinet (and any nodes that hardware contains)
cani update node --uuid abcdef12-3456-2789-abcd-ef1234567890 --role Management --subrole Worker --nid 4 # update hardware metadata
cani export --format hpcm # export to another inventory format (can also commit and init a new session with a different provider)
cani session apply --commit # apply changes, retaining the CANI format, and optionally posting data to the provider
```

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

## Currently Supported Providers

- CSM
- HPCM (WIP)
