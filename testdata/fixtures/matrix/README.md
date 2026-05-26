# Matrix Test Fixtures

Reference topology for the import/export provider matrix test.

## Canonical Topology

A single-rack GPU cluster slice exercising compute, service, management, and
high-speed networking paths.

| Entity        | Count | Slug / Identifier                                |
|---------------|-------|--------------------------------------------------|
| Location      | 1     | matrix-site (type: site)                         |
| Rack          | 1     | hpe-48u-800mmx1200mm-g2-enterprise-shock-rack    |
| GPU nodes     | 2     | hpe-xd670 (matrix-gpu-01, matrix-gpu-02)         |
| Service nodes | 2     | hpe-proliant-dl380-gen11-8sff (matrix-serv-01/02)|
| Mgmt switch   | 1     | hpe-aruba-2930f-48g-4sfp (matrix-mgmt-sw)        |
| HSN switch    | 1     | nvidia-infiniband-ndr-64-port-osfp-switch        |
| GPU modules   | 16    | nvidia-h100-sxm-gpu (8 per XD670, GPU0–GPU7)    |
| CX-7 modules  | 2     | nvidia-connectx-7-ndr-infiniband-osfp-pcie5      |
| CX-6 modules  | 2     | nvidia-connectx-6-dx-100gbe-2p-qsfp28            |
| Cables        | 6     | 4× iLO→mgmt-sw, 2× HSN→hsn-sw                   |
| **Totals**    |       | **6 devices, 20 modules, 6 cables**              |

## Naming Convention

- GPU nodes: matrix-gpu-01, matrix-gpu-02
- Service nodes: matrix-serv-01, matrix-serv-02
- Switches: matrix-mgmt-sw, matrix-hsn-sw
- BMC addresses: 10.0.100.1–.4 (servers only)
- Serial numbers: SN-GPU-01, SN-GPU-02, SN-SERV-01, SN-SERV-02, etc.
- UUIDs: 11111111-1111-1111-1111-00000000000{1..6}
- Xnames (CSM): x9999 cabinet, x9999c0s{1,6,11,13}b0n0, x9999c0w{42,48}

## Cable Connectivity

| # | A-Device:Port           | B-Device:Port       | Type              |
|---|-------------------------|---------------------|-------------------|
| 1 | matrix-gpu-01:iLO       | matrix-mgmt-sw:1    | Cat6 management   |
| 2 | matrix-gpu-02:iLO       | matrix-mgmt-sw:2    | Cat6 management   |
| 3 | matrix-serv-01:iLO      | matrix-mgmt-sw:3    | Cat6 management   |
| 4 | matrix-serv-02:iLO      | matrix-mgmt-sw:4    | Cat6 management   |
| 5 | matrix-gpu-01:HSN 0     | matrix-hsn-sw:1     | NDR InfiniBand    |
| 6 | matrix-gpu-02:HSN 0     | matrix-hsn-sw:2     | NDR InfiniBand    |

## Per-Provider Representation

- **example.csv**: Full section-based CSV with all entities
- **ochami.json**: DiscoverySnapshot with flat rawData[] and parent-child via serialNumber
- **redfish.json**: Array of 4 ServiceRoot objects (servers only; switches not BMC-discovered)
- **hpcm.json**: Array of 6 node objects (all devices as nodes)
- **csm_sls.json**: SLS dumpstate with Hardware map + empty Networks

## Expected Entity Counts After Import

| Provider | Devices | Modules | Cables |
|----------|---------|---------|--------|
| example  | 6       | 20      | 6      |
| ochami   | 6       | 20      | 6      |
| redfish  | 4       | 0       | 0      |
| hpcm     | 6       | 0       | 0      |
| csm      | 6       | 0       | 0      |
