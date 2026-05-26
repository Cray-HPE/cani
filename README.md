<p align="center">
  <img src="docs/custom_theme/img/HPE_logo_full-clr_rev_rgb.png" width="200" height="57" alt="HPE Logo">
  <br>
  <strong>Continuous And Never-ending Inventory</strong>
</p>

# What is CANI?

`cani` is a provider-agnostic hardware inventory tool built around an **ETL (Extract-Transform-Load) pipeline**. It imports inventory from any supported source, converts it into a portable Nautobot/Netbox-inspired format, and exports it to any supported target — enabling migrations, consolidation, and validation across systems.

Every piece of hardware — devices, racks, locations, modules, cables, and FRUs — is tracked with a UUID, self-contained metadata, and parent/child relationships. The same CRUD commands work regardless of provider; only the import and export logic is custom.

> **Note:** Commands currently live under the `cani alpha` subcommand while the CLI stabilizes.

---

# Key Features

- **Multi-provider ETL pipeline** — Import from CSM, HPCM, Nautobot, Redfish, Ochami, or CSV/YAML files. Export to any supported target. Mix and match sources in a single inventory.
- **Portable inventory format** — A Nautobot/Netbox-inspired schema where every item is UUID-keyed with self-contained metadata and relationships. No external lookups required.
- **Uniform CRUD** — The same `add`, `remove`, `update`, and `show` commands work for every provider. Providers only customize their import and export logic.
- **Slug-based hardware** — Add hardware with just a device-type slug (e.g. `cani alpha add hpe-xl225n-gen10-plus`). Metadata, interfaces, module bays, and power ports are auto-populated from the device-type library.
- **Auto-classification** — Devices without a known type can be classified interactively or automatically via fuzzy matching against the device-type library.
- **Visual rack diagrams** — ASCII rack views, compact symbolic layouts, and cable routing visualization directly in the terminal.
- **Provider scaffolding** — `cani init <name>` generates a complete provider skeleton so new integrations start with working boilerplate.
- **Orphan resolution** — Devices imported without a parent rack can be resolved interactively or via a saved plan.
- **Stepped import** — Walk through each ETL phase record-by-record with `--step` for debugging or review.
- **Zero third-party runtime dependencies** — Pure Go, standard library only.

---

# Supported Providers

| Slug | Import Source | Export Target | Description |
|------|-------------|---------------|-------------|
| `csm` | SLS/SMD JSON files or live CSM API | SLS API or CSV | Cray System Management (HPE Cray EX) |
| `hpcm` | Node JSON or `cm.config` files | — | HPE Performance Cluster Manager |
| `nautobot` | Nautobot REST API | Nautobot REST API | Nautobot DCIM (Netbox-compatible) |
| `ochami` | Ochami JSON export | — | OpenCHAMI hardware database |
| `redfish` | Redfish ServiceRoot JSON (file or stdin) | — | BMC/iLO discovery (any Redfish endpoint) |
| `example` | CSV or YAML files | Visual hierarchy | Reference implementation for bootstrapping |

---

# How Inventory Maps Between Formats

`cani` acts as a universal translator. Import from one format, store in CANI's portable schema, export to another.

**Example: Redfish → CANI → Nautobot**

**1. Source — Redfish ServiceRoot (from a BMC)**

```json
{
  "UUID": "946a7d44-9967-4940-9490-f2d581950512",
  "Product": "ProLiant DL325 Gen11",
  "Vendor": "HPE",
  "Oem": {
    "Hpe": {
      "Manager": [{
        "ManagerType": "iLO 6",
        "ManagerFirmwareVersion": "1.61",
        "FQDN": "bmc01.example.com"
      }]
    }
  }
}
```

**2. CANI Portable Format (after Transform)**

```json
{
  "946a7d44-9967-4940-9490-f2d581950512": {
    "id": "946a7d44-9967-4940-9490-f2d581950512",
    "name": "cani-device-bmc01",
    "slug": "hpe-proliant-dl325-gen11",
    "manufacturer": "HPE",
    "model": "ProLiant DL325 Gen11",
    "hardwareType": "node",
    "status": "staged",
    "providerMetadata": {
      "redfishFQDN": "bmc01.example.com",
      "managerType": "iLO 6",
      "firmwareVersion": "1.61"
    }
  }
}
```

**3. Target — Nautobot Device (after Export)**

```json
{
  "name": "cani-device-bmc01",
  "device_type": { "slug": "hpe-proliant-dl325-gen11" },
  "status": { "name": "Staged" },
  "location": { "name": "Default" },
  "role": { "name": "Server" }
}
```

The ETL flow:

```
┌─────────┐     ┌───────────┐     ┌──────────────┐     ┌───────────┐
│  Import  │────▶│ Transform │────▶│  CANI Store  │────▶│  Export   │
│ (Extract)│     │           │     │  (portable)  │     │  (Load)   │
└─────────┘     └───────────┘     └──────────────┘     └───────────┘
  Redfish          Provider          UUID-keyed          Nautobot
  CSM API          logic             devices, racks,     CSM API
  CSV/YAML         (custom)          modules, cables     CSV
  Ochami                             FRUs, interfaces
```

---

# Quick Start

## Import, View, and Export

```shell
# Import from a Redfish ServiceRoot file
cani alpha import redfish --root ./redfish-roots.json

# See what was imported
cani alpha show device

# Add a rack for the orphaned devices
cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack

# Resolve orphans (devices without a parent rack)
cani alpha update orphans --dry-run
cani alpha update orphans --apply-plan ~/.cani/resolve-plan.json

# Export to Nautobot
cani alpha export nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN
```

## Multi-Source Workflow

```shell
# Import from Redfish BMCs
cani alpha import redfish --root ./redfish-roots.json

# Import from HPCM cluster definition
cani alpha import hpcm --node-json-file ./hpcm-nodes.json

# Classify any unresolved device types
cani alpha classify --auto

# Resolve orphans
cani alpha update orphans

# Export the combined inventory to Nautobot
cani alpha export nautobot
```

## Adding Hardware by Slug

Adding hardware only requires a device-type slug. Metadata is auto-populated from the device-type library:

```shell
# Add a server — interfaces, power ports, and module bays are auto-populated
cani alpha add hpe-xl225n-gen10-plus

# Add a rack — the slug is all you need
cani alpha add rack hpe-ex4000

# Add multiple devices at once
cani alpha add device hpe-proliant-dl325-gen11 --qty 4

# List supported types
cani alpha add device --list-supported-types
```

## Visual Output

```shell
# Full ASCII rack diagram
cani alpha show --visual

# Compact symbolic view (S=switch, N=node, B=blade, C=CDU, P=PDU)
cani alpha show --rack-view

# Cable routing between racks
cani alpha show --show-routing -VV
```

---

# Portable Inventory Format

The inventory is a JSON datastore keyed by UUID. Every hardware item is self-contained with its metadata and relationships.

The `Inventory` contains seven maps:

| Map | Type | Purpose |
|-----|------|---------|
| `locations` | `CaniLocationType` | Sites, buildings, floors, rooms |
| `racks` | `CaniRackType` | Physical racks and cabinets |
| `devices` | `CaniDeviceType` | Servers, switches, blades, chassis, PDUs |
| `modules` | `CaniModuleType` | NICs, GPUs, PSUs, and other internal modules |
| `cables` | `CaniCableType` | Network and power cables with terminations |
| `frus` | `CaniFruType` | Field-replaceable units (spares tracking) |
| `interfaces` | `InterfaceInstance` | Network interfaces (1GbE, 10GbE, 100GbE, etc.) |

**Example device entry:**

```json
{
  "946a7d44-9967-4940-9490-f2d581950512": {
    "id": "946a7d44-9967-4940-9490-f2d581950512",
    "name": "compute-node-01",
    "slug": "hpe-xl225n-gen10-plus",
    "partNumber": "P21163-B21",
    "manufacturer": "HPE",
    "model": "ProLiant XL225n Gen10 Plus",
    "hardwareType": "node",
    "status": "staged",
    "subdeviceRole": "child",
    "parent": "4c7e02e7-5068-4dc2-b727-089a6b11eb66",
    "rack": "a1b2c3d4-5678-9abc-def0-123456789abc",
    "rackPosition": 12,
    "face": "front"
  }
}
```

Parent/child relationships form a hierarchy:

```
Location (site/building/floor/room)
  └── Rack
       └── Device (chassis, switch, standalone server)
            ├── Device (blade — subdeviceRole: child)
            │    └── Module (NIC, GPU in a module bay)
            │         └── FRU (field-replaceable spare)
            └── Interface (management, data, fabric)
```

---

# Migrating from Legacy CANI

If you used the previous version of `cani` (session-based workflow), here is what changed:

## Command Changes

| Legacy Command | New Command | Notes |
|---------------|-------------|-------|
| `cani session init csm` | `cani alpha import csm` | Sessions removed; import directly |
| `cani session apply --commit` | Changes save automatically | No explicit apply step |
| `cani add cabinet <slug>` | `cani alpha add rack <slug>` | Generic noun: `rack` |
| `cani add blade <slug>` | `cani alpha add device <slug>` | Generic noun: `device` |
| `cani add node <slug>` | `cani alpha add device <slug>` | Generic noun: `device` |
| `cani list` | `cani alpha show device` | Renamed to `show` |
| `cani export --format <fmt>` | `cani alpha export <provider>` | Provider is a subcommand |

## Key Differences

- **Sessions removed** — Changes are saved to the datastore immediately. No `session init` or `session apply` required.
- **Generic hardware nouns** — Hardware-specific commands (`add blade`, `add cabinet`, `add node`) are replaced by generic types (`add device`, `add rack`, `add location`, `add module`, `add cable`).
- **New providers** — Nautobot, Redfish, Ochami, and Example providers are now available alongside CSM and HPCM.
- **Auto-classification** — The `classify` command can automatically resolve untyped devices.
- **Same config location** — Config remains at `~/.cani/cani.yml` and the datastore at `~/.cani/canidb.json`.

---

# Architecture

```
┌──────────────────────────────────────────────────────────┐
│                      CLI (cobra)                         │
│  add / remove / update / show / import / export / init   │
└────────────────────────┬─────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
   ┌──────────┐   ┌──────────┐   ┌──────────┐
   │   CSM    │   │ Nautobot │   │ Redfish  │  ...providers
   │ provider │   │ provider │   │ provider │
   └────┬─────┘   └────┬─────┘   └────┬─────┘
        │               │               │
        └───────────────┼───────────────┘
                        ▼
              ┌──────────────────┐
              │   CANI Inventory │
              │   (portable fmt) │
              └────────┬─────────┘
                       ▼
              ┌──────────────────┐
              │    Datastore     │
              │  (JSON / future  │
              │   PostgreSQL)    │
              └──────────────────┘
```

**Provider interface** — Each provider implements three concerns:

| Method | Purpose |
|--------|---------|
| `Import()` | Extract data from the external source (files, APIs) |
| `Transform()` | Convert external data into CANI's portable format |
| `Export()` | Sync the CANI inventory to the external target |

CRUD operations (`add`, `remove`, `update`, `show`) are provider-independent — they operate directly on the portable inventory.

---

# Building

```bash
make bin    # build the binary to ./bin/cani
```

# Configuration

Config file: `~/.cani/cani.yml`

```yaml
datastore: ~/.cani/canidb.json
debug: false
strict: false
providers:
  nautobot:
    url: http://localhost:8081
    token: ""
  csm:
    sls-file: ""
    smd-file: ""
```

**Precedence:** CLI flags > environment variables (`CANI_*`) > config file > defaults

Provider-specific options are auto-populated in the config file when a provider is first used.

# License

MIT — see [LICENSE](LICENSE) for details.
