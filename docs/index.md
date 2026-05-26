# CANI Overview

**C**ontinuous **A**nd **N**ever-ending **I**nventory

## In One Sentence, What Is It?

A miniature DCIM (Datacenter Infrastructure Management).

### Tell Me More

**CANI** (Continuous And Never-ending Inventory) is a provider-agnostic hardware inventory tool built on an ETL (Extract-Transform-Load) pipeline. It gives you a single CLI to import inventory from any supported source, manage it in a portable format, and export it to any supported target.

---

## Three Things to Know

1. **ETL Pipeline** — Import from any source, export to any target; the inventory format in the middle is always the same.
2. **Devicetype Library** — A YAML catalog of real hardware (servers, switches, GPUs, cables) that auto-populates interfaces and physical attributes when you add a device by slug.
3. **Uniform CLI** — One set of commands (`add`, `remove`, `update`, `show`) works identically regardless of provider.

### 1. Provider-Agnostic ETL Pipeline

CANI separates *where your data comes from* and *where it goes*. You can import from one system and export to another — or manage everything offline from scratch.

**Supported providers:**

| Provider | Description |
|----------|-------------|
| `csm` | Cray System Management (SLS / HSM APIs) |
| `hpcm` | HPE Performance Cluster Manager |
| `nautobot` | Nautobot REST API (Netbox-compatible DCIM) |
| `ochami` | OpenCHAMI hardware database |
| `redfish` | Redfish BMC / iLO discovery |
| `example` | Reference implementation for building new providers |

The ETL flow works like this:

```
  Import (Extract)          Transform              Export (Load)
┌──────────────────┐   ┌─────────────────┐   ┌──────────────────┐
│  CSM / Redfish / │──▶│  Normalize into │──▶│  Nautobot / CSM  │
│  CSV / Nautobot  │   │  CANI inventory │   │  / HPCM / etc.   │
└──────────────────┘   └─────────────────┘   └──────────────────┘
```

You can mix and match providers freely. Import from Redfish, export to Nautobot. Import from CSM, export to Ochami. The inventory format in the middle is always the same.


### 2. YAML-Based Devicetype Library

The devicetype library is the foundational piece of the software. It models real-world hardware expectations in code: when you buy a ProLiant DL385 Gen11, you know it ships with four onboard GbE ports, an iLO management port, two PSU bays, a FlexLOM slot, and up to eight PCIe slots across three risers. CANI encodes those same assumptions in a simple YAML file:

```yaml
manufacturer: HPE
model: ProLiant DL385 Gen11 8SFF
slug: hpe-proliant-dl385-gen11-8sff
part_number: P53921-B21
u_height: 2
is_full_depth: true

module-bays:
  - name: PSU1
    position: PSU1
  - name: PSU2
    position: PSU2
  - name: FlexLOM
    position: FlexLOM
  - name: PCIe1
    position: PCIe1
    label: primary riser
  # ... PCIe2-PCIe8 across primary/secondary/tertiary risers

interfaces:
  - name: Gig-E 1
    type: 1000base-t
  - name: Gig-E 2
    type: 1000base-t
  - name: Gig-E 3
    type: 1000base-t
  - name: Gig-E 4
    type: 1000base-t
  - name: iLO
    type: 1000base-t
    mgmt_only: true
```

When you run `cani add device hpe-proliant-dl385-gen11-8sff`, the tool looks up this slug and instantiates the device with all its interfaces, module bays, and physical attributes already populated — just like unboxing the real thing.

And if new hardware is developed tomorrow, a user simply creates a new YAML file describing it. The CLI operations remain completely unchanged because they already know how to work with the common format. No code changes, no recompilation — just drop in a YAML file and the new hardware is instantly supported.

You can also load additional devicetype definitions from local directories or remote git repos at runtime using the `--types-dirs` and `--types-repos` flags.

#### Why This Matters

Nautobot (and its predecessor Netbox) is a widely used data center infrastructure management (DCIM) platform. Its data model is built on a hierarchy of core objects:

```
Locations  →  Racks  →  Devices  →  Modules  →  Interfaces  →  Cables
```

Without these fundamental building blocks populated, you cannot do much of anything useful in Nautobot — no IP address management, no circuit tracking, no topology maps, no automated provisioning. Getting these core objects right is a prerequisite for everything else.

**CANI is a stripped-down, offline version of this data model**, focused exclusively on these foundational objects. It gives you a simple interface to:

- **Start new** — Build a complete inventory from scratch before pushing it to Nautobot
- **Maintain existing** — Import from a running system, make changes offline, and push updates back

> CANI includes a lightweight IPAM (IP Address Management) layer for assigning prefixes, IP addresses, and VLANs to your inventory. It does not replicate Nautobot's full feature set (circuits, tenancy, etc.), but covers the physical inventory and basic network addressing that everything else depends on.

#### Library Structure

The devicetype library lives under `pkg/devicetypes/` and is organized by category and manufacturer:

```
pkg/devicetypes/
├── device-types/          # Servers, switches, chassis, controllers
│   ├── Cisco/
│   ├── HPE/
│   ├── NVIDIA/
│   └── ...
├── module-types/          # GPUs, NICs, PSUs, transceivers, memory
│   ├── HPE/
│   ├── NVIDIA/
│   └── ...
├── rack-types/            # Physical racks and cabinets
│   └── HPE/
├── cable-types/           # DACs, AOCs, fiber, copper
│   └── HPE/
├── location-types/        # Sites, buildings, rooms
└── connections/           # Topology templates
```

#### Core Type Categories

| Category | What It Defines | Examples |
|----------|----------------|----------|
| **location-type** | locations| dc, site, level/room |
| **rack-types** | Physical racks and cabinets | 42U enterprise racks, liquid-cooled cabinets |
| **device-types** | Servers, switches, chassis, controllers | ProLiant DL380, Cray EX blades, Cisco switches |
| **module-types** | Components installed in devices | GPUs, NICs, PSUs, transceivers, memory DIMMs |\
| **cable-types** | Physical cables | 100Gb QSFP28 DACs, Cat6 patch cables, fiber |

---

### 3. Uniform CLI for Every Provider

The same core commands work regardless of which provider you use:

```
cani add       # Add locations, racks, devices, modules, cables, connections
cani remove    # Remove inventory items
cani update    # Modify existing items
cani show      # Display inventory (tables, JSON, trees, ASCII rack diagrams)
```

Provider-specific behavior only appears during `import` and `export`. Everything else is provider-agnostic — you learn one set of commands and they work everywhere.

> **Note:** Many commands currently live under `cani alpha` (e.g., `cani alpha add`, `cani alpha export`). This prefix indicates the command surface is still evolving.

### Try It: Build an Inventory from Scratch

```bash
# 1. Create a location and a rack
cani alpha add location level --name "Server-Room-1"
cani alpha add rack hpe-42u-800mmx1200mm-g2-enterprise-shock-rack --name "rack1"

# 2. Add two servers
cani alpha add device hpe-proliant-dl325-gen11-8sff --rack "rack1" --name "DL380-1" --position 39
cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "rack1" --name "DL325-1" --position 40
cani alpha add device hpe-aruba-6300m-48g --rack "rack1" --name "Mgmt-Switch" --position 42

# 3. Connect the iLO ports to a management switch with a cable topology file
cat <<EOF > my-cables.yml
version: v1
cable_defaults:
  status: Connected
connections:
  - a: { device: "DL380-1", port: "iLO" }
    b: { device: "Mgmt-Switch", port: "1" }
    cable: { type: hpe-3m-cat6-stp, color: blue }
  - a: { device: "DL325-1", port: "iLO" }
    b: { device: "Mgmt-Switch", port: "2" }
    cable: { type: hpe-3m-cat6-stp, color: blue }
EOF
cani alpha add connections my-cables.yml

# 4. See what you built
cani alpha show rack -o minimap                   # Compact overview of all racks
cani alpha show rack "rack1" -o routing           # Rack diagram with cable routes
cani alpha show cable                             # Table of all cables
```

### Try It: Import from a Running System

```bash
# Pull inventory from Cray System Management (SLS)
cani alpha import csm --csm-url-sls https://api-gw-service-nmn.local/apis/sls/v1

# Inspect the imported racks
cani alpha show rack -o minmap

# Inspect the imported racks with cabling and labels
cani alpha show rack -o routing -l
```

---

### IPAM: IP Address Management

CANI includes a slim IPAM layer that lets you define subnets, assign IP addresses to device interfaces, and associate VLANs — all offline, in the same inventory file. This bridges the gap between "I have hardware in racks" and "I know what IP each management port uses."

The IPAM model has three objects:

| Object | Purpose | Example |
|--------|---------|---------|
| **Prefix** | An IPv4/IPv6 subnet in CIDR notation | `10.0.1.0/24` |
| **IP Address** | A single host address assigned to interface(s) | `10.0.1.10/24` on `node1:iLO` |
| **VLAN** | A layer-2 domain (ID 1–4094) optionally linked to a prefix | VLAN 100 "Management" |

Prefixes form a hierarchy automatically: `10.0.1.0/24` nests inside `10.0.0.0/16`. IP addresses are parented under their most-specific matching prefix.

```bash
# Define network infrastructure
cani alpha add vlan 100 --name "Management" --status active
cani alpha add prefix 10.0.0.0/16 --type container --role infrastructure
cani alpha add prefix 10.0.1.0/24 --type network --role management --vlan "Management"

# Assign IPs to device interfaces
cani alpha add ip 10.0.1.10/24 --interface "node1:iLO" --status active
cani alpha add ip 10.0.1.11/24 --interface "node2:iLO" --status active

# Reserve an IP for future use (no interface)
cani alpha add ip 10.0.1.254/24 --status reserved --description "gateway"

# Set a device's primary management address
cani alpha update device node1 --primary-ipv4 10.0.1.10/24

# View IPAM data
cani alpha show prefix           # Table of all prefixes
cani alpha show prefix --tree    # Hierarchical view
cani alpha show ip               # All IP addresses
cani alpha show vlan             # All VLANs
```

When you export to Nautobot, CANI maps these objects directly to Nautobot's Prefix, IP Address, and VLAN models — no manual data entry required on the Nautobot side.

---

## Writing a Provider Plugin

Adding a new provider to CANI is straightforward. The scaffold generator does most of the setup for you.

### Step 1: Generate the Skeleton

```bash
cani alpha init myprovider
```

This creates a complete directory structure under `pkg/provider/myprovider/` with all the files you need:

```
pkg/provider/myprovider/
├── init.go              # Registration and CLI command hook
├── provider.go          # Provider struct and Slug()
├── options.go           # Configuration options
├── import.go            # Import wrapper
├── export.go            # Export wrapper
├── transform.go         # Transform wrapper
├── commands/
│   └── commands.go      # CLI subcommands
├── export/
│   └── export.go        # Export logic
├── import/
│   └── import.go        # Import logic
└── transform/
    └── transform.go     # Transform logic
```

Every generated file includes `TODO` comments marking where to add your implementation.

### Step 2: Implement the Core Interface

A provider only needs to satisfy three methods:

```go
type Provider interface {
    // A short identifier for this provider (e.g., "myprovider")
    Slug() string

    // Convert imported data into CANI's portable format
    Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error)

    // Return CLI commands for provider-specific operations
    NewProviderCmd(base *cobra.Command) (*cobra.Command, error)
}
```

### Step 3: Add Optional Capabilities

Implement additional interfaces as needed:

| Interface | Purpose |
|-----------|---------|
| `Importer` | Pull data from an external source |
| `Exporter` | Push inventory to an external target |
| `HasOptions` | Expose default configuration options |
| `HasImportOptions` | Add import-specific CLI flags |
| `HasExportOptions` | Add export-specific CLI flags |
| `DeviceStager` | Auto-stage devices during the add workflow |
| `RackStager` | Create default devices when racks are added |
| `RackPostAddHook` | Run provider-specific logic after a rack is added |

### Step 4: Register and Build

Add one import line to `main.go` and build:

```go
import _ "github.com/Cray-HPE/cani/pkg/provider/myprovider"
```

```bash
make bin
```

Your provider is now available as `cani alpha import myprovider` and `cani alpha export myprovider`.

The `pkg/provider/example/` directory contains a fully working reference implementation you can study.

---

## Quick Reference

### Key Commands

| Command | Description |
|---------|-------------|
| `cani alpha add <type> <slug>` | Add a location, rack, device, module, or cable |
| `cani alpha add connections <file>` | Apply a cable topology from a YAML file |
| `cani alpha remove <type>` | Remove an inventory item |
| `cani alpha update <type>` | Modify an existing item |
| `cani show <type>` | Display inventory (table, JSON, tree, or rack diagram) |
| `cani alpha import <provider>` | Import inventory from an external system |
| `cani alpha export <provider>` | Export inventory to an external system |
| `cani alpha classify` | Auto-classify devices against the devicetype library |
| `cani alpha init <name>` | Scaffold a new provider plugin |

### Configuration

CANI stores its configuration and inventory at `~/.cani/`. Configuration values can be set via:

- Config file (`~/.cani/config.yaml`)
- Environment variables (prefix `CANI_`, e.g., `CANI_DATASTORE`)
- CLI flags (`--config`, `--datastore`, `--types-dirs`, etc.)

### Building from Source

```bash
make bin    # Produces ./bin/cani
```

CANI is pure Go with no third-party runtime dependencies beyond the standard library and a small set of well-known Go modules (cobra, viper, uuid, yaml).
