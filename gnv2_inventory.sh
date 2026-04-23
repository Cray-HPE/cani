#!/usr/bin/env bash
# make nautobot-down && make nautobot-up
# rm -f ~/.cani/canidb.json

# create locations (required)
bin/cani alpha add location dc --name "CAN-QBC-Q1"
bin/cani alpha add location level --name "L3" --parent "CAN-QBC-Q1"
bin/cani alpha add location section --name "NON-MSFT" --parent "L3" --content-types "rack,device,module"

# create roles (required)
bin/cani alpha add metadata role ComputeNode --content-types dcim.device
bin/cani alpha add metadata role ServiceNode --content-types dcim.device
bin/cani alpha add metadata role Gateway --content-types dcim.device
bin/cani alpha add metadata role ManagementSwitch --content-types dcim.device
bin/cani alpha add metadata role HSNSwitch --content-types dcim.device
bin/cani alpha add metadata role CDU --content-types dcim.device
bin/cani alpha add metadata role PDU --content-types dcim.device

# add racks (required)
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "NON-MSFT" --status "Available" --name x3701
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "NON-MSFT" --status "Available" --name x3507
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "NON-MSFT" --status "Available" --name x3508
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "NON-MSFT" --status "Available" --name x3516

# populate x3701 with devices
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "x3701" --face rear --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch  --status Active
bin/cani alpha add device hpe-aruba-6300m-48g --rack "x3701" --face rear --position 47 --name "FORGE-%{RACK}u%{U}M" --metadata role=ManagementSwitch  --status Active 
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3701" --face rear --position 45 --name "FORGE-%{RACK}u%{U}L" --metadata role=Gateway --status Active --serial TW39KM301C
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3701" --face rear --position 46 --name "FORGE-%{RACK}u%{U}L" --metadata role=Gateway --status Active
# 2x Nvidia InfiniBand NDR switches (rear, 1RU each) at U43, U42
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "x3701" --face rear --position 43 --name "HSNS-%{RACK}u%{U}" --metadata role=HSNSwitch --status Active --serial 1I033601TL
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "x3701" --face rear --position 42 --name "HSNL-%{RACK}u%{U}" --metadata role=HSNSwitch --status Active --serial 1I03360215
# 4x HPE Cray XD670 DLC GPU nodes (front, 5RU each) at U34, U26, U18, U10
bin/cani alpha add device hpe-xd670 --rack "x3701" --face front --position 34 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --status Active --serial 5UF435KF42
bin/cani alpha add device hpe-xd670 --rack "x3701" --face front --position 26 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --status Active --serial 5UF435KF41
bin/cani alpha add device hpe-xd670 --rack "x3701" --face front --position 18 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --status Active --serial 5UF435KF40
bin/cani alpha add device hpe-xd670 --rack "x3701" --face front --position 10 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --status Active --serial 5UF435KF3Z
bin/cani alpha add device motivair-mcdu-4u --rack "x3701" --face rear --position 3 --name "CDU-%{RACK}u%{U}" --metadata role=CDU --status Active --serial MCDU-4U-F-R2-2024-0675441
bin/cani alpha add device hpe-metered-3ph-rack-pdu --rack "x3701" --face rear --name "%{RACK}-RPDU-A" --metadata role=PDU --status Active --serial 1JO3800069
bin/cani alpha add device hpe-metered-3ph-rack-pdu --rack "x3701" --face rear --name "%{RACK}-RPDU-B" --metadata role=PDU --status Active --serial 1JO3C00453
bin/cani alpha add device hpe-metered-3ph-rack-pdu --rack "x3701" --face rear --name "%{RACK}-RPDU-C" --metadata role=PDU --status Active --serial 1JO4400155
# add gpu modules
bin/cani alpha add module nvidia-h100-sxm-gpu --device '%{FILL}' --name 'gpu-%{DEVICE}-%{BAY}'
# add ConnectX-6 100GbE NIC modules in PCIe 9 slot of all XD670s in x3701
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-x3701u34" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-x3701u26" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-x3701u18" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-x3701u10" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active

# populate x3507 with devices
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "x3507" --face rear --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch  --status Active
bin/cani alpha add device hpe-aruba-6300m-48g --rack "x3507" --face rear --position 47 --name "FORGE-%{RACK}u%{U}M" --metadata role=ManagementSwitch  --status Active
bin/cani alpha add device hpe-aruba-9300-32d --rack "x3507" --face rear --position 46 --name "FORGE-%{RACK}u%{U}S" --metadata role=Gateway --status Active
bin/cani alpha add device hpe-aruba-9300-32d --rack "x3507" --face rear --position 45 --name "FORGE-%{RACK}u%{U}S" --metadata role=Gateway --status Active
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3507" --face rear --position 44 --name "FORGE-%{RACK}u%{U}L" --metadata role=Gateway --status Active --serial TW39KM301C
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3507" --face rear --position 43 --name "FORGE-%{RACK}u%{U}L" --metadata role=Gateway --status Active
# 2x Nvidia InfiniBand NDR switches (rear, 1RU each) at U43, U42 and an NVIDIA UFM Appliance 3.0 (rear, 1RU) at U41
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "x3507" --face rear --position 41 --name "HSNS-%{RACK}u%{U}" --metadata role=HSNSwitch --status Active --serial 1I033601Q3
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "x3507" --face rear --position 40 --name "HSNL-%{RACK}u%{U}" --metadata role=HSNSwitch --status Active --serial 1I033601Q7
# 2x DL380 Gen11 blade servers (front, 2RU each) at U25, U23
# bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 25 --name "DL-%{RACK}u%{U}" --metadata role=ComputeNode --status Active --serial 2M240402LD
# bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 23 --name "DL-%{RACK}u%{U}" --metadata role=ComputeNode --status Active --serial 2M240402LC
# 9x DL380 Gen11 service nodes (front, 2RU each) at U21, U19, U17, U15, U13, U11, U9, U7, U5
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 21 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 19 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 17 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 15 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 13 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 11 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 9 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 7 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "x3507" --face front --position 5 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
# add ConnectX-6 100GbE NIC modules in PCIe5 slot of all DL380s
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u21" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u19" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u17" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u15" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u13" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u11" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u9" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u7" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-x3507u5" --bay "PCIe5" --name "CX6-%{DEVICE}" --status Active

# populate x3508 with devices
# 1x Aruba 2930F management switch (rear, 1RU) at U48
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "x3508" --face rear --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
# 2x Aruba 8325-32C backbone-rear switches (rear, 1RU each) at U47, U46
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 47 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 46 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
# 1x Fortinet FortiGate 4401F (front, 4RU) at U41 — no device-type slug defined
bin/cani alpha add device fortinet-fortigate-4401f --rack "x3508" --face front --position 41 --name "FORT-%{RACK}u%{U}" --metadata role=Gateway --status Active
# 2x Aruba 8325-32C backbone-spine switches (rear, 1RU each) at U39, U38
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 39 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 38 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
# 1x F5 Networks r10000 load balancer (front, 1RU) at U36 — no device-type slug defined
bin/cani alpha add device f5-networks-r10000 --rack "x3508" --face front --position 36 --name "F5-%{RACK}u%{U}" --metadata role=Gateway --status Active
# 3x Aruba 8325-48Y8C management-backbone switches (rear, 1RU each) at U27, U26, U25
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "x3508" --face rear --position 27 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "x3508" --face rear --position 26 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "x3508" --face rear --position 25 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
# 4x Aruba 8325-32C backbone-backbone switches (rear, 1RU each) at U11, U10, U8, U7
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 11 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 10 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 8 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3508" --face rear --position 7 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active
# 2x Raritan PX3-1215-N1Q1V2K002 rack PDUs (ZeroU) — no device-type slug defined
bin/cani alpha add device raritan-px3-1215 --rack "x3508" --face rear --name "%{RACK}-RPDU-A" --metadata role=PDU --status Active
bin/cani alpha add device raritan-px3-1215 --rack "x3508" --face rear --name "%{RACK}-RPDU-B" --metadata role=PDU --status Active

# populate x3516 with devices
# 1x Aruba 2930F management switch (front, 1RU) at U48
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "x3516" --face front --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW29HKV130
# 2x Aruba 8325-32C backbone-rear switches (front, 1RU each) at U47, U46
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 47 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW3AKM301S
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 46 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW35KM303R
# 1x Fortinet FortiGate 4401F (front, 4RU) at U41
bin/cani alpha add device fortinet-fortigate-4401f --rack "x3516" --face front --position 41 --name "FORT-%{RACK}u%{U}" --metadata role=Gateway --status Active --serial FG441FTK22900384
# 2x Aruba 8325-32C backbone-spine switches (front, 1RU each) at U39, U38
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 39 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW35KM304T
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 38 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW33KM300H
# 1x F5 Networks r10000 load balancer (front, 1RU) at U36
bin/cani alpha add device f5-networks-r10000 --rack "x3516" --face front --position 36 --name "F5-%{RACK}" --metadata role=Gateway --status Active --serial f5-zphp-pqdk
# 4x Aruba 8325-48Y8C management-backbone switches (front, 1RU each) at U35, U27, U26, U25 
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "x3516" --face front --position 35 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW29KM00QW
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "x3516" --face front --position 27 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW32KM005Q
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "x3516" --face front --position 26 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW35KM000F
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "x3516" --face front --position 25 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW32KM004L
# 4x Aruba 8325-32C backbone-backbone switches (front, 1RU each) at U11, U10, U8, U7
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 11 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW35KM303Q
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 10 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW35KM303G
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 8 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW33KM301L
bin/cani alpha add device hpe-aruba-8325-32c --rack "x3516" --face front --position 7 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Active --serial TW35KM302L
# 2x Raritan PX3-1215-N1Q1V2K002 rack PDUs (ZeroU)
bin/cani alpha add device raritan-px3-1215 --rack "x3516" --face front --name "%{RACK}-RPDU-A" --metadata role=PDU --status Active --serial 1JO1800026
bin/cani alpha add device raritan-px3-1215 --rack "x3516" --face front --name "%{RACK}-RPDU-B" --metadata role=PDU --status Active --serial 1JO3600078


###############################################################################
# CONNECTIONS — Spine-Leaf Topology
#
# x3516 is the core networking rack (spine). All racks home back to x3516.
#
# Topology layers:
#   Layer 1  — Management (1GbE):  device iLO → local MAN switch → MANB-x3516
#   Layer 2  — Backbone  (100GbE): BBL leaf → BBS-x3516 spine → BBB-x3516 super-spine
#   Layer 3  — HSN       (400G NDR): XD670 HSN → local NDR leaf/spine → cross-rack ISLs
#   Gateway  — FortiGate/F5 → backbone switches in x3516
#
# Cable color convention:
#   blue   = management         green  = HSN intra-rack (400G NDR)
#   orange = backbone (100GbE)  yellow = HSN cross-rack ISLs (400G NDR fiber)
#   red    = gateway / security
#
# Three methods are demonstrated below:
#   Method 1 — Individual 'add cable' CLI commands (one cable at a time)
#   Method 2 — Declarative YAML file via 'add connections' (bulk import)
#   Method 3 — Auto-generated topology via 'add connections generate'
###############################################################################


# ─────────────────────────────────────────────────────────────────────────────
# METHOD 1: Individual cable commands
#
# Use 'add cable <cable-type-slug>' with --a-device/--a-port and
# --b-device/--b-port for one-off or special-purpose cables.
# Supports --label, --color, --name, and --status flags.
# ─────────────────────────────────────────────────────────────────────────────

# # FortiGate HA heartbeat pair — x3516 ↔ x3508 (10G SFP+ DAC, red)
# bin/cani alpha add cable hpe-aruba-10g-sfpp-3m-dac \
#   --a-device "FORT-x3516u41" --a-port "ha1" \
#   --b-device "FORT-x3508u41" --b-port "ha1" \
#   --color red --label "FW-HA-1"
# bin/cani alpha add cable hpe-aruba-10g-sfpp-3m-dac \
#   --a-device "FORT-x3516u41" --a-port "ha2" \
#   --b-device "FORT-x3508u41" --b-port "ha2" \
#   --color red --label "FW-HA-2"

# # NDR spine switch management — single cable example (CAT6, blue)
# bin/cani alpha add cable hpe-3m-cat6-stp \
#   --a-device "HSNS-x3701u43" --a-port "mgmt0" \
#   --b-device "MAN-x3701u48" --b-port "5" \
#   --color blue --label "NDR-MGMT-x3701-SPINE"

# # UFM appliance management — second mgmt port to x3508 MAN switch (CAT6, blue)
# bin/cani alpha add cable hpe-3m-cat6-stp \
#   --a-device "UFM-x3507" --a-port "mgmt1" \
#   --b-device "MAN-x3508u48" --b-port "7" \
#   --color blue --label "UFM-MGMT-REDUNDANT"

# # XD670 100G management port → backbone leaf switch (100G QSFP28 DAC, blue)
# # The XD670 MGMT 0 port is 100GbE — connects to the backbone for in-band management
# bin/cani alpha add cable hpe-aruba-100g-qsfp28-3m-dac \
#   --a-device "GH-x3701u34" --a-port "MGMT 0" \
#   --b-device "BBL-x3701u45" --b-port "1/1/20" \
#   --color blue --label "XD670-INBAND-MGMT"

# # DL380 compute node Gig-E data port → MANB switch in x3516 (CAT5 4.3m, blue)
# bin/cani alpha add cable hpe-cat5-rj45-4-3m-cable \
#   --a-device "DL-x3507u25" --a-port "Gig-E 1" \
#   --b-device "MANB-x3516u27" --b-port "1/1/3" \
#   --color blue --label "DL380-DATA"


# ─────────────────────────────────────────────────────────────────────────────
# METHOD 2: Declarative YAML connection map
#
# Bulk-create connections from a YAML file. Supports brace-expansion patterns
# for compact definitions (e.g., "GH-x3701u{34,26,18,10}" expands to 4 devices).
# The YAML file defines all management, backbone, HSN, and gateway connections.
# ─────────────────────────────────────────────────────────────────────────────

# # Preview what would be created (dry-run — no changes to inventory)
# bin/cani alpha add connections connections-gn.yml --dry-run

# # Apply all connections from the YAML file
# bin/cani alpha add connections connections-gn.yml


# ─────────────────────────────────────────────────────────────────────────────
# METHOD 3: Auto-generated topology patterns
#
# Generate connection maps from topology patterns instead of writing YAML
# by hand. Supports: leaf-spine, star, ring.
# NOTE: These examples overlap with Method 2 — in practice, use one or the
# other for each connection set, not both.
# ─────────────────────────────────────────────────────────────────────────────

# --- Leaf-Spine: Backbone fabric ---
# 4 leaf switches, 2 spine switches, 2 uplinks per leaf per spine
# Generates 16 connections (4 leaves × 2 spines × 2 uplinks)
# bin/cani alpha add connections generate leaf-spine \
#   --leaves "BBL-x3701u45" --leaves "BBL-x3702u45" \
#   --leaves "BBR-x3516u47" --leaves "BBR-x3516u46" \
#   --spines "BBS-x3516u39" --spines "BBS-x3516u38" \
#   --uplinks-per-leaf 2 \
#   --cable-type hpe-aruba-100g-qsfp28-15m-aoc \
#   --cable-color orange

# --- Star: Management hub-spoke ---
# MAN-x3516u48 is the hub; x3507 devices (no local mgmt switch) are spokes
# Hub ports 7-17 connect to 11 spoke devices' iLO ports
# bin/cani alpha add connections generate star \
#   --hub "MAN-x3516u48" \
#   --hub-ports "{7..17}" \
#   --spokes "DL-x3507u25" --spokes "DL-x3507u23" \
#   --spokes "SERV-x3507u21" --spokes "SERV-x3507u19" \
#   --spokes "SERV-x3507u17" --spokes "SERV-x3507u15" \
#   --spokes "SERV-x3507u13" --spokes "SERV-x3507u11" \
#   --spokes "SERV-x3507u9" --spokes "SERV-x3507u7" \
#   --spokes "SERV-x3507u5" \
#   --spoke-port "iLO" \
#   --cable-type hpe-cat5-rj45-4-3m-cable \
#   --cable-color blue

# --- Ring: NDR spine-to-spine ISL ring ---
# Creates a ring of ISL links between all 3 NDR spine switches
# bin/cani alpha add connections generate ring \
#   --devices "HSNS-x3701u43" --devices "HSNS-x3702u43" \
#   --devices "HSNS-x3507u41" \
#   --port-a "41" \
#   --port-b "42" \
#   --cable-type hpe-ib-ndr-mpo-mpo-sm-15m \
#   --cable-color yellow


# show ASCII diagram of rack elevations
# bin/cani alpha show rack --visual

# export to nautobot
# bin/cani alpha export nautobot
