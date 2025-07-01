/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package hpcm

// Discover represents an entry in the [discover] section in the hpcm.config
// The fields here can vary widly and are sometimes duplicated with those in Templates/Attributes
// TODO: de-dupe these fields into re-useable structs
type Discover struct {
	Hostname1                string   `json:"hostname1" yaml:"hostname1,omitempty"`
	InternalName             string   `json:"internal_name" yaml:"internal_name,omitempty"`
	TemplateName             string   `json:"template_name,omitempty" yaml:"template_name,omitempty"`
	MgmtBmcNetName           string   `json:"mgmt_bmc_net_name,omitempty" yaml:"mgmt_bmc_net_name,omitempty"`
	MgmtBmcNetMacs           []string `json:"mgmt_bmc_net_macs,omitempty" yaml:"mgmt_bmc_net_macs,omitempty"`
	MgmtBmcNetIp             []string `json:"mgmt_bmc_net_ip,omitempty" yaml:"mgmt_bmc_net_ip,omitempty"`
	MgmtBmcNetIfIp           []string `json:"mgmt_bmc_net_if_ip,omitempty" yaml:"mgmt_bmc_net_if_ip,omitempty"`
	MgmtBmcNetIf             bool     `json:"mgmt_bmc_net_if,omitempty" yaml:"mgmt_bmc_net_if,omitempty"`
	MgmtNetName              string   `json:"mgmt_net_name,omitempty" yaml:"mgmt_net_name,omitempty"`
	MgmtNetBondingMaster     string   `json:"mgmt_net_bonding_master,omitempty" yaml:"mgmt_net_bonding_master,omitempty"`
	MgmtNetBodingMode        string   `json:"mgmt_net_bonding_mode,omitempty" yaml:"mgmt_net_bonding_mode,omitempty"`
	MgmtNetMacs              []string `json:"mgmt_net_macs,omitempty" yaml:"mgmt_net_macs,omitempty"`
	MgmtNetInterfaces        []string `json:"mgmt_net_interfaces,omitempty" yaml:"mgmt_net_interfaces,omitempty"`
	MgmtNetInterfaceName     string   `json:"mgmt_net_interface_name,omitempty" yaml:"mgmt_net_interface_name,omitempty"`
	MgmtNetIp                string   `json:"mgmt_net_ip,omitempty" yaml:"mgmt_net_ip,omitempty"`
	Data1NetName             string   `json:"data1_net_name,omitempty" yaml:"data1_net_name,omitempty"`
	Data1NetInterfaces       []string `json:"data1_net_interfaces,omitempty" yaml:"data1_net_interfaces,omitempty"`
	Data1NetInterfaceName    []string `json:"data1_net_interface_name,omitempty" yaml:"data1_net_interface_name,omitempty"`
	Data1NetMacs             []string `json:"data1_net_macs,omitempty" yaml:"data1_net_macs,omitempty"`
	Data1NetIp               []string `json:"data1_net_ip,omitempty" yaml:"data1_net_ip,omitempty"`
	NetworkGroup             string   `json:"network_group,omitempty" yaml:"network_group,omitempty"`
	RootFs                   string   `json:"rootfs,omitempty" yaml:"rootfs,omitempty"`
	NfsWritableType          string   `json:"nfs_writable_type,omitempty" yaml:"nfs_writable_type,omitempty"`
	Transport                string   `json:"transport,omitempty" yaml:"transport,omitempty"`
	AliasGroups              []string `json:"alias_group,omitempty" yaml:"alias_group,omitempty"`
	ConserverLogging         bool     `json:"conserver_logging,omitempty" yaml:"conserver_logging,omitempty"`
	ConserverOnDemand        bool     `json:"conserver_on_demand,omitempty" yaml:"conserver_on_demand,omitempty"`
	DhcpBootfile             string   `json:"dhcp_bootfile,omitempty" yaml:"dhcp_bootfile,omitempty"`
	DiskBootloader           bool     `json:"disk_bootloader,omitempty" yaml:"disk_bootloader,omitempty"`
	PredicatableNetNames     bool     `json:"predictable_net_names,omitempty" yaml:"predictable_net_names,omitempty"`
	RedundantMgmtNetwork     bool     `json:"redundant_mgmt_network,omitempty" yaml:"redundant_mgmt_network,omitempty"`
	SuLeader                 []string `json:"su_leader,omitempty" yaml:"su_leader,omitempty"`
	SwitchMgmtNetwork        bool     `json:"switch_mgmt_network,omitempty" yaml:"switch_mgmt_network,omitempty"`
	TpmBoot                  bool     `json:"tpm_boot,omitempty" yaml:"tpm_boot,omitempty"`
	ConsoleDevice            string   `json:"console_device,omitempty" yaml:"console_device,omitempty"`
	Architecture             string   `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	CardType                 string   `json:"card_type,omitempty" yaml:"card_type,omitempty"`
	Image                    string   `json:"image,omitempty" yaml:"image,omitempty"`
	Kernel                   string   `json:"kernel,omitempty" yaml:"kernel,omitempty"`
	BaudRate                 int      `json:"baud_rate,omitempty" yaml:"baud_rate,omitempty"`
	BmcUsername              string   `json:"bmc_username,omitempty" yaml:"bmc_username,omitempty"`
	BmcPassword              string   `json:"-" yaml:"-"`
	RackNr                   int      `json:"rack_nr,omitempty" yaml:"rack_nr,omitempty"`
	Chassis                  int      `json:"chassis,omitempty" yaml:"chassis,omitempty"`
	NodeNr                   int      `json:"node_nr,omitempty" yaml:"node_nr,omitempty"`
	Tray                     int      `json:"tray,omitempty" yaml:"tray,omitempty"`
	ControllerNr             int      `json:"controller_nr,omitempty" yaml:"controller_nr,omitempty"`
	CmcInventoryManaged      bool     `json:"cmc_inventory_managed,omitempty" yaml:"cmc_inventory_managed,omitempty"`
	NodeController           string   `json:"node_controller,omitempty" yaml:"node_controller,omitempty"`
	Password                 string   `json:"-" yaml:"-"`
	Username                 string   `json:"username,omitempty" yaml:"username,omitempty"`
	AdminHouseInterface      string   `json:"admin_house_interface,omitempty" yaml:"admin_house_interface,omitempty"`
	ExtraRoutes              bool     `json:"extra_routes,omitempty" yaml:"extra_routes,omitempty"`
	CmmParent                string   `json:"cmm_parent,omitempty" yaml:"cmm_parent,omitempty"`
	Ice                      string   `json:"ice,omitempty" yaml:"ice,omitempty"`
	Net                      string   `json:"net,omitempty" yaml:"net,omitempty"`
	Type                     string   `json:"type,omitempty" yaml:"type,omitempty"`
	MgmtSwPartner            string   `json:"mgmtsw_partner" yaml:"mgmtsw_partner,omitempty"`
	MgmtswIsls               string   `json:"mgmtsw_isls,omitempty" yaml:"mgmtsw_isls,omitempty"`
	DiscoverSkipSwitchconfig string   `json:"discover_skip_switchconfig,omitempty" yaml:"discover_skip_switchconfig,omitempty"`
	PduProtocol              string   `json:"pdu_protocol,omitempty" yaml:"pdu_protocol,omitempty"`
}

// Template represents an entry in the [template] section in the hpcm.config
// The fields here can vary widly and are sometimes duplicated with those in Discover/Attributes
// TODO: de-dupe these fields into re-useable structs
type Template struct {
	Name                 string   `json:"name,omitempty" yaml:"name,omitempty"`
	MgmtBmcNetName       string   `json:"mgmt_bmc_net_name,omitempty" yaml:"mgmt_bmc_net_name,omitempty"`
	MgmtNetName          string   `json:"mgmt_net_name,omitempty" yaml:"mgmt_net_name,omitempty"`
	RedundantMgmtNetwork bool     `json:"redundant_mgmt_network,omitempty" yaml:"redundant_mgmt_network,omitempty"`
	SwitchMgmtNetwork    bool     `json:"switch_mgmt_network,omitempty" yaml:"switch_mgmt_network,omitempty"`
	DhcpBootfile         string   `json:"dhcp_bootfile,omitempty" yaml:"dhcp_bootfile,omitempty"`
	ForceDisk            string   `json:"force_disk" yaml:"force_disk,omitempty"`
	ConserverLogging     bool     `json:"conserver_logging,omitempty" yaml:"conserver_logging,omitempty"`
	ConserverOnDemand    bool     `json:"conserver_on_demand,omitempty" yaml:"conserver_on_demand,omitempty"`
	RootFs               string   `json:"rootfs,omitempty" yaml:"rootfs,omitempty"`
	ConsoleDevice        string   `json:"console_device,omitempty" yaml:"console_device,omitempty"`
	TpmBoot              bool     `json:"tpm_boot,omitempty" yaml:"tpm_boot,omitempty"`
	MgmtSwPartner        string   `json:"mgmtsw_partner,omitempty" yaml:"mgmtsw_partner,omitempty"`
	MgmtSw               string   `json:"mgmtsw,omitempty" yaml:"mgmtsw,omitempty"`
	PredicatableNetNames bool     `json:"predictable_net_names,omitempty" yaml:"predictable_net_names,omitempty"`
	Transport            string   `json:"transport,omitempty" yaml:"transport,omitempty"`
	BaudRate             int      `json:"baud_rate,omitempty" yaml:"baud_rate,omitempty"`
	BmcUsername          string   `json:"bmc_username,omitempty" yaml:"bmc_username,omitempty"`
	BmcPassword          string   `json:"-" yaml:"-"`
	MgmtNetInterfaces    []string `json:"mgmt_net_interfaces,omitempty" yaml:"mgmt_net_interfaces,omitempty"`
	MgmtNetBondingMaster string   `json:"mgmt_net_bonding_master,omitempty" yaml:"mgmt_net_bonding_master,omitempty"`
	DiskBootloader       bool     `json:"disk_bootloader,omitempty" yaml:"disk_bootloader,omitempty"`
	Architecture         string   `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	CardType             string   `json:"card_type,omitempty" yaml:"card_type,omitempty"`
	MgmtNetBodingMode    string   `json:"mgmt_net_bonding_mode,omitempty" yaml:"mgmt_net_bonding_mode,omitempty"`
	SuLeaderRole         string   `json:"su_leader_role,omitempty" yaml:"su_leader_role,omitempty"`
	Image                string   `json:"image,omitempty" yaml:"image,omitempty"`
	CtrlModel            string   `json:"ctrl_model,omitempty" yaml:"ctrl_model,omitempty"`
	Password             string   `json:"-" yaml:"-"`
	Username             string   `json:"username,omitempty" yaml:"username,omitempty"`
	Data1NetName         string   `json:"data1_net_name,omitempty" yaml:"data1_net_name,omitempty"`
	Data1NetInterfaces   []string `json:"data1_net_interfaces,omitempty" yaml:"data1_net_interfaces,omitempty"`
	Data2NetName         string   `json:"data2_net_name,omitempty" yaml:"data2_net_name,omitempty"`
	Data2NetInterfaces   []string `json:"data2_net_interfaces,omitempty" yaml:"data2_net_interfaces,omitempty"`
	NfsWritableType      string   `json:"nfs_writable_type,omitempty" yaml:"nfs_writable_type,omitempty"`
	SuLeader             []string `json:"su_leader,omitempty" yaml:"su_leader,omitempty"`
	MgmtswIsls           string   `json:"mgmtsw_isls,omitempty" yaml:"mgmtsw_isls,omitempty"`
	DestroyDiskLabel     bool     `json:"destroy_disk_label,omitempty" yaml:"destroy_disk_label,omitempty"`
	MdMetadata           string   `json:"md_metadata,omitempty" yaml:"md_metadata,omitempty"`
}

// NicTemplate represents an entry in the [network] section in the hpcm.config
type NicTemplate struct {
	Template      string   `json:"template,omitempty" yaml:"template,omitempty"`
	Network       string   `json:"network,omitempty" yaml:"network,omitempty"`
	BondingMaster string   `json:"bonding_master,omitempty" yaml:"bonding_master,omitempty"`
	BondingMode   string   `json:"bonding_mode,omitempty" yaml:"bonding_mode,omitempty"`
	NetIfs        []string `json:"net_ifs,omitempty" yaml:"net_ifs,omitempty"`
	BrNane        string   `json:"br_name,omitempty" yaml:"br_name,omitempty"`
}

// Dns represents the [dns] section in the hpcm.config
type Dns struct {
	ClusterDomain string `json:"cluster_domain,omitempty" yaml:"cluster_domain,omitempty"`
	Nameserver1   string `json:"nameserver1,omitempty" yaml:"nameserver1,omitempty"`
	Nameserver2   string `json:"nameserver2,omitempty" yaml:"nameserver2,omitempty"`
}

// Attributes represents the [attributes] section in the hpcm.config
// The fields here can vary widly and are sometimes duplicated with those in Discover/Templates
// TODO: de-dupe these fields into re-useable structs
type Attributes struct {
	AdminHouseInterface            string   `json:"admin_house_interface,omitempty" yaml:"admin_house_interface,omitempty"`
	AdminManagementInterfaces      []string `json:"admin_mgmt_interfaces,omitempty" yaml:"admin_mgmt_interfaces,omitempty"`
	AdminManagementBmcInterfaces   []string `json:"admin_mgmt_bmc_interfaces,omitempty" yaml:"admin_mgmt_bmc_interfaces,omitempty"`
	AdminUdpcastTtl                int      `json:"admin_udpcast_ttl,omitempty" yaml:"admin_udpcast_ttl,omitempty"`
	AdminUdpcastMcaseRdvAddr       string   `json:"admin_udpcast_mcast_rdv_addr,omitempty" yaml:"admin_udpcast_mcast_rdv_addr,omitempty"`
	AdminMgmtBondingMode           string   `json:"admin_mgmt_bonding_mode,omitempty" yaml:"admin_mgmt_bonding_mode,omitempty"`
	BladeMondScanInterval          string   `json:"blademond_scan_interval,omitempty" yaml:"blademond_scan_interval,omitempty"`
	CmcsPerMgmtVlan                int      `json:"cmcs_per_mgmt_vlan" yaml:"cmcs_per_mgmt_vlan,omitempty"`
	CmcsPerRack                    int      `json:"cmcs_per_rack" yaml:"cmcs_per_rack,omitempty"`
	CmmsPerRack                    int      `json:"cmms_per_rack" yaml:"cmms_per_rack,omitempty"`
	ConserverLogging               bool     `json:"conserver_logging,omitempty" yaml:"conserver_logging,omitempty"`
	ConserverOnDemand              bool     `json:"conserver_on_demand,omitempty" yaml:"conserver_on_demand,omitempty"`
	CopyAdminSshConfig             bool     `json:"copy_admin_ssh_config,omitempty" yaml:"copy_admin_ssh_config,omitempty"`
	DhcpBootfile                   string   `json:"dhcp_bootfile,omitempty" yaml:"dhcp_bootfile,omitempty"`
	DiscoverSkipSwitchconfig       bool     `json:"discover_skip_switchconfig,omitempty" yaml:"discover_skip_switchconfig,omitempty"`
	DomainSearchPath               []string `json:"domain_search_path,omitempty" yaml:"domain_search_path,omitempty"`
	HeadVlan                       int      `json:"head_vlan,omitempty" yaml:"head_vlan,omitempty"`
	Ipv6LocalSiteUla               string   `json:"ipv6_local_site_ula,omitempty" yaml:"ipv6_local_site_ula,omitempty"`
	MacRackIrus                    int      `json:"max_rack_irus,omitempty" yaml:"max_rack_irus,omitempty"`
	McellNetwork                   bool     `json:"mcell_network,omitempty" yaml:"mcell_network,omitempty"`
	McellVlan                      int      `json:"mcell_vlan,omitempty" yaml:"mcell_vlan,omitempty"`
	MgmtCtrlVlanEnd                int      `json:"mgmt_ctrl_vlan_end,omitempty" yaml:"mgmt_ctrl_vlan_end,omitempty"`
	MgmtCtrlVlanStart              int      `json:"mgmt_ctrl_vlan_start,omitempty" yaml:"mgmt_ctrl_vlan_start,omitempty"`
	MgmtNetRoutingProtocol         string   `json:"mgmt_net_routing_protocol,omitempty" yaml:"mgmt_net_routing_protocol,omitempty"`
	MgmtNetSubetSelection          string   `json:"mgmt_net_subnet_selection,omitempty" yaml:"mgmt_net_subnet_selection,omitempty"`
	MgmtVlanEnd                    int      `json:"mgmt_vlan_end,omitempty" yaml:"mgmt_vlan_end,omitempty"`
	MgmtVlanStart                  int      `json:"mgmt_vlan_start,omitempty" yaml:"mgmt_vlan_start,omitempty"`
	PredicatableNetNames           bool     `json:"predictable_net_names,omitempty" yaml:"predictable_net_names,omitempty"`
	RackStartNumber                int      `json:"rack_start_number,omitempty" yaml:"rack_start_number,omitempty"`
	RackVlanEnd                    int      `json:"rack_vlan_end,omitempty" yaml:"rack_vlan_end,omitempty"`
	RackVlanStart                  int      `json:"rack_vlan_start,omitempty" yaml:"rack_vlan_start,omitempty"`
	RedundantMgmtNetwork           bool     `json:"redundant_mgmt_network,omitempty" yaml:"redundant_mgmt_network,omitempty"`
	SwitchMgmtNetwork              bool     `json:"switch_mgmt_network,omitempty" yaml:"switch_mgmt_network,omitempty"`
	UdpcastMaxBitrate              string   `json:"udpcast_max_bitrate,omitempty" yaml:"udpcast_max_bitrate,omitempty"`
	UdpcastMaxWait                 int      `json:"udpcast_max_wait,omitempty" yaml:"udpcast_max_wait,omitempty"`
	UdpcastMcastRdvAddr            string   `json:"udpcast_mcast_rdv_addr,omitempty" yaml:"udpcast_mcast_rdv_addr,omitempty"`
	UdpcastMinRecievers            int      `json:"udpcast_min_receivers,omitempty" yaml:"udpcast_min_receivers,omitempty"`
	UdpcastMinWait                 int      `json:"udpcast_min_wait,omitempty" yaml:"udpcast_min_wait,omitempty"`
	UdpcastRexmitHelloInterval     int      `json:"udpcast_rexmit_hello_interval,omitempty" yaml:"udpcast_rexmit_hello_interval,omitempty"`
	MonitoringKafkaElkAlertEnabled bool     `json:"monitoring_kafka_elk_alerta_enabled,omitempty" yaml:"monitoring_kafka_elk_alerta_enabled,omitempty"`
	MonitoringNativeEnabled        bool     `json:"monitoring_native_enabled,omitempty" yaml:"monitoring_native_enabled,omitempty"`
	ConserverTimestamp             bool     `json:"conserver_timestamp,omitempty" yaml:"conserver_timestamp,omitempty"`
	DhcpdDeafultLeaseTime          int      `json:"dhcpd_default_lease_time,omitempty" yaml:"dhcpd_default_lease_time,omitempty"`
	DhcpdMaxLeaseTime              int      `json:"dhcpd_max_lease_time,omitempty" yaml:"dhcpd_max_lease_time,omitempty"`
	MySqlReplication               bool     `json:"my_sql_replication,omitempty" yaml:"my_sql_replication,omitempty"`
	MonitoringGangliaEnabled       bool     `json:"monitoring_ganglia_enabled,omitempty" yaml:"monitoring_ganglia_enabled,omitempty"`
	MonitoringNagiosEnabled        bool     `json:"monitoring_nagios_enabled,omitempty" yaml:"monitoring_nagios_enabled,omitempty"`
	MgmtNetAliasSelection          string   `json:"mgmt_net_alias_selection,omitempty" yaml:"mgmt_net_alias_selection,omitempty"`
}

// Network represents an entry in the [network] section in the hpcm.config
type Network struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	Subnet      string `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	Netmask     string `json:"netmask,omitempty" yaml:"netmask,omitempty"`
	RackNetmask string `json:"rack_netmask,omitempty" yaml:"rack_netmask,omitempty"`
	Gateway     string `json:"gateway,omitempty" yaml:"gateway,omitempty"`
}

// Images represents the [images] section in the hpcm.config
type Images struct {
	ImageTypes []string `json:"image_types,omitempty" yaml:"image_types"`
}

type HpcmConfig struct {
	Templates    map[string]Template    `json:"templates,omitempty" yaml:"templates,omitempty"`
	NicTemplates map[string]NicTemplate `json:"nic_templates,omitempty" yaml:"nic_templates,omitempty"`
	Discover     map[string]Discover    `json:"discover,omitempty" yaml:"discover,omitempty"`
	Dns          map[string]Dns         `json:"dns,omitempty" yaml:"dns,omitempty"`
	Attributes   Attributes             `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Networks     map[string]Network     `json:"networks,omitempty" yaml:"networks,omitempty"`
	Images       []Images               `json:"images,omitempty" yaml:"images,omitempty" toml:"images,omitempty"`
}
