/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

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

// LoadCmConfig loads a HPCM cluster definition config file
// This file is normally parsed in a custom way using Perl ParseWords
// Since the format is close to that if INI and to prevent using a smaller third-party lib,
// The file is loaded with the google ini module, which can get most of the way there
// This also allows easy export to INI later, in an effort drive away from the Perl stuff
func LoadCmConfig(path string) (hpcmConfig HpcmConfig, err error) {
	// ShadowLoad since each line contains more than one key/val pair
	cfg, err := ini.ShadowLoad(path)
	if err != nil {
		return hpcmConfig, err
	}

	// load each section
	templates, err := importCfgTemplatesSection(cfg)
	if err != nil {
		return hpcmConfig, err
	}

	discover, err := importCfgDiscoverSection(cfg)
	if err != nil {
		return hpcmConfig, err
	}

	nic_templates, err := importCfgNicTemplatesSection(cfg)
	if err != nil {
		return hpcmConfig, err
	}

	attributes, err := importCfgAttributesSection(cfg)
	if err != nil {
		return hpcmConfig, err
	}

	dns, err := importCfgDnsSection(cfg)
	if err != nil {
		return hpcmConfig, err
	}

	networks, err := importCfgNetworksSection(cfg)
	if err != nil {
		return hpcmConfig, err
	}

	images, err := importCfgImagesSection(cfg)
	if err != nil {
		return hpcmConfig, err
	}

	// add each section to the config object
	hpcmConfig = HpcmConfig{
		Templates:    templates,
		NicTemplates: nic_templates,
		Discover:     discover,
		Attributes:   attributes,
		Dns:          dns,
		Networks:     networks,
		Images:       images,
	}

	return hpcmConfig, nil
}

// importCfgDiscoverSection parses the [discover] section of an hpcm file and translates it to a Discover object
func importCfgDiscoverSection(cfg *ini.File) (map[string]Discover, error) {
	discover := map[string]Discover{}
	for _, section := range cfg.Sections() {
		if section.Name() == "discover" {
			// various vintages have different key names
			secName := ""
			if section.HasKey("hostname1") {
				secName = "hostname1"
			}
			if section.HasKey("internal_name") {
				secName = "internal_name"
			}
			if section.HasKey("alias1") {
				secName = "alias1"
			}
			if section.HasKey("temponame") {
				secName = "temponame"
			}

			d := Discover{}
			for _, v := range section.Key(secName).ValueWithShadows() {
				subkeys := strings.Split(v, ", ")
				for _, subkey := range subkeys {
					kvp := strings.Split(subkey, "=")
					if len(kvp) == 2 {
						sk, sv := strings.TrimSpace(kvp[0]), kvp[1]
						sv = strings.Trim(sv, `"`)
						switch sk {
						case "internal_name":
							d.InternalName = sv
						case "template_name":
							d.TemplateName = sv
						case "mgmt_bmc_net_name":
							d.MgmtBmcNetName = sv
						case "mgmt_bmc_net_macs":
							// simple sanitize and append macs to the list
							for _, m := range strings.Split(sv, ",") {
								mac, err := net.ParseMAC(strings.Trim(m, `"`))
								if err != nil {
									return discover, err
								}
								if !ContainS(d.MgmtBmcNetMacs, mac.String()) {
									d.MgmtBmcNetMacs = append(d.MgmtBmcNetMacs, mac.String())
								}
							}
						case "mgmt_net_bonding_mode":
							d.MgmtNetBodingMode = sv
						case "mgmt_net_macs":
							// simple sanitize and append macs to the list
							for _, m := range strings.Split(sv, ",") {
								mac, err := net.ParseMAC(strings.Trim(m, `"`))
								if err != nil {
									return discover, err
								}
								if !ContainS(d.MgmtNetMacs, mac.String()) {
									d.MgmtNetMacs = append(d.MgmtNetMacs, mac.String())
								}
							}
						case "mgmt_net_interfaces":
							for _, i := range strings.Split(sv, ",") {
								iface := strings.Trim(i, `"`)
								if !ContainS(d.MgmtNetInterfaces, iface) {
									d.MgmtNetInterfaces = append(d.MgmtNetInterfaces, iface)
								}
							}
						case "mgmt_net_interface_name":
							d.MgmtNetInterfaceName = sv
						case "mgmt_net_ip":
							ip := net.ParseIP(sv)
							d.MgmtNetIp = ip.String()
						case "data1_net_name":
							d.Data1NetName = sv
						case "data1_net_interfaces":
							for _, i := range strings.Split(sv, ",") {
								iface := strings.Trim(i, `"`)
								if !ContainS(d.Data1NetInterfaces, iface) {
									d.Data1NetInterfaces = append(d.Data1NetInterfaces, iface)
								}
							}
						case "data1_net_interface_name":
							for _, i := range strings.Split(sv, ",") {
								iface := strings.Trim(i, `"`)
								if !ContainS(d.Data1NetInterfaceName, iface) {
									d.Data1NetInterfaceName = append(d.Data1NetInterfaceName, iface)
								}
							}
						case "data1_net_ip":
							for _, i := range strings.Split(sv, ",") {
								ip := net.ParseIP(i)
								if !ContainS(d.Data1NetIp, ip.String()) {
									d.Data1NetIp = append(d.Data1NetIp, ip.String())
								}
							}
						case "network_group":
							d.NetworkGroup = sv
						case "rack_nr":
							n, err := strconv.Atoi(sv)
							if err != nil {
								return discover, err
							}
							d.RackNr = n
						case "chassis":
							n, err := strconv.Atoi(sv)
							if err != nil {
								return discover, err
							}
							d.Chassis = n
						case "node_nr":
							n, err := strconv.Atoi(sv)
							if err != nil {
								return discover, err
							}
							d.NodeNr = n
						case "tray":
							n, err := strconv.Atoi(sv)
							if err != nil {
								return discover, err
							}
							d.Tray = n
						case "controller_nr":
							n, err := strconv.Atoi(sv)
							if err != nil {
								return discover, err
							}
							d.ControllerNr = n
						case "rootfs":
							d.RootFs = sv
						case "nfs_writable_type":
							d.NfsWritableType = sv
						case "transport":
							d.Transport = sv
						case "alias_groups":
							// TODO:
						case "cmcinventory_managed":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.CmcInventoryManaged = boolVal
						case "conserver_logging":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.ConserverLogging = boolVal
						case "conserver_ondemand":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.ConserverOnDemand = boolVal
						case "dhcp_bootfile":
							d.DhcpBootfile = sv
						case "disk_bootloader":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.DiskBootloader = boolVal
						case "predictable_net_names":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.PredicatableNetNames = boolVal
						case "redundant_mgmt_network":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.RedundantMgmtNetwork = boolVal
						case "su_leader":
							for _, i := range strings.Split(sv, ",") {
								ip := net.ParseIP(i)
								if !ContainS(d.SuLeader, ip.String()) {
									d.SuLeader = append(d.SuLeader, ip.String())
								}
							}
						case "switch_mgmt_network":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.SwitchMgmtNetwork = boolVal
						case "tpm_boot":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.TpmBoot = boolVal
						case "console_device":
							d.ConsoleDevice = sv
						case "architecture":
							d.Architecture = sv
						case "card_type":
							d.CardType = sv
						case "image":
							d.Image = sv
						case "kernel":
							d.Kernel = sv
						case "baud_rate":
							n, err := strconv.Atoi(sv)
							if err != nil {
								n = 115200
								// return discover, err
							}
							d.BaudRate = n
						case "node_controller":
							d.NodeController = sv
						case "bmc_username":
							d.BmcUsername = sv
						case "bmc_password":
							d.BmcPassword = sv
						case "password":
							d.Password = sv
						case "username":
							d.Username = sv
						case "mgmt_net_name":
							d.MgmtNetName = sv
						case "mgmt_bmc_net_ip":
							for _, i := range strings.Split(sv, ",") {
								ip := net.ParseIP(i)
								if !ContainS(d.MgmtBmcNetIp, ip.String()) {
									d.MgmtBmcNetIp = append(d.MgmtBmcNetIp, ip.String())
								}
							}
						case "data1_net_macs":
							for _, m := range strings.Split(sv, ",") {
								mac, err := net.ParseMAC(strings.Trim(m, `"`))
								if err != nil {
									return discover, err
								}
								if !ContainS(d.Data1NetMacs, mac.String()) {
									d.Data1NetMacs = append(d.Data1NetMacs, mac.String())
								}
							}
						case "mgmt_net_bonding_master":
							d.MgmtNetBondingMaster = sv
						case "admin_house_interface":
							d.AdminHouseInterface = sv
						case "extra_routes":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return discover, err
							}
							d.ExtraRoutes = boolVal
						case "mgmt_bmc_net_if":
							// TODO:
						case "mgmt_bmc_net_if_ip":
							for _, i := range strings.Split(sv, ",") {
								ip := net.ParseIP(i)
								if !ContainS(d.MgmtBmcNetIfIp, ip.String()) {
									d.MgmtBmcNetIfIp = append(d.MgmtBmcNetIfIp, ip.String())
								}
							}
						case "cmm_parent":
							d.CmmParent = sv
						case "ice":
							d.Ice = sv
						case "net":
							d.Net = sv
						case "type":
							d.Type = sv
						case "hostname1":
							d.Hostname1 = sv
						case "mgmtsw_partner":
							d.MgmtSwPartner = sv
						case "mgmtsw_isls":
							d.MgmtswIsls = sv
						case "discover_skip_switchconfig":
							d.DiscoverSkipSwitchconfig = sv
						case "pdu_protocol":
							d.PduProtocol = sv
						default:
							log.Debug().Msgf("UNKNOWN discover key/val %+v: %v", sk, sv)
						}
					} else {
						d.Hostname1 = kvp[0]
					}
				}
				discover[d.Hostname1] = d
			}
		}
	}

	return discover, nil
}

// importCfgDiscoverSection parses the [discover] section of an hpcm file and translates it to a Discover object
func importCfgTemplatesSection(cfg *ini.File) (map[string]Template, error) {
	templates := map[string]Template{}
	for _, section := range cfg.Sections() {
		if section.Name() == "templates" {
			t := Template{}
			for _, v := range section.Key("name").ValueWithShadows() {
				subkeys := strings.Split(v, ", ")
				for _, subkey := range subkeys {
					kvp := strings.Split(subkey, "=")
					if len(kvp) == 2 {
						sk, sv := strings.TrimSpace(kvp[0]), kvp[1]
						switch sk {
						case "mgmt_bmc_net_name":
							t.MgmtBmcNetName = sv
						case "mgmt_net_name":
							t.MgmtNetName = sv
						case "redundant_mgmt_network":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.RedundantMgmtNetwork = boolVal
						case "switch_mgmt_network":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.SwitchMgmtNetwork = boolVal
						case "dhcp_bootfile":
							t.DhcpBootfile = sv
						case "force_disk":
							t.ForceDisk = sv
						case "conserver_logging":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.ConserverLogging = boolVal
						case "conserver_ondemand":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.ConserverOnDemand = boolVal
						case "rootfs":
							t.RootFs = sv
						case "console_device":
							t.ConsoleDevice = sv
						case "tpm_boot":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.TpmBoot = boolVal
						case "mgmt_net_bonding_master":
							t.MgmtNetBondingMaster = sv
						case "disk_bootloader":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.DiskBootloader = boolVal
						case "predictable_net_names":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.PredicatableNetNames = boolVal
						case "mgmtsw":
							t.MgmtSw = sv
						case "transport":
							t.Transport = sv
						case "bmc_username":
							t.BmcUsername = sv
						case "bmc_password":
							t.BmcPassword = sv
						case "baud_rate":
							n, err := strconv.Atoi(sv)
							if err != nil {
								n = 115200
								// return discover, err
							}
							t.BaudRate = n
						case "mgmt_net_interfaces":
							for _, i := range strings.Split(sv, ",") {
								iface := strings.Trim(i, `"`)
								if !ContainS(t.MgmtNetInterfaces, iface) {
									t.MgmtNetInterfaces = append(t.MgmtNetInterfaces, iface)
								}
							}
						case "architecture":
							t.Architecture = sv
						case "card_type":
							t.CardType = sv
						case "mgmt_net_bonding_mode":
							t.MgmtNetBodingMode = sv
						case "su_leader_role":
							t.SuLeaderRole = sv
						case "image":
							t.Image = sv
						case "ctrl_model":
							t.CtrlModel = sv
						case "data1_net_name":
							t.Data1NetName = sv
						case "data1_net_interfaces":
							for _, i := range strings.Split(sv, ",") {
								iface := strings.Trim(i, `"`)
								if !ContainS(t.Data1NetInterfaces, iface) {
									t.Data1NetInterfaces = append(t.Data1NetInterfaces, iface)
								}
							}
						case "data2_net_name":
							t.Data1NetName = sv
						case "data2_net_interfaces":
							for _, i := range strings.Split(sv, ",") {
								iface := strings.Trim(i, `"`)
								if !ContainS(t.Data2NetInterfaces, iface) {
									t.Data2NetInterfaces = append(t.Data2NetInterfaces, iface)
								}
							}
						case "password":
							t.Password = sv
						case "username":
							t.Username = sv
						case "nfs_writable_type":
							t.NfsWritableType = sv
						case "su_leader":
							for _, i := range strings.Split(sv, ",") {
								ip := net.ParseIP(i)
								if !ContainS(t.SuLeader, ip.String()) {
									t.SuLeader = append(t.SuLeader, ip.String())
								}
							}
						case "mgmtsw_isls":
							t.MgmtswIsls = sv
						case "destroy_disk_label":
							boolVal, err := ParseBool(sv)
							if err != nil {
								return templates, err
							}
							t.DestroyDiskLabel = boolVal
						case "md_metadata":
							t.MdMetadata = sv
						default:
							log.Debug().Msgf("UNKNOWN templates key/val %+v: %v", sk, sv)
						}
						templates[t.Name] = t
					} else {
						t.Name = kvp[0]
					}
				}
			}
		}
	}

	return templates, nil
}

// importCfgDiscoverSection parses the [discover] section of an hpcm file and translates it to a Discover object
func importCfgNicTemplatesSection(cfg *ini.File) (map[string]NicTemplate, error) {
	templates := map[string]NicTemplate{}
	for _, section := range cfg.Sections() {
		if section.Name() == "nic_templates" {
			n := NicTemplate{}
			for _, v := range section.Key("name").ValueWithShadows() {
				subkeys := strings.Split(v, ", ")
				for _, subkey := range subkeys {
					kvp := strings.Split(subkey, "=")
					if len(kvp) == 2 {
						sk, sv := strings.TrimSpace(kvp[0]), kvp[1]
						switch sk {
						case "template":
							n.Template = sv
						case "network":
							n.Network = sv
						case "bonding_master":
							n.BondingMaster = sv
						case "bonding_mode":
							n.BondingMode = sv
						case "net_ifs":
							for _, i := range strings.Split(sv, ",") {
								iface := strings.Trim(i, `"`)
								if !ContainS(n.NetIfs, iface) {
									n.NetIfs = append(n.NetIfs, iface)
								}
							}
						case "br_name":
							n.BrNane = sv
						default:
							log.Debug().Msgf("UNKNOWN nic_templates key/val %+v: %v", sk, sv)
						}
					}
				}
			}
		}
	}

	return templates, nil
}

// importCfgDiscoverSection parses the [discover] section of an hpcm file and translates it to a Discover object
func importCfgDnsSection(cfg *ini.File) (map[string]Dns, error) {
	dns := map[string]Dns{}
	for _, section := range cfg.Sections() {
		if section.Name() == "dns" {
			o := Dns{}
			for k, v := range section.KeysHash() {
				switch k {
				case "cluster_domain":
					o.ClusterDomain = v
				case "nameserver1":
					o.Nameserver1 = v
				case "nameserver2":
					o.Nameserver2 = v
				default:
					log.Debug().Msgf("UNKOWN dns key/val %+v %+v", k, v)
				}
				dns[o.ClusterDomain] = o
			}
		}
	}

	return dns, nil
}

// importCfgAttributesSection parses the [attributes] section of an hpcm file and translates it to an Attributes object
func importCfgAttributesSection(cfg *ini.File) (Attributes, error) {
	o := Attributes{}
	for _, section := range cfg.Sections() {
		if section.Name() == "attributes" {
			// o := Attributes{}
			for k, v := range section.KeysHash() {
				switch k {
				case "admin_house_interface":
					o.AdminHouseInterface = v
				case "admin_mgmt_interfaces":
					for _, i := range strings.Split(v, ",") {
						iface := strings.Trim(i, `"`)
						if !ContainS(o.AdminManagementInterfaces, iface) {
							o.AdminManagementInterfaces = append(o.AdminManagementInterfaces, iface)
						}
					}
				case "admin_mgmt_bmc_interfaces":
					for _, i := range strings.Split(v, ",") {
						iface := strings.Trim(i, `"`)
						if !ContainS(o.AdminManagementBmcInterfaces, iface) {
							o.AdminManagementBmcInterfaces = append(o.AdminManagementBmcInterfaces, iface)
						}
					}
				case "admin_udpcast_ttl":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.AdminUdpcastTtl = n
				case "admin_udpcast_mcast_rdv_addr":
					o.AdminUdpcastMcaseRdvAddr = v
				case "admin_mgmt_bonding_mode":
					o.AdminMgmtBondingMode = v
				case "blademond_scan_interval":
					o.BladeMondScanInterval = v
				case "cmcs_per_mgmt_vlan":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.CmcsPerMgmtVlan = n
				case "cmcs_per_rack":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.CmcsPerRack = n
				case "cmms_per_rack":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.CmmsPerRack = n
				case "conserver_logging":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.ConserverLogging = boolVal
				case "conserver_ondemand":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.ConserverOnDemand = boolVal
				case "copy_admin_ssh_config":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.CopyAdminSshConfig = boolVal
				case "dhcp_bootfile":
					o.DhcpBootfile = v
				case "discover_skip_switchconfig":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.DiscoverSkipSwitchconfig = boolVal
				case "domain_search_path":
					for _, sd := range strings.Split(v, ",") {
						if !ContainS(o.DomainSearchPath, sd) {
							o.DomainSearchPath = append(o.DomainSearchPath, sd)
						}
					}
				case "head_vlan":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.HeadVlan = n
				case "ipv6_local_site_ula":
					o.Ipv6LocalSiteUla = v
				case "max_rack_irus":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.MacRackIrus = n
				case "mcell_network":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.McellNetwork = boolVal
				case "mcell_vlan":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.McellVlan = n
				case "mgmt_ctrl_vlan_end":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.MgmtCtrlVlanEnd = n
				case "mgmt_ctrl_vlan_start":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.MgmtCtrlVlanStart = n
				case "mgmt_net_routing_protocol":
					o.MgmtNetRoutingProtocol = v
				case "mgmt_net_subnet_selection":
					o.MgmtNetSubetSelection = v
				case "mgmt_vlan_end":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.MgmtVlanEnd = n
				case "mgmt_vlan_start":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.MgmtVlanStart = n
				case "predictable_net_names":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.PredicatableNetNames = boolVal
				case "rack_start_number":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.RackStartNumber = n
				case "rack_vlan_end":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.RackVlanEnd = n
				case "rack_vlan_start":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.RackVlanStart = n
				case "redundant_mgmt_network":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.RedundantMgmtNetwork = boolVal
				case "switch_mgmt_network":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.SwitchMgmtNetwork = boolVal
				case "udpcast_max_bitrate":
					o.UdpcastMaxBitrate = v
				case "udpcast_max_wait":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.UdpcastMaxWait = n
				case "udpcast_mcast_rdv_addr":
					o.UdpcastMcastRdvAddr = v
				case "udpcast_min_receivers":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.UdpcastMinRecievers = n
				case "udpcast_min_wait":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.UdpcastMinWait = n
				case "udpcast_rexmit_hello_interval":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.UdpcastRexmitHelloInterval = n
				case "monitoring_kafka_elk_alerta_enabled":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.MonitoringKafkaElkAlertEnabled = boolVal
				case "monitoring_native_enabled":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.MonitoringNativeEnabled = boolVal
				case "conserver_timestamp":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.ConserverTimestamp = boolVal
				case "dhcpd_max_lease_time":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.DhcpdMaxLeaseTime = n
				case "dhcpd_default_lease_time":
					n, err := strconv.Atoi(v)
					if err != nil {
						return o, err
					}
					o.DhcpdDeafultLeaseTime = n
				case "my_sql_replication":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.MySqlReplication = boolVal
				case "monitoring_ganglia_enabled":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.MonitoringGangliaEnabled = boolVal
				case "monitoring_nagios_enabled":
					boolVal, err := ParseBool(v)
					if err != nil {
						return o, err
					}
					o.MonitoringNagiosEnabled = boolVal
				case "mgmt_net_alias_selection":
					o.MgmtNetAliasSelection = v
				default:
					log.Debug().Msgf("UNKOWN %+v %+v", k, v)
				}
			}
		}
	}

	return o, nil
}

// importCfgDiscoverSection parses the [discover] section of an hpcm file and translates it to a Discover object
func importCfgNetworksSection(cfg *ini.File) (map[string]Network, error) {
	networks := map[string]Network{}
	for _, section := range cfg.Sections() {
		if section.Name() == "networks" {
			o := Network{}
			for _, v := range section.Key("name").ValueWithShadows() {
				subkeys := strings.Split(v, ", ")
				for _, subkey := range subkeys {
					kvp := strings.Split(subkey, "=")
					if len(kvp) == 2 {
						sk, sv := strings.TrimSpace(kvp[0]), kvp[1]
						switch sk {
						case "name":
							o.Name = sv
						case "type":
							o.Type = sv
						case "subnet":
							o.Subnet = sv
						case "netmask":
							o.Netmask = sv
						case "rack_netmask":
							o.RackNetmask = sv
						case "gateway":
							o.Gateway = sv

						default:
							log.Debug().Msgf("UNKNOWN networks key/val %+v: %v", sk, sv)
						}
					} else {
						o.Name = kvp[0]
					}
					networks[o.Name] = o
				}
			}
		}
	}

	return networks, nil
}

// importCfgAttributesSection parses the [attributes] section of an hpcm file and translates it to an Attributes object
func importCfgImagesSection(cfg *ini.File) ([]Images, error) {
	images := []Images{}
	for _, section := range cfg.Sections() {
		if section.Name() == "images" {
			o := Images{}
			for k, v := range section.KeysHash() {
				switch k {
				case "image_types":
					for _, i := range strings.Split(v, ",") {
						imagek := strings.Trim(i, `"`)
						if !ContainS(o.ImageTypes, imagek) {
							o.ImageTypes = append(o.ImageTypes, imagek)
						}
					}
					images = append(images, o)
				default:
					log.Debug().Msgf("UNKOWN %+v %+v", k, v)
				}
			}
		}
	}

	return images, nil
}

var kvPairRe = regexp.MustCompile(`(.*?)=([^=]*)(?:,|$)`)

// ParseKV parses a key/val
func ParseKV(kvStr string) map[string]string {
	res := map[string]string{}
	for _, kv := range kvPairRe.FindAllStringSubmatch(kvStr, -1) {
		res[kv[1]] = kv[2]
	}
	return res
}

// ParseBool returns a bool based off of several different possible strings
func ParseBool(str string) (bool, error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "Yes", "yes", "y":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "No", "no", "n":
		return false, nil
	}
	return false, nil
}

// ContainS checks if a slice already contains a string
func ContainS(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

// ContainI checks if a slice already contains an int
func ContainI(ints []int, x int) bool {
	for _, i := range ints {
		if x == i {
			return true
		}
	}
	return false
}
