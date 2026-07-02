#!/usr/bin/env bash
# make nautobot-down && make nautobot-up
rm -f ~/.cani/canidb.json

# create locations (required)
bin/cani alpha add location dc --name "CAN-QBC-Q1"
bin/cani alpha add location level --name "L3" --parent "CAN-QBC-Q1"
bin/cani alpha add location section --name "NON-MSFT" --parent "L3" --content-types "rack,device,module"
bin/cani alpha add location section --name "POD5" --parent "L3" --content-types "rack,device,module"
bin/cani alpha add location section --name "POD7" --parent "L3" --content-types "rack,device,module"

# create roles (required)
# required by forge
bin/cani alpha add metadata role ComputeNode --content-types dcim.device
bin/cani alpha add metadata role ServiceNode --content-types dcim.device
bin/cani alpha add metadata role ManagementNode --content-types dcim.device

# required by forge 
bin/cani alpha add metadata role EdgeSwitch --content-types dcim.device
bin/cani alpha add metadata role LeafSwitch --content-types dcim.device
bin/cani alpha add metadata role SpineSwitch --content-types dcim.device
bin/cani alpha add metadata role LeafBMCSwitch --content-types dcim.device

# MgmtNetworkUplink: the VLAN-bearing construct (LAG interface when bonded, else single physical interface; never on LAG member ports). In-band mgmt-network NIC; cable traversal to leaf-side switch interface for Island VLAN application.
bin/cani alpha add metadata role MgmtNetworkUplink --content-types dcim.interface
# EdgeExternalUplink → dcim.Interface on edge devices (EdgeSwitch, or LeafSwitch+forge:hosts-edge. Customer-facing edge interfaces placed in EDGE VRF.
bin/cani alpha add metadata role EdgeExternalUplink --content-types dcim.interface
# UnderlayLinkPrefix → ipam.Prefix. AFC-only. Inter-switch P2P /31 source.
bin/cani alpha add metadata role UnderlayLinkPrefix --content-types ipam.prefix
# RouterIdLoopbackPrefix → ipam.Prefix (v4 or v6 — ip_version discriminates). Router-id loopback source; cross-provider semantic floor. AFC: one Prefix per ip_version, single range for all fabric devices. Apstra: two Prefixes per ip_version, each additionally tagged with one of the Apstra tier Tags below.
bin/cani alpha add metadata role RouterIdLoopbackPrefix --content-types ipam.prefix
# VtepSourceLoopbackPrefix → ipam.Prefix. AFC-only consumption. Separate VTEP source (Apstra collapses into RouterIdLoopback).
bin/cani alpha add metadata role VtepSourceLoopbackPrefix --content-types ipam.prefix
# AfcTransitVlan → ipam.VLAN (fabric Location scope). AFC-only. Fabric-mgmt transit VLAN.
bin/cani alpha add metadata role AfcTransitVlan --content-types ipam.vlan

# tags:
# forge:hosts-edge → dcim.Device (LeafSwitch). Marks leaf also hosting edge function (collapsed deployments).
# forge:apstra-spine-loopback-pool → ipam.Prefix (carrying RouterIdLoopbackPrefix Role). Apstra-only. Feeds spine_loopback_ips (v4) / spine_loopback_ips_ipv6 (v6); ip_version discriminates family.
# forge:apstra-leaf-loopback-pool → ipam.Prefix (carrying RouterIdLoopbackPrefix Role). Apstra-only. Feeds leaf_loopback_ips (v4) / leaf_loopback_ips_ipv6 (v6); ip_version discriminates family.
# forge:apstra-asn-pool → nautobot_bgp_models.AutonomousSystemRange. Apstra spine_asns/leaf_asns source. Tag-not-Role forced by upstream (no role FK).
# forge:apstra-vni-pool → nautobot_evpn_vxlan.VXLANPool. Apstra vni_ids (L2) / evpn_l3_vnis (L3) source; segment_type discriminates. Tag-not-Role forced by upstream.
bin/cani alpha add metadata role CDU --content-types dcim.device # not required by forge
bin/cani alpha add metadata role PDU --content-types dcim.device # not required by forge
bin/cani alpha add metadata role Unavailable --content-types dcim.device # for devices in rack, but not part of the cluster
bin/cani alpha add metadata role Storage --content-types dcim.device

# create interface roles (required for Nautobot dcim.interface)
bin/cani alpha add metadata role ManagementInterface --content-types dcim.interface
bin/cani alpha add metadata role HSNInterface --content-types dcim.interface
bin/cani alpha add metadata role DataInterface --content-types dcim.interface
bin/cani alpha add metadata role UplinkInterface --content-types dcim.interface
bin/cani alpha add metadata role StorageInterface --content-types dcim.interface
bin/cani alpha add metadata role BMCInterface --content-types dcim.interface

# add racks (required)
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "POD7" --status "Available" --name 3701
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "POD5" --status "Available" --name 3507
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "NON-MSFT" --status "Available" --name 3508
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "NON-MSFT" --status "Available" --name 3516
# netapp 3502 rack
bin/cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack --location "NON-MSFT" --status "Available" --name 3502
# populate 3502 with devices
# Network: 1x Aruba 8325-48Y8C management switch (front, 1RU) at U48
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3502" --face front --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
# Network: 2x Cisco Nexus 9336C-FX2 switches (front, 1RU each) at U31, U30
bin/cani alpha add device cisco-nexus-9336c-fx2 --rack "3502" --face front --position 31 --name "NAS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial FLM26460C4Y
bin/cani alpha add device cisco-nexus-9336c-fx2 --rack "3502" --face front --position 30 --name "NAS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial FLM26460C55
# Storage: 2x NetApp AFF A800 (front, 4RU each) at U26, U22
bin/cani alpha add device netapp-aff-a800 --rack "3502" --face front --position 26 --name "NAS-%{RACK}u%{U}" --metadata role=Storage --status Active --serial 952143000000
bin/cani alpha add device netapp-aff-a800 --rack "3502" --face front --position 22 --name "NAS-%{RACK}u%{U}" --metadata role=Storage --status Active --serial 952252000000
# Storage: 1x NetApp FAS8300 (front, 4RU) at U18
bin/cani alpha add device netapp-fas8300 --rack "3502" --face front --position 18 --name "NAS-%{RACK}u%{U}" --metadata role=Storage --status Active --serial 042240025642
# Storage: 3x NetApp DS460C disk shelves (front, 4RU each) at U14, U10, U6
bin/cani alpha add device netapp-ds460c --rack "3502" --face front --position 14 --name "NAS-%{RACK}u%{U}" --metadata role=Storage --status Active --serial SHJGD2309900133
bin/cani alpha add device netapp-ds460c --rack "3502" --face front --position 10 --name "NAS-%{RACK}u%{U}" --metadata role=Storage --status Active --serial SHJGD2309900135
bin/cani alpha add device netapp-ds460c --rack "3502" --face front --position 6 --name "NAS-%{RACK}u%{U}" --metadata role=Storage --status Active --serial SHJGD2305900263

# assign interface roles for rack 3502
bin/cani alpha update interface --device "MAN-3502u48" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MAN-3502u48" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "NAS-3502u31" --name "mgmt*" --role ManagementInterface
bin/cani alpha update interface --device "NAS-3502u31" --name "Ethernet*" --role StorageInterface
bin/cani alpha update interface --device "NAS-3502u30" --name "mgmt*" --role ManagementInterface
bin/cani alpha update interface --device "NAS-3502u30" --name "Ethernet*" --role StorageInterface

# populate 3701 with devices
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "3701" --face rear --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch  --status Planned
# NOTE: FORGE-3701u47M (Aruba 6300M) relocated to rack 3507 as FORGE-3507u37M (System Management Recabling); see the 3507 section
bin/cani alpha add device hpe-aruba-8325-32c --rack "3701" --face rear --position 45 --name "FORGE-%{RACK}u%{U}L" --metadata role=LeafSwitch --status Active --serial TW39KM301C
bin/cani alpha add device hpe-aruba-8325-32c --rack "3701" --face rear --position 46 --name "FORGE-%{RACK}u%{U}L" --metadata role=LeafSwitch --status Active --serial TW46KM30FS
# 2x Nvidia InfiniBand NDR switches (rear, 1RU each) at U43, U42
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "3701" --face rear --position 43 --name "HSNS-%{RACK}u%{U}" --metadata role=HSNSwitch --status Planned --serial 1I033601TL
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "3701" --face rear --position 42 --name "HSNL-%{RACK}u%{U}" --metadata role=HSNSwitch --status Planned --serial 1I03360215
# 4x HPE Cray XD670 DLC GPU nodes (front, 5RU each) at U34, U26, U18, U10
bin/cani alpha add device hpe-xd670 --rack "3701" --face front --position 34 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --metadata tenant=Forge --status Active --serial 5UF435KF42
bin/cani alpha add device hpe-xd670 --rack "3701" --face front --position 26 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --metadata tenant=Forge --status Active --serial 5UF435KF41
bin/cani alpha add device hpe-xd670 --rack "3701" --face front --position 18 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --metadata tenant=Forge --status Active --serial 5UF435KF40
bin/cani alpha add device hpe-xd670 --rack "3701" --face front --position 10 --name "GH-%{RACK}u%{U}" --metadata role=ComputeNode --metadata tenant=Forge --status Active --serial 5UF435KF3Z
bin/cani alpha add device motivair-mcdu-4u --rack "3701" --face rear --position 3 --name "CDU-%{RACK}u%{U}" --metadata role=CDU --status Active --serial MCDU-4U-F-R2-2024-0675441
bin/cani alpha add device hpe-metered-3ph-rack-pdu --rack "3701" --face rear --name "%{RACK}-RPDU-A" --metadata role=PDU --status Active --serial 1JO3800069
bin/cani alpha add device hpe-metered-3ph-rack-pdu --rack "3701" --face rear --name "%{RACK}-RPDU-B" --metadata role=PDU --status Active --serial 1JO3C00453
bin/cani alpha add device hpe-metered-3ph-rack-pdu --rack "3701" --face rear --name "%{RACK}-RPDU-C" --metadata role=PDU --status Active --serial 1JO4400155
# add gpu modules
bin/cani alpha add module nvidia-h100-sxm-gpu --device '%{FILL}' --name 'gpu-%{DEVICE}-%{BAY}'
# add ConnectX-6 100GbE NIC modules in PCIe 9 slot of all XD670s in x3701
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-3701u34" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-3701u26" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-3701u18" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "GH-3701u10" --bay "PCIe 9" --name "CX6-%{DEVICE}" --status Active

# assign interface roles for rack 3701
bin/cani alpha update interface --device "MAN-3701u48" --name "*" --role DataInterface
# FORGE-3701u47M interface roles moved to FORGE-3507u37M (see rack 3507 section)
bin/cani alpha update interface --device "FORGE-3701u46L" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3701u46L" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3701u45L" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3701u45L" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "HSNS-3701u43" --name "mgmt0" --role ManagementInterface
bin/cani alpha update interface --device "HSNS-3701u43" --name "[0-9]*" --role HSNInterface
bin/cani alpha update interface --device "HSNL-3701u42" --name "mgmt0" --role ManagementInterface
bin/cani alpha update interface --device "HSNL-3701u42" --name "[0-9]*" --role HSNInterface
bin/cani alpha update interface --device "GH-3701u34" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "GH-3701u34" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "GH-3701u26" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "GH-3701u26" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "GH-3701u18" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "GH-3701u18" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "GH-3701u10" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "GH-3701u10" --name "HSN *" --role HSNInterface

# populate 3507 with devices
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "3507" --face rear --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch  --status Planned
# PR 216 explicitly says "Do not add this switch to inventory"
# bin/cani alpha add device hpe-aruba-6300m-48g --rack "3507" --face rear --position 47 --name "FORGE-%{RACK}u%{U}M" --metadata role=LeafBMCSwitch  --status Planned --serial VN43M3W01X
# FORGE-3507u37M (Aruba 6300M) — switch-management aggregation, relocated from x3701 (System Management Recabling)
bin/cani alpha add device hpe-aruba-6300m-48g --rack "3507" --face rear --position 37 --name "FORGE-%{RACK}u%{U}M" --metadata role=LeafBMCSwitch  --status Planned --serial VN43M3W017
bin/cani alpha add device hpe-aruba-9300-32d --rack "3507" --face rear --position 46 --name "FORGE-%{RACK}u%{U}S" --metadata role=SpineSwitch --status Active --serial TW2AL6D01G
bin/cani alpha add device hpe-aruba-9300-32d --rack "3507" --face rear --position 45 --name "FORGE-%{RACK}u%{U}S" --metadata role=SpineSwitch --status Active --serial TW44L6D011
bin/cani alpha add device hpe-aruba-8325-32c --rack "3507" --face rear --position 44 --name "FORGE-%{RACK}u%{U}L" --metadata role=LeafSwitch --status Active --serial TW47KM30MH
bin/cani alpha add device hpe-aruba-8325-32c --rack "3507" --face rear --position 43 --name "FORGE-%{RACK}u%{U}L" --metadata role=LeafSwitch --status Active --serial TW44KM300R
# 2x Nvidia InfiniBand NDR switches (rear, 1RU each) at U43, U42 and an NVIDIA UFM Appliance 3.0 (rear, 1RU) at U41
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "3507" --face rear --position 41 --name "HSNS-%{RACK}u%{U}" --metadata role=HSNSwitch --status Planned --serial 1I033601Q3
bin/cani alpha add device nvidia-infiniband-ndr-64-port-osfp-switch --rack "3507" --face rear --position 40 --name "HSNL-%{RACK}u%{U}" --metadata role=HSNSwitch --status Planned --serial 1I033601Q7
# 2x DL380 Gen11 blade servers (front, 2RU each) at U25, U23
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 25 --name "dl-%{RACK}u%{U}" --metadata role=Unavailable --status Active --serial SBDYKQP7 # in-rack, but not part of the cluster
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 23 --name "dl-%{RACK}u%{U}" --metadata role=Unavailable --status Active --serial QCZYKQ22 # in-rack, but not part of the cluster
# 9x DL380 Gen11 service nodes (front, 2RU each) at U21, U19, U17, U15, U13, U11, U9, U7, U5
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 21 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 19 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 17 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 15 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 13 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 11 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 9 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 7 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
bin/cani alpha add device hpe-proliant-dl380-gen11-8sff --rack "3507" --face front --position 5 --name "SERV-%{RACK}u%{U}" --metadata role=ServiceNode --status Active 
# add ConnectX-6 100GbE NIC modules in PCIe1 slot of all DL380s
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u21" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u19" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u17" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u15" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u13" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u11" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u9" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u7" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-6-dx-100gbe-2p-qsfp28 --device "SERV-3507u5" --bay "PCIe1" --name "CX6-%{DEVICE}" --status Active
# add ConnectX-7 400GB NIC modules in PCIe2 slot of all DL380s
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u21" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u19" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u17" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u15" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u13" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u11" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u9" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u7" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active
bin/cani alpha add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 --device "SERV-3507u5" --bay "PCIe2" --name "CX7-%{DEVICE}" --status Active	

# assign interface roles for rack 3507
bin/cani alpha update interface --device "MAN-3507u48" --name "*" --role DataInterface
bin/cani alpha update interface --device "FORGE-3507u47M" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3507u47M" --name "1/1/49" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u47M" --name "1/1/50" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u47M" --name "1/1/51" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u47M" --name "1/1/52" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u37M" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3507u37M" --name "1/1/49" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u37M" --name "1/1/50" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u37M" --name "1/1/51" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u37M" --name "1/1/52" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u46S" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3507u46S" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u45S" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3507u45S" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u44L" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3507u44L" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "FORGE-3507u43L" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "FORGE-3507u43L" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "HSNS-3507u41" --name "mgmt0" --role ManagementInterface
bin/cani alpha update interface --device "HSNS-3507u41" --name "[0-9]*" --role HSNInterface
bin/cani alpha update interface --device "HSNL-3507u40" --name "mgmt0" --role ManagementInterface
bin/cani alpha update interface --device "HSNL-3507u40" --name "[0-9]*" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u21" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u21" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u21" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u19" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u19" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u19" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u17" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u17" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u17" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u15" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u15" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u15" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u13" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u13" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u13" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u11" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u11" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u11" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u9" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u9" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u9" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u7" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u7" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u7" --name "Gig-E *" --role DataInterface
bin/cani alpha update interface --device "SERV-3507u5" --name "iLO" --role ManagementInterface
bin/cani alpha update interface --device "SERV-3507u5" --name "HSN *" --role HSNInterface
bin/cani alpha update interface --device "SERV-3507u5" --name "Gig-E *" --role DataInterface

# populate 3508 with devices
# 1x Aruba 2930F management switch (rear, 1RU) at U48
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "3508" --face rear --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
# 2x Aruba 8325-32C backbone-rear switches (rear, 1RU each) at U47, U46
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 47 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 46 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
# 1x Fortinet FortiGate 4401F (front, 4RU) at U41 — no device-type slug defined
bin/cani alpha add device fortinet-fortigate-4401f --rack "3508" --face front --position 41 --name "FORT-%{RACK}u%{U}" --metadata role=Gateway --status Active
# 2x Aruba 8325-32C backbone-spine switches (rear, 1RU each) at U39, U38
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 39 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 38 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
# 1x F5 Networks r10000 load balancer (front, 1RU) at U36 — no device-type slug defined
bin/cani alpha add device f5-networks-r10000 --rack "3508" --face front --position 36 --name "F5-%{RACK}u%{U}" --metadata role=Gateway --status Active
# 3x Aruba 8325-48Y8C management-backbone switches (rear, 1RU each) at U27, U26, U25
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3508" --face rear --position 27 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3508" --face rear --position 26 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3508" --face rear --position 25 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
# 4x Aruba 8325-32C backbone-backbone switches (rear, 1RU each) at U11, U10, U8, U7
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 11 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 10 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 8 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
bin/cani alpha add device hpe-aruba-8325-32c --rack "3508" --face rear --position 7 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned
# 2x Raritan PX3-1215-N1Q1V2K002 rack PDUs (ZeroU) — no device-type slug defined
bin/cani alpha add device raritan-px3-1215 --rack "3508" --face rear --name "%{RACK}-RPDU-A" --metadata role=PDU --status Active
bin/cani alpha add device raritan-px3-1215 --rack "3508" --face rear --name "%{RACK}-RPDU-B" --metadata role=PDU --status Active

# assign interface roles for rack 3508
bin/cani alpha update interface --device "MAN-3508u48" --name "*" --role DataInterface
bin/cani alpha update interface --device "BBR-3508u47" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBR-3508u47" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBR-3508u46" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBR-3508u46" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBS-3508u39" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBS-3508u39" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBS-3508u38" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBS-3508u38" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "MANB-3508u27" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MANB-3508u27" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "MANB-3508u26" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MANB-3508u26" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "MANB-3508u25" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MANB-3508u25" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "BBB-3508u11" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3508u11" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBB-3508u10" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3508u10" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBB-3508u8" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3508u8" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBB-3508u7" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3508u7" --name "1/1/*" --role UplinkInterface

# populate 3516 with devices
# 1x Aruba 2930F management switch (front, 1RU) at U48
bin/cani alpha add device hpe-aruba-2930f-48g-4sfp --rack "3516" --face front --position 48 --name "MAN-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW29HKV130
# 2x Aruba 8325-32C backbone-rear switches (front, 1RU each) at U47, U46
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 47 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW3AKM301S
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 46 --name "BBR-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW35KM303R
# 1x Fortinet FortiGate 4401F (front, 4RU) at U41
bin/cani alpha add device fortinet-fortigate-4401f --rack "3516" --face front --position 41 --name "FORT-%{RACK}u%{U}" --metadata role=Gateway --status Active --serial FG441FTK22900384
# 2x Aruba 8325-32C backbone-spine switches (front, 1RU each) at U39, U38
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 39 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW35KM304T
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 38 --name "BBS-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW33KM300H
# 1x F5 Networks r10000 load balancer (front, 1RU) at U36
bin/cani alpha add device f5-networks-r10000 --rack "3516" --face front --position 36 --name "F5-%{RACK}" --metadata role=Gateway --status Active --serial f5-zphp-pqdk
# 4x Aruba 8325-48Y8C management-backbone switches (front, 1RU each) at U35, U27, U26, U25 
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3516" --face front --position 35 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW29KM00QW
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3516" --face front --position 27 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW32KM005Q
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3516" --face front --position 26 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW35KM000F
bin/cani alpha add device hpe-aruba-8325-48y8c --rack "3516" --face front --position 25 --name "MANB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW32KM004L
# 4x Aruba 8325-32C backbone-backbone switches (front, 1RU each) at U11, U10, U8, U7
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 11 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW35KM303Q
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 10 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW35KM303G
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 8 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW33KM301L
bin/cani alpha add device hpe-aruba-8325-32c --rack "3516" --face front --position 7 --name "BBB-%{RACK}u%{U}" --metadata role=ManagementSwitch --status Planned --serial TW35KM302L
# 2x Raritan PX3-1215-N1Q1V2K002 rack PDUs (ZeroU)
bin/cani alpha add device raritan-px3-1215 --rack "3516" --face front --name "%{RACK}-RPDU-A" --metadata role=PDU --status Active --serial 1JO1800026
bin/cani alpha add device raritan-px3-1215 --rack "3516" --face front --name "%{RACK}-RPDU-B" --metadata role=PDU --status Active --serial 1JO3600078

# assign interface roles for rack 3516
bin/cani alpha update interface --device "MAN-3516u48" --name "*" --role DataInterface
bin/cani alpha update interface --device "BBR-3516u47" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBR-3516u47" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBR-3516u46" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBR-3516u46" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBS-3516u39" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBS-3516u39" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBS-3516u38" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBS-3516u38" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "MANB-3516u35" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MANB-3516u35" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "MANB-3516u27" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MANB-3516u27" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "MANB-3516u26" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MANB-3516u26" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "MANB-3516u25" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "MANB-3516u25" --name "1/1/*" --role DataInterface
bin/cani alpha update interface --device "BBB-3516u11" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3516u11" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBB-3516u10" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3516u10" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBB-3516u8" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3516u8" --name "1/1/*" --role UplinkInterface
bin/cani alpha update interface --device "BBB-3516u7" --name "mgmt" --role ManagementInterface
bin/cani alpha update interface --device "BBB-3516u7" --name "1/1/*" --role UplinkInterface


# ═══════════════════════════════════════════════════════════════════════════
# FORGE FABRIC INTERFACE ROLES
# Applied AFTER the per-rack UplinkInterface/HSNInterface globs above so these
# forge-specific roles override the broad assignments on the affected ports.
# ═══════════════════════════════════════════════════════════════════════════

# EdgeExternalUplink → customer/external-facing leaf uplinks (EDGE VRF). This is
# a collapsed deployment: the leaves FORGE-3507u44L/43L host the edge function,
# and their northbound uplink to CFCANB1S1 (1/1/25) is the external edge uplink.
bin/cani alpha update interface --device "FORGE-3507u44L" --name "1/1/25" --role EdgeExternalUplink
bin/cani alpha update interface --device "FORGE-3507u43L" --name "1/1/25" --role EdgeExternalUplink

# MgmtNetworkUplink → in-band mgmt-network NIC on each node: the ConnectX-6
# 100GbE port(s) that cable to the leaf pair (distinct from the CX7 InfiniBand
# HSN). Targeted via --module so only the CX6 leaf-uplink ports are re-roled.
bin/cani alpha update interface --module "CX6-GH-3701u34" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-GH-3701u26" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-GH-3701u18" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-GH-3701u10" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u21" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u19" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u17" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u15" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u13" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u11" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u9" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u7" --name "HSN *" --role MgmtNetworkUplink
bin/cani alpha update interface --module "CX6-SERV-3507u5" --name "HSN *" --role MgmtNetworkUplink


# ═══════════════════════════════════════════════════════════════════════════
# IPAM: VLANs, IP Prefixes, and IP Addresses
# ═══════════════════════════════════════════════════════════════════════════

# --- VLANs ---
bin/cani alpha add vlan 1 --name "Default" --status Active --location "POD5"
bin/cani alpha add vlan 3 --name "Harvester" --status Active --location "POD5"
bin/cani alpha add vlan 5 --name "BMCs" --status Active --location "POD5"
bin/cani alpha add vlan 7 --name "Storage" --status Active --location "POD5"
bin/cani alpha add vlan 9 --name "FIPs" --status Active --location "POD5"
bin/cani alpha add vlan 10 --name "Transit" --status Active --location "POD5" --role AfcTransitVlan
bin/cani alpha add vlan 13 --name "BlackHole" --status Active --location "POD5" --description "Black hole VLAN 13"
bin/cani alpha add vlan 2000 --name "Legacy CSM CMN" --status Active --location "POD5" --description "Legacy CSM CMN"

# --- IP Prefixes (containers first, then networks, then host routes) ---

# Container prefixes (aggregates)
bin/cani alpha add prefix 10.101.48.0/22 --type container --description "Management" --status Active --location "POD5"
bin/cani alpha add prefix 10.1.0.0/16 --type container --description "Underlay wiring - P2P links and future infrastructure" --status Active --location "POD5"
bin/cani alpha add prefix 192.168.64.0/18 --type container --description "Fabric infrastructure - loopbacks" --status Active --location "POD5"
bin/cani alpha add prefix 172.17.0.0/16 --type container --description "Overlay workload segments - VLANs 3,5,7,9" --status Active --location "POD5"

# Network prefixes (subnets)
bin/cani alpha add prefix 10.1.255.0/24 --type network --role UnderlayLinkPrefix --description "RPIs spine<=>leaf" --status Active --location "POD5"
bin/cani alpha add prefix 192.168.64.0/24 --type network --role RouterIdLoopbackPrefix --description "Spine Loopback0 router-IDs" --status Active --location "POD5"
bin/cani alpha add prefix 10.2.0.0/24 --type network --role RouterIdLoopbackPrefix --description "Leaf Loopback0 BGP router-IDs" --status Active --location "POD5"
bin/cani alpha add prefix 172.17.2.0/23 --type network --description "VLAN 3 - Harvester" --status Reserved --vlan "Harvester" --location "POD5"
bin/cani alpha add prefix 172.17.4.0/23 --type network --description "VLAN 5 - BMCs" --status Reserved --vlan "BMCs" --location "POD5"
bin/cani alpha add prefix 172.17.6.0/23 --type network --description "VLAN 7 - Storage dataplane" --status Reserved --vlan "Storage" --location "POD5"
bin/cani alpha add prefix 10.102.13.0/24 --type network --description "VLAN 9 - FIPs" --status Active --vlan "FIPs" --location "POD5"
bin/cani alpha add prefix 10.102.12.0/24 --type network --description "VLAN 2000 - Legacy CSM CMN" --status Active --vlan "Legacy CSM CMN" --location "POD5"
bin/cani alpha add prefix 192.168.255.0/31 --type network --description "VSX Keepalive x3507" --status Reserved --location "POD5"
bin/cani alpha add prefix 192.168.255.2/31 --type network --description "VSX Keepalive x3701" --status Reserved --location "POD5"
bin/cani alpha add prefix 192.168.64.2/31 --type network --description "VLAN 10 interfaces on leaf switches" --status Reserved --location "POD5"

# Host route prefixes (per-device loopbacks)
bin/cani alpha add prefix 192.168.64.0/32 --type network --description "Loopback 0 FORGE-3507u46S" --status Reserved --location "POD5"
bin/cani alpha add prefix 192.168.64.1/32 --type network --description "Loopback 0 FORGE-3507u45S" --status Reserved --location "POD5"
bin/cani alpha add prefix 10.2.0.2/32 --type network --description "Loopback 0 FORGE-3507u44L" --status Reserved --location "POD5"
bin/cani alpha add prefix 10.2.0.3/32 --type network --description "Loopback 0 FORGE-3507u43L" --status Reserved --location "POD5"
bin/cani alpha add prefix 10.2.0.4/32 --type network --description "Loopback 0 FORGE-3701u46L" --status Reserved --location "POD5"
bin/cani alpha add prefix 10.2.0.5/32 --type network --description "Loopback 0 FORGE-3701u45L" --status Reserved --location "POD5"
bin/cani alpha add prefix 172.17.0.0/32 --type network --role VtepSourceLoopbackPrefix --description "Loopback 1 on leaf switches" --status Reserved --location "POD5"

# Upstream / edge prefixes
bin/cani alpha add prefix 10.102.30.48/30 --type network --description "CFCANB1S1 FORGE-3507u44L" --status Reserved --location "POD5"
bin/cani alpha add prefix 10.102.30.52/30 --type network --description "CFCANB1S1 FORGE-3507u43L" --status Reserved --location "POD5"

# --- IP Addresses ---

# Gateway / anycast VIPs
bin/cani alpha add ip 10.101.48.1/22 --status Reserved --role anycast --description "Default gateway for Management"
bin/cani alpha add ip 10.1.0.1/16 --status Reserved --role anycast --description "Default gateway for VLAN 1 and underlay"
bin/cani alpha add ip 172.17.2.1/23 --status Reserved --role anycast --description "Default gateway for VLAN 3"
bin/cani alpha add ip 172.17.4.1/23 --status Reserved --role anycast --description "Default gateway for VLAN 5"
bin/cani alpha add ip 172.17.6.1/23 --status Reserved --role anycast --description "Default gateway for VLAN 7"
bin/cani alpha add ip 10.102.13.1/24 --status Reserved --role anycast --description "Default gateway for VLAN 9"
bin/cani alpha add ip 10.102.12.1/24 --status Reserved --role anycast --description "Default gateway for VLAN 2000"

# Spine loopback IPs
bin/cani alpha add ip 192.168.64.0/32 --interface "FORGE-3507u46S:loopback0" --status Active --description "BGP router-ID"
bin/cani alpha add ip 192.168.64.1/32 --interface "FORGE-3507u45S:loopback0" --status Active --description "BGP router-ID"

# Leaf loopback IPs (loopback0 = BGP router-ID)
bin/cani alpha add ip 10.2.0.2/32 --interface "FORGE-3507u44L:loopback0" --status Active --description "BGP router-ID"
bin/cani alpha add ip 10.2.0.3/32 --interface "FORGE-3507u43L:loopback0" --status Active --description "BGP router-ID"
bin/cani alpha add ip 10.2.0.4/32 --interface "FORGE-3701u46L:loopback0" --status Active --description "BGP router-ID"
bin/cani alpha add ip 10.2.0.5/32 --interface "FORGE-3701u45L:loopback0" --status Active --description "BGP router-ID"

# Leaf loopback IPs (loopback1 = BGP VXLAN overlay, shared anycast)
bin/cani alpha add ip 172.17.0.0/32 --interface "FORGE-3507u44L:loopback1" --interface "FORGE-3507u43L:loopback1" --interface "FORGE-3701u46L:loopback1" --interface "FORGE-3701u45L:loopback1" --status Active --role anycast --description "BGP VXLAN overlay"

# Spine-leaf RPI point-to-point links (FORGE-3507u46S)
bin/cani alpha add ip 10.1.255.0/31 --interface "FORGE-3507u46S:1/1/1" --status Active --description "RPI to FORGE-3507u44L"
bin/cani alpha add ip 10.1.255.1/31 --interface "FORGE-3507u44L:1/1/29" --status Active --description "RPI to FORGE-3507u46S"
bin/cani alpha add ip 10.1.255.2/31 --interface "FORGE-3507u46S:1/1/2" --status Active --description "RPI to FORGE-3507u43L"
bin/cani alpha add ip 10.1.255.3/31 --interface "FORGE-3507u43L:1/1/29" --status Active --description "RPI to FORGE-3507u46S"
bin/cani alpha add ip 10.1.255.8/31 --interface "FORGE-3507u46S:1/1/3" --status Active --description "RPI to FORGE-3701u46L"
bin/cani alpha add ip 10.1.255.9/31 --interface "FORGE-3701u46L:1/1/29" --status Active --description "RPI to FORGE-3507u46S"
bin/cani alpha add ip 10.1.255.10/31 --interface "FORGE-3507u46S:1/1/4" --status Active --description "RPI to FORGE-3701u45L"
bin/cani alpha add ip 10.1.255.11/31 --interface "FORGE-3701u45L:1/1/29" --status Active --description "RPI to FORGE-3507u46S"

# Spine-leaf RPI point-to-point links (FORGE-3507u45S)
bin/cani alpha add ip 10.1.255.4/31 --interface "FORGE-3507u45S:1/1/1" --status Active --description "RPI to FORGE-3507u44L"
bin/cani alpha add ip 10.1.255.5/31 --interface "FORGE-3507u44L:1/1/28" --status Active --description "RPI to FORGE-3507u45S"
bin/cani alpha add ip 10.1.255.6/31 --interface "FORGE-3507u45S:1/1/2" --status Active --description "RPI to FORGE-3507u43L"
bin/cani alpha add ip 10.1.255.7/31 --interface "FORGE-3507u43L:1/1/28" --status Active --description "RPI to FORGE-3507u45S"
bin/cani alpha add ip 10.1.255.12/31 --interface "FORGE-3507u45S:1/1/3" --status Active --description "RPI to FORGE-3701u46L"
bin/cani alpha add ip 10.1.255.13/31 --interface "FORGE-3701u46L:1/1/28" --status Active --description "RPI to FORGE-3507u45S"
bin/cani alpha add ip 10.1.255.14/31 --interface "FORGE-3507u45S:1/1/4" --status Active --description "RPI to FORGE-3701u45L"
bin/cani alpha add ip 10.1.255.15/31 --interface "FORGE-3701u45L:1/1/28" --status Active --description "RPI to FORGE-3507u45S"

# VSX keepalive IPs
# FIXME (follow-up): the FORGE leaves are Aruba 8325-32C switches with only 32
# ports (1/1/1-1/1/32), so "1/1/48" below does not exist. Re-point the VSX
# keepalive to a real, unused port (or the dedicated OOBM/keepalive link) once
# the intended port is confirmed.
bin/cani alpha add ip 192.168.255.0/31 --interface "FORGE-3507u44L:1/1/48" --status Active --description "VSX keepalive"
bin/cani alpha add ip 192.168.255.1/31 --interface "FORGE-3507u43L:1/1/48" --status Active --description "VSX keepalive"
bin/cani alpha add ip 192.168.255.2/31 --interface "FORGE-3701u46L:1/1/48" --status Active --description "VSX keepalive"
bin/cani alpha add ip 192.168.255.3/31 --interface "FORGE-3701u45L:1/1/48" --status Active --description "VSX keepalive"

# VLAN interface IPs on leaves (x3507)
bin/cani alpha add ip 172.17.2.2/23 --interface "FORGE-3507u44L:vlan3" --status Active --description "Harvester"
bin/cani alpha add ip 172.17.2.3/23 --interface "FORGE-3507u43L:vlan3" --status Active --description "Harvester"
bin/cani alpha add ip 10.102.13.2/24 --interface "FORGE-3507u44L:vlan9" --status Active --description "Floating IPs"
bin/cani alpha add ip 10.102.13.3/24 --interface "FORGE-3507u43L:vlan9" --status Active --description "Floating IPs"
bin/cani alpha add ip 192.168.64.2/31 --interface "FORGE-3507u44L:vlan10" --status Active --description "Transit VLAN"
bin/cani alpha add ip 192.168.64.3/31 --interface "FORGE-3507u43L:vlan10" --status Active --description "Transit VLAN"
bin/cani alpha add ip 10.102.12.2/24 --interface "FORGE-3507u44L:vlan2000" --status Active --description "Legacy CSM CMN"
bin/cani alpha add ip 10.102.12.3/24 --interface "FORGE-3507u43L:vlan2000" --status Active --description "Legacy CSM CMN"

# VLAN interface IPs on leaves (x3701)
bin/cani alpha add ip 172.17.2.4/23 --interface "FORGE-3701u46L:vlan3" --status Active --description "Harvester"
bin/cani alpha add ip 172.17.2.5/23 --interface "FORGE-3701u45L:vlan3" --status Active --description "Harvester"
bin/cani alpha add ip 10.102.13.4/24 --interface "FORGE-3701u46L:vlan9" --status Active --description "Floating IPs"
bin/cani alpha add ip 10.102.13.5/24 --interface "FORGE-3701u45L:vlan9" --status Active --description "Floating IPs"
bin/cani alpha add ip 192.168.64.4/31 --interface "FORGE-3701u46L:vlan10" --status Active --description "Transit VLAN"
bin/cani alpha add ip 192.168.64.5/31 --interface "FORGE-3701u45L:vlan10" --status Active --description "Transit VLAN"
bin/cani alpha add ip 10.102.12.4/24 --interface "FORGE-3701u46L:vlan2000" --status Active --description "Legacy CSM CMN"
bin/cani alpha add ip 10.102.12.5/24 --interface "FORGE-3701u45L:vlan2000" --status Active --description "Legacy CSM CMN"

# Upstream LEGACY VRF IPs
bin/cani alpha add ip 10.102.30.50/30 --interface "FORGE-3507u44L:1/1/25" --status Active --description "CFCANB1S1-1/1/31"
bin/cani alpha add ip 10.102.30.54/30 --interface "FORGE-3507u43L:1/1/25" --status Active --description "CFCANB1S1-1/1/32"

# Switch management IPs
bin/cani alpha add ip 10.101.49.93/22 --interface "FORGE-3507u46S:mgmt" --status Active --description "Management Interface"
bin/cani alpha add ip 10.101.49.94/22 --interface "FORGE-3507u45S:mgmt" --status Active --description "Management Interface"
bin/cani alpha add ip 10.101.49.90/22 --interface "FORGE-3507u44L:mgmt" --status Active --description "Management Interface"
bin/cani alpha add ip 10.101.49.91/22 --interface "FORGE-3507u43L:mgmt" --status Active --description "Management Interface"
bin/cani alpha add ip 10.101.49.95/22 --interface "FORGE-3701u46L:mgmt" --status Active --description "Management Interface"
bin/cani alpha add ip 10.101.49.96/22 --interface "FORGE-3701u45L:mgmt" --status Active --description "Management Interface"


# ═══════════════════════════════════════════════════════════════════════════
# INTERFACE MAC ADDRESSES: Assign hardware MACs 
# ═══════════════════════════════════════════════════════════════════════════
# x3701: XD670 GPU node iLOs
bin/cani alpha update interface --device "GH-3701u34" --name "iLO" --mac "10:ff:e0:30:5c:76"
bin/cani alpha update interface --device "GH-3701u26" --name "iLO" --mac "10:ff:e0:37:21:66"
bin/cani alpha update interface --device "GH-3701u18" --name "iLO" --mac "10:ff:e0:37:21:e6"
bin/cani alpha update interface --device "GH-3701u10" --name "iLO" --mac "10:ff:e0:35:b8:b2"
# x3507: DL380 node iLO / Gig-E 1 / HSN 0 / HSN 1
bin/cani alpha update interface --device "SERV-3507u5" --name "iLO" --mac "5c:ed:8c:ed:fd:74"
bin/cani alpha update interface --device "SERV-3507u5" --name "Gig-E 1" --mac "04:32:01:5c:8a:34"
bin/cani alpha update interface --device "SERV-3507u5" --name "HSN 0" --mac "88:e9:a4:a7:19:d8"
bin/cani alpha update interface --device "SERV-3507u5" --name "HSN 1" --mac "88:e9:a4:a7:19:d9"
bin/cani alpha update interface --device "SERV-3507u7" --name "iLO" --mac "5c:ed:8c:ed:fa:5a"
bin/cani alpha update interface --device "SERV-3507u7" --name "Gig-E 1" --mac "04:32:01:5b:d2:c8"
bin/cani alpha update interface --device "SERV-3507u7" --name "HSN 0" --mac "88:e9:a4:a7:99:58"
bin/cani alpha update interface --device "SERV-3507u7" --name "HSN 1" --mac "88:e9:a4:a7:99:59"
bin/cani alpha update interface --device "SERV-3507u9" --name "iLO" --mac "5c:ed:8c:ed:fa:30"
bin/cani alpha update interface --device "SERV-3507u9" --name "Gig-E 1" --mac "04:32:01:5b:bd:8c"
bin/cani alpha update interface --device "SERV-3507u9" --name "HSN 0" --mac "88:e9:a4:a7:29:18"
bin/cani alpha update interface --device "SERV-3507u9" --name "HSN 1" --mac "88:e9:a4:a7:29:19"
bin/cani alpha update interface --device "SERV-3507u11" --name "iLO" --mac "5c:ed:8c:ed:fd:5e"
bin/cani alpha update interface --device "SERV-3507u11" --name "Gig-E 1" --mac "04:32:01:5c:92:7a"
bin/cani alpha update interface --device "SERV-3507u11" --name "HSN 0" --mac "88:e9:a4:a7:29:10"
bin/cani alpha update interface --device "SERV-3507u11" --name "HSN 1" --mac "88:e9:a4:a7:29:11"
bin/cani alpha update interface --device "SERV-3507u13" --name "iLO" --mac "5c:ed:8c:ed:fa:5e"
bin/cani alpha update interface --device "SERV-3507u13" --name "Gig-E 1" --mac "04:32:01:5b:ea:80"
bin/cani alpha update interface --device "SERV-3507u13" --name "HSN 0" --mac "88:e9:a4:a7:99:70"
bin/cani alpha update interface --device "SERV-3507u13" --name "HSN 1" --mac "88:e9:a4:a7:99:71"
bin/cani alpha update interface --device "SERV-3507u15" --name "iLO" --mac "5c:ed:8c:ed:fa:44"
bin/cani alpha update interface --device "SERV-3507u15" --name "Gig-E 1" --mac "04:32:01:5c:8a:76"
bin/cani alpha update interface --device "SERV-3507u15" --name "HSN 0" --mac "88:e9:a4:a7:99:80"
bin/cani alpha update interface --device "SERV-3507u15" --name "HSN 1" --mac "88:e9:a4:a7:99:81"
bin/cani alpha update interface --device "SERV-3507u17" --name "iLO" --mac "5c:ed:8c:ed:fb:28"
bin/cani alpha update interface --device "SERV-3507u17" --name "Gig-E 1" --mac "04:32:01:5b:f8:de"
bin/cani alpha update interface --device "SERV-3507u17" --name "HSN 0" --mac "88:e9:a4:a7:99:68"
bin/cani alpha update interface --device "SERV-3507u17" --name "HSN 1" --mac "88:e9:a4:a7:99:69"
bin/cani alpha update interface --device "SERV-3507u19" --name "iLO" --mac "5c:ed:8c:ed:fa:1c"
bin/cani alpha update interface --device "SERV-3507u19" --name "Gig-E 1" --mac "04:32:01:5b:f2:a2"
bin/cani alpha update interface --device "SERV-3507u19" --name "HSN 0" --mac "88:e9:a4:a7:69:a8"
bin/cani alpha update interface --device "SERV-3507u19" --name "HSN 1" --mac "88:e9:a4:a7:69:a9"
bin/cani alpha update interface --device "SERV-3507u21" --name "iLO" --mac "5c:ed:8c:ed:fa:60"
bin/cani alpha update interface --device "SERV-3507u21" --name "Gig-E 1" --mac "04:32:01:5b:f9:8c"
bin/cani alpha update interface --device "SERV-3507u21" --name "HSN 0" --mac "88:e9:a4:a7:99:78"
bin/cani alpha update interface --device "SERV-3507u21" --name "HSN 1" --mac "88:e9:a4:a7:99:79"

# x3507: DL380 400G NDR InfiniBand (ConnectX-7) HSN 0. The CX7 port shares the
# name "HSN 0" with the node's CX6 and chassis HSN ports, so target the module
# directly with --module to land the 400G MAC on the CX7 alone.
bin/cani alpha update interface --module "CX7-SERV-3507u5" --name "HSN 0" --mac "94:6d:ae:be:97:36"
bin/cani alpha update interface --module "CX7-SERV-3507u7" --name "HSN 0" --mac "94:6d:ae:be:9a:16"
bin/cani alpha update interface --module "CX7-SERV-3507u9" --name "HSN 0" --mac "94:6d:ae:be:97:4a"
bin/cani alpha update interface --module "CX7-SERV-3507u11" --name "HSN 0" --mac "94:6d:ae:be:96:b6"
bin/cani alpha update interface --module "CX7-SERV-3507u13" --name "HSN 0" --mac "94:6d:ae:be:99:fe"
bin/cani alpha update interface --module "CX7-SERV-3507u15" --name "HSN 0" --mac "94:6d:ae:be:97:4c"
bin/cani alpha update interface --module "CX7-SERV-3507u17" --name "HSN 0" --mac "94:6d:ae:be:99:14"
bin/cani alpha update interface --module "CX7-SERV-3507u19" --name "HSN 0" --mac "94:6d:ae:be:97:3a"
bin/cani alpha update interface --module "CX7-SERV-3507u21" --name "HSN 0" --mac "94:6d:ae:be:99:c4"


# ═══════════════════════════════════════════════════════════════════════════
# CONNECTIONS: Import all cable connections from YAML connection map
# ═══════════════════════════════════════════════════════════════════════════
bin/cani alpha add connections gnv2-connections.yml
