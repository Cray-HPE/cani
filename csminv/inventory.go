/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package csminv

import (
	"net"
	"time"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

// Extract is any data coming in from an external system
type Extract struct {
	// The entire paddle file
	CanuConfig CanuConfig `yaml:"canu"`
	// The entire csi system_config.yaml file
	CsiConfig CsiConfig `yaml:"csi"`
	// The dumpstate from SLS
	SlsConfig SlsConfig `yaml:"sls"`
	// any other arbitrary data
	// ....
	// FromRedfish RedfishConfig `yaml:"redfish"`
	// FromHMS HMSConfig `yaml:"hms"`
}

// Inventory is greenfield data structures for the new inventory
// An Extract is transformed into this portable inventory data structure, which is loaded somewhere
type Inventory struct {
	// This is everything from the legacy "inventory"
	Extract Extract
	// Slice of all cabinets in the system
	Cabinets []Hardware
	// Slice of all switches in the system
	Switches []Hardware
	// Slice of all nodes in the system
	Nodes []Hardware
	// Slice of all PDUs in the system
	PDUs []Hardware
}

type Hardware struct {
	// SLS:   "Hardware.ExtraProperties.Aliases" ===> "Hardware.Names"
	// CANU:  "Topology.CommonName" ===> "Hardware.Names"
	Names []string
	// SLS:   "Hardware.Type" ===> "Hardware.Type"
	// SLS:   "Hardware.TypeString" ===> "Hardware.Type"
	// CANU:  "Topology.Type" ===> "Hardware.Type"
	Type string
	// SLS:   "Hardware.TypeString" ===> "Hardware.Type"
	// SLS:   "Hardware.Class" ===> "Hardware.Class"
	Class string
	// CANU:  "Topology.Model" ===> "Hardware.Model"
	Model string
	// CANU:  "Architechture" ===> "Hardware.Architechture"
	Architechture string
	// SLS:   "Hardware.ExtraProperties.Brand" ===> "Hardware.Manufacturer"
	// CANU:  "Topology.Vendor" ===> "Hardware.Manufacturer"
	Manufacturer string
	// SLS:   "Hardware.ExtraProperties.Brand" ===> "Hardware.Vendor"
	// CANU:  "Topology.Vendor" ===> "Hardware.Vendor"
	Vendor string
	// CSI:   "csm_version" ===> "Hardware.CsmVersion"
	CsmVersion string
	Networking Networking
	// SLS:   "Hardware.ExtraProperties.IP4Addr" ===> "Hardware.IP"
	IP net.IP
	// TODO: Generate or pull from Redfish
	GUID string
	// New keys we want to add
	// ...
}

type Networking struct {
	// The name of the interface
	SiteIP     net.IP
	CanGW      net.IP
	SiteDNS    []net.IP
	SiteDomain string
}

type SlsConfig sls_common.SLSState

// CsiConfig is the configuration that comes from CSI's system_config.yaml
type CsiConfig struct {
	// The path to the application_node_config.yaml file
	ApplicationNodeConfigYaml string `json:"application_node_config_yaml" yaml:"application-node-config-yaml" env:"CSI_APPPLICATION_NODE_CONFIG" default:"application-node-config.yaml" flag:"application-node-config" usage:"Path the application node config yaml file" jsonschema:"required"`
	// Border Gate Protocol Autonomous System Number
	BgpAsn string `json:"bgp_asn" yaml:"bgp-asn" default:"65533" env:"CSI_BGP_ASN" usage:"ASN for BGP" jsonschema:"required"`
	// Border Gate Protocol Autonomous System Number for CHN
	BgpChnAsn string `json:"ggp_chn_asn" yaml:"bgp-chn-asn" default:"65530" env:"CSI_BGP_CHN_ASN" usage:"ASN for BGP on the CHN" jsonschema:"required"`
	// Border Gate Protocol Autonomous System Number for CMN
	BgpCmnAsn string `json:"bgp_cmn_asn" yaml:"bgp-cmn-asn" default:"65532" env:"CSI_BGP_CMN_ASN" usage:"ASN for BGP on the CMN" jsonschema:"required"`
	// Border Gate Protocol Autonomous System Number for NMN
	BgpNmnAsn string `json:"bgp_nmn_asn" yaml:"bgp-nmn-asn" default:"65531" env:"CSI_BGP_NMN_ASN" usage:"ASN for BGP on the NMN" jsonschema:"required"`
	// BGP peer types
	BgpPeerTypes []string `json:"bgp_peer_types" yaml:"bgp-peer-types" default:"spine," env:"CSI_BPG_PEER_TYPE" usage:"Types of BGP peers" jsonschema:"required"`
	// BGP peers
	BgpPeers string `json:"bgp-peers" yaml:"bgp-peers" default:"spine" env:"CSI_BGP_PEERS" usage:"BGP peers" jsonschema:"required"`
	// Network name for the bifurcated-customer-access-network
	BicanUserNetworkName string `json:"bican_user_network_name" yaml:"bican-user-network-name" default:"CAN" env:"CSI_BICAN_USER_NETWORK_NAME" usage:"Name of the user bifurcated customer access network" jsonschema:"required"`
	// The BMC password used during bootstrapping
	BootstrapNcnBmcPass string `json:"bootstrap_ncn_bmc_pass" yaml:"bootstrap-ncn-bmc-pass" default:"" env:"CSI_NCN_BMC_PASS" usage:"Password for the BMC of the bootstrap node" jsonschema:"required"`
	// The BMC username used during bootstrapping
	BootstrapNcnBmcUser string `json:"bootstrap_ncn_bmc_user" yaml:"bootstrap-ncn-bmc-user" default:"" env:"CSI_BOOTSTRAP_NCN_BMC_USER" usage:"Username for the BMC of the bootstrap node" jsonschema:"required"`
	// The path to the cabinets.yaml file
	CabinetsYaml string `json:"cabinets_yaml" yaml:"cabinets-yaml" default:"" env:"CSI_CABINETS_YAML" usage:"Path to the cabinets.yaml file" jsonschema:"required"`
	// The CAN bootstrap VLAN ID
	CanBootstrapVlan int `json:"can_bootstrap_vlan" yaml:"can-bootstrap-vlan" default:"6" env:"CSI_CAN_BOOTSTRAP_VLAN" usage:"VLAN ID for the bootstrap CAN" jsonschema:"required"`
	// The CAN CIDR
	CanCidr string `json:"can_cidr" yaml:"can-cidr" default:"" env:"CSI_CAN_CIDR" usage:"CIDR for the CAN" jsonschema:"required"`
	// The CAN dynamic IP pool
	CanDynamicPool string `json:"can_dynamic_pool" yaml:"can-dynamic-pool" default:"" env:"CSI_CAN_DYNAMIC_POOL" usage:"Dynamic IP pool for the CAN" jsonschema:"required"`
	// The CAN gateway
	CanGateway string `json:"can_gateway" yaml:"can-gateway" default:"" env:"CSI_CAN_GATEWAY" usage:"Gateway for the CAN" jsonschema:"required"`
	// The CAN static IP pool
	CanStaticPool string `json:"can_static_pool" yaml:"can-static-pool" default:"" env:"CSI_CAN_STATIC_POOL" usage:"Static IP pool for the CAN" jsonschema:"required"`
	// The default ceph cehpfs image
	CephCephfsImage string `json:"ceph_cephfs_image" yaml:"ceph-cephfs-image" default:"dtr.dev.cray.com/cray/cray-cephfs-provisioner:0.1.0-nautilus-1.3" env:"CSI_CEPH_CEPHFS_IMAGE" usage:"Ceph CephFS image" jsonschema:"required"`
	// The default ceph rbd image
	CephRbdImage string `json:"ceph_rbd_image" yaml:"ceph-rbd-image" default:"dtr.dev.cray.com/cray/cray-rbd-provisioner:0.1.0-nautilus-1.3" env:"CSI_CEPH_RBD_IMAGE" usage:"Ceph RBD image" jsonschema:"required"`
	// The CHN CIDR
	ChnCidr string `json:"chn_cidr" yaml:"chn-cidr" default:"" env:"CSI_CHN_CIDR" usage:"CIDR for the CHN" jsonschema:"required"`
	// The CHN dynamic IP pool
	ChnDynamicPool string `json:"chn_dynamic_pool" yaml:"chn-dynamic-pool" default:"" env:"CSI_CHN_DYNAMIC_POOL" usage:"Dynamic IP pool for the CHN" jsonschema:"required"`
	// The CHN gateway
	ChnGateway string `json:"chn_gateway" yaml:"chn-gateway" default:"" env:"CSI_CHN_GATEWAY" usage:"Gateway for the CHN" jsonschema:"required"`
	// The CHN static IP pool
	ChnStaticPool string `json:"chn_static_pool" yaml:"chn-static-pool" default:"" env:"CSI_CHN_STATIC_POOL" usage:"Static IP pool for the CHN" jsonschema:"required"`
	// The CMN bootstrap VLAN ID
	CmnBootstrapVlan int `json:"cmn_bootstrap_vlan" yaml:"cmn-bootstrap-vlan" default:"0" env:"CSI_CMN_BOOTSTRAP_VLAN" usage:"VLAN ID for the bootstrap CMN" jsonschema:"required"`
	// The CMN CIDR
	CmnCidr string `json:"cmn_cidr" yaml:"cmn-cidr" default:"" env:"CSI_CMN_CIDR" usage:"CIDR for the CMN" jsonschema:"required"`
	// The CMN dynamic IP pool
	CmnDynamicPool string `json:"cmn_dynamic_pool" yaml:"cmn-dynamic-pool" default:"" env:"CSI_CMN_DYNAMIC_POOL" usage:"Dynamic IP pool for the CMN" jsonschema:"required"`
	// The CMN external DNS
	CmnExternalDNS string `json:"cmn_external_dns" yaml:"cmn-external-dns" default:"" env:"CSI_CMN_EXTERNAL_DNS" usage:"External DNS for the CMN" jsonschema:"required"`
	// The CMN gateway
	CmnGateway string `json:"cmn_gateway" yaml:"cmn-gateway" default:"" env:"CSI_CMN_GATEWAY" usage:"Gateway for the CMN" jsonschema:"required"`
	// The CMN static IP pool
	CmnStaticPool string `json:"cmn_static_pool" yaml:"cmn-static-pool" default:"" env:"CSI_CMN_STATIC_POOL" usage:"Static IP pool for the CMN" jsonschema:"required"`
	// The path to the CSM config file
	Config string `json:"config" yaml:"config" default:"" env:"CSI_CONFIG" usage:"Path to the CSI config file" jsonschema:"required"`
	// The CSM version this config file is used for
	CsmVersion string `json:"csm_version" yaml:"csm-version" default:"1.4" env:"CSI_CSM_VERSION" usage:"Version of CSM to install" jsonschema:"required"`
	// The name of the docker image registry
	DockerImageRegistry string `json:"docker_image_registry" yaml:"docker-image-registry" default:"" env:"CSI_DOCKER_IMAGE_REGISTRY" usage:"Docker image registry" jsonschema:"required"`
	// The first master hostname--usually ncn-m002--since ncn-m001 is used for the initial bootstrapping
	FirstMasterHostname string `json:"first_master_hostname" yaml:"first-master-hostname" default:"ncn-m002" env:"CSI_FIRST_MASTER_HOSTNAME" usage:"Hostname of the first master node" jsonschema:"required"`
	// The number of hill cabinets
	HillCabinets int `json:"hill_cabinets" yaml:"hill-cabinets" default:"0" env:"CSI_hill_cabinets" usage:"Number of hill cabinets" jsonschema:"required"`
	// The number of hill nodes
	HmnBootstrapVlan int `json:"hmn_bootstrap_vlan" yaml:"hmn-bootstrap-vlan" default:"4" env:"CSI_HMN_BOOTSTRAP_VLAN" usage:"VLAN ID for the bootstrap HMN" jsonschema:"required"`
	// The HMN CIDR
	HmnCidr string `json:"hmn_cidr" yaml:"hmn-cidr" default:"10.254.0.0/17" env:"CSI_HMN_CIDR" usage:"CIDR for the HMN" jsonschema:"required"`
	// The path to the hmn_connections.json file
	HmnConnections string `json:"hmn_connections" yaml:"hmn-connections" default:"hmn_connections.json" env:"CSI_HMN_CONNECTIONS" usage:"Path to the HMN connections file" jsonschema:"required"`
	// The HMN dynamic IP pool
	HmnDynamicPool string `json:"hmn_dynamic_pool" yaml:"hmn-dynamic-pool" default:"10.94.100.0/24" env:"CSI_HMN_DYNAMIC_POOL" usage:"Dynamic IP pool for the HMN" jsonschema:"required"`
	// The HMN CIDR
	HmnMtnCidr string `json:"hmn_mtn_cidr" yaml:"hmn-mtn-cidr" default:"10.104.0.0/17" env:"CSI_HMN_MTN_CIDR" usage:"CIDR for the HMN mountain" jsonschema:"required"`
	// The HMN river CIDR
	HmnRvrCidr string `json:"hmn_rvr_cidr" yaml:"hmn-rvr-cidr" default:"10.107.0.0/17" env:"CSI_HMN_RVR_CIDR" usage:"CIDR for the HMN river" jsonschema:"required"`
	// The HMN static IP pool
	HmnStaticPool string `json:"hmn_static_pool" yaml:"hmn-static-pool" default:"" env:"CSI_HMN_STATIC_POOL" usage:"Static IP pool for the HMN" jsonschema:"required"`
	// The HSN CIDR
	HsnCidr string `json:"hsn_cidr" yaml:"hsn-cidr" default:"10.253.0.0/16" env:"CSI_HSN_CIDR" usage:"CIDR for the HSN" jsonschema:"required"`
	// The HSN dynamic IP pool
	HsnDynamicPool string `json:"hsn_dynamic_pool" yaml:"hsn-dynamic-pool" default:"" env:"CSI_HSN_DYNAMIC_POOL" usage:"Dynamic IP pool for the HSN" jsonschema:"required"`
	// The HSN static IP pool
	HsnStaticPool string `json:"hsn_static_pool" yaml:"hsn-static-pool" default:"" env:"CSI_HSN_STATIC_POOL" usage:"Static IP pool for the HSN" jsonschema:"required"`
	// The ncn used for the initial bootstrapping
	InstallNcn string `json:"install_ncn" yaml:"install-ncn" default:"ncn-m001" env:"CSI_INSTALL_NCN" usage:"Hostname of the node to install on" jsonschema:"required"`
	// The interface name(s) for the bond on the install node
	InstallNcnBondMembers string `json:"install_ncn_bond_members" yaml:"install-ncn-bond-members" default:"p1p1,p1p2" env:"CSI_INSTALL_NCN_BOND_MEMBERS" usage:"Interface names for bond members on the install node" jsonschema:"required"`
	// The
	Ipv4Resolvers string `json:"ipv4_resolvers" yaml:"ipv4-resolvers" default:"8.8.8.8,9.9.9.9" env:"CSI_IPV4_RESOLVERS" usage:"IPv4 resolvers" jsonschema:"required"`
	// Enable or disabling the Kubernetes API auditing
	K8SAPIAuditingEnabled bool `json:"k8s_api_auditing_enabled" yaml:"k8s-api-auditing-enabled" default:"false" env:"CSI_K8S_API_AUDITING_ENABLED" usage:"Enable Kubernetes API auditing" jsonschema:"required"`
	// Number of IPs to allocate for the management network
	ManagementNetIps int `json:"management_net_ips" yaml:"management-net-ips" default:"0" env:"CSI_MANAGEMENT_NET_IPS" usage:"Number of IPs to allocate for the management network" jsonschema:"required"`
	// Release of the manifests to install
	ManifestRelease string `json:"manifest_release" yaml:"manifest-release" default:"" env:"CSI_MANIFEST_RELEASE" usage:"Release of the manifests to install" jsonschema:"required"`
	// The number of mountain cabinets
	MountainCabinets int `json:"mountain_cabinets" yaml:"mountain-cabinets" default:"4" env:"CSI_MOUNTAIN_CABINETS" usage:"Number of mountain cabinets" jsonschema:"required"`
	// The MTL CIDR
	MtlCidr string `json:"mtl_cidr" yaml:"mtl-cidr" default:"10.1.1.0/16" env:"CSI_MTL_CIDR" usage:"CIDR for the MTL" jsonschema:"required"`
	// Path to the ncn_metadata.csv file
	NcnMetadata string `json:"ncn_metadata" yaml:"ncn-metadata" default:"ncn_metadata.csv" env:"CSI_NCN_METADATA" usage:"Path to the NCN metadata file" jsonschema:"required"`
	// Enable auditing on the NCN management node
	NcnMgmtNodeAuditingEnabled bool `json:"ncn_mgmt_node_auditing_enabled" yaml:"ncn-mgmt-node-auditing-enabled" default:"false" env:"CSI_NCN_MGMT_NODE_AUDITING_ENABLED" usage:"Enable auditing on the NCN management node" jsonschema:"required"`
	// The NMN bootstrap VLAN ID
	NmnBootstrapVlan int `json:"nmn_bootstrap_vlan" yaml:"nmn-bootstrap-vlan" default:"2" env:"CSI_NMN_BOOTSTRAP_VLAN" usage:"VLAN ID for the bootstrap NMN" jsonschema:"required"`
	// The NMN CIDR
	NmnCidr string `json:"nmn_cidr" yaml:"nmn-cidr" default:"10.252.0.0/17" env:"CSI_NMN_CIDR" usage:"CIDR for the NMN" jsonschema:"required"`
	// The NMN dynamic IP pool
	NmnDynamicPool string `json:"nmn_dynamic_pool" yaml:"nmn-dynamic-pool" default:"10.92.100.0/24" env:"CSI_NMN_DYNAMIC_POOL" usage:"Dynamic IP pool for the NMN" jsonschema:"required"`
	// The NMN mountain CIDR
	NmnMtnCidr string `json:"nmn_mtn_cidr" yaml:"nmn-mtn-cidr" default:"10.100.0.0/17" env:"CSI_NMN_MTN_CIDR" usage:"CIDR for the NMN mountain" jsonschema:"required"`
	// The NMN river CIDR
	NmnRvrCidr string `json:"nmn_rvr_cidr" yaml:"nmn-rvr-cidr" default:"10.106.0.0/17" env:"CSI_NMN_RVR_CIDR" usage:"CIDR for the NMN river" jsonschema:"required"`
	// The NMN static IP pool
	NmnStaticPool string `json:"nmn_static_pool" yaml:"nmn-static-pool" default:"" env:"CSI_NMN_STATIC_POOL" usage:"Static IP pool for the NMN" jsonschema:"required"`
	// DNS notify zones
	NotifyZones string `json:"notify_zones" yaml:"notify-zones" default:"" env:"CSI_NOTIFY_ZONES" usage:"DNS notify zones" jsonschema:"required"`
	// A comma separate list of hostname/IPs of the NCNs that will participate as an NTP peer
	NtpPeers []string `json:"ntp_peers" yaml:"ntp-peers" default:"ncn-m001,ncn-m002,ncn-m003,ncn-w001,ncn-w002,ncn-w003,ncn-s001,ncn-s002,ncn-s003" env:"CSI_NTP_PEERS" usage:"List of NTP peers" jsonschema:"required"`
	// A comma separated list of upstream NTP pools
	NtpPools []interface{} `json:"ntp_pools" yaml:"ntp-pools" default:"" env:"CSI_NTP_POOLS" usage:"List of NTP pools" jsonschema:"required"`
	// A comma separated list of upstream NTP servers
	NtpServers []string `json:"ntp_servers" yaml:"ntp-servers" default:"ncn-m001," env:"CSI_NTP_SERVERS" usage:"List of NTP servers" jsonschema:"required"`
	// Timezone for NTP
	NtpTimezone string `json:"ntp_timezone" yaml:"ntp-timezone" default:"UTC" env:"CSI_NTP_TIMEZONE" usage:"Timezone for NTP" jsonschema:"required"`
	// Name of the primary server
	PrimaryServerName string `json:"primary_server_name" yaml:"primary-server-name" default:"primary" env:"CSI_PRIMARY_SERVER_NAME" usage:"Name of the primary server" jsonschema:"required"`
	// Retain unused user network
	RetainUnusedUserNetwork bool `json:"retain_unused_user_network" yaml:"retain-unused-user-network" default:"false" env:"CSI_RETAIN_UNUSED_USER_NETWORK" usage:"Retain unused user network" jsonschema:"required"`
	// The number of river cabinets
	RiverCabinets int `json:"river_cabinets" yaml:"river-cabinets" default:"1" env:"CSI_RIVER_CABINETS" usage:"Number of river cabinets" jsonschema:"required"`
	// The RPM repository to use
	RpmRepository string `json:"rpm_repository" yaml:"rpm-repository" default:"" env:"CSI_RPM_REPOSITORY" usage:"RPM repository to use" jsonschema:"required"`
	// List of secondary servers
	SecondaryServers string `json:"secondary_servers" yaml:"secondary-servers" default:"" env:"CSI_SECONDARY_SERVERS" usage:"List of secondary servers" jsonschema:"required"`
	// The site DNS server
	SiteDNS string `json:"site_dns" yaml:"site-dns" default:"" env:"CSI_SITE_DNS" usage:"Site DNS servers" jsonschema:"required"`
	// The site domain
	SiteDomain string `json:"site_domain" yaml:"site-domain" default:"" env:"CSI_SITE_DOMAIN" usage:"Site domain" jsonschema:"required"`
	// The site gateway
	SiteGw string `json:"site_gw" yaml:"site-gw" default:"" env:"CSI_SITE_GW" usage:"Site gateway" jsonschema:"required"`
	// The site IP address
	SiteIP string `json:"site_ip" yaml:"site-ip" default:"" env:"CSI_SITE_IP" usage:"Site IP address" jsonschema:"required"`
	// The interface for the site IP address
	SiteNic string `json:"site_nic" yaml:"site-nic" default:"em1" env:"CSI_SITE_NIC" usage:"Site NIC" jsonschema:"required"`
	// The starting hill cabinet number
	StartingHillCabinet int `json:"starting_hill_cabinet" yaml:"starting-hill-cabinet" default:"9000" env:"CSI_STARTING_HILL_CABINET" usage:"Starting hill cabinet number" jsonschema:"required"`
	// The starting mountain cabinet number
	StartingMountainCabinet int `json:"starting_mountain_cabinet" yaml:"starting-mountain-cabinet" default:"1000" env:"CSI_STARTING_MOUNTAIN_CABINET" usage:"Starting mountain cabinet number" jsonschema:"required"`
	// The starting mountain NID
	StartingMountainNid int `json:"starting_mountain_nid" yaml:"starting-mountain-nid" default:"1000" env:"CSI_STARTING_MOUNTAIN_NID" usage:"Starting mountain NID" jsonschema:"required"`
	// The starting river cabinet number
	StartingRiverCabinet int `json:"starting_river_cabinet" yaml:"starting-river-cabinet" default:"3000" env:"CSI_STARTING_RIVER_CABINET" usage:"Starting river cabinet number" jsonschema:"required"`
	// The starting river NID
	StartingRiverNid int `json:"starting_river_nid" yaml:"starting-river-nid" default:"1" env:"CSI_STARTING_RIVER_NID" usage:"Starting river NID" jsonschema:"required"`
	// Use supernet
	Supernet bool `json:"supernet" yaml:"supernet" default:"true" env:"CSI_SUPERNET" usage:"Use supernet" jsonschema:"required"`
	// The path to the switch_metadata.csv
	SwitchMetadata string `json:"switch_metadata" yaml:"switch-metadata" default:"switch_metadata.csv" env:"CSI_SWITCH_METADATA" usage:"Path to the switch metadata file" jsonschema:"required"`
	// The name of the system
	SystemName string `json:"system_name" yaml:"system-name" default:"sn-2024" env:"CSI_SYSTEM_NAME" usage:"Name of the system" jsonschema:"required"`
	// V2 registy path
	V2Registry string `json:"v2_registry" yaml:"v2-registry" default:"https://registry.nmn/" env:"CSI_V2_REGISTRY" usage:"V2 registry to use" jsonschema:"required"`
	// Version info (populated by CSI)
	Versioninfo struct {
		Version   string    `json:"version" yaml:"version" default:"" env:"CSI_VERSION" usage:"Version of the software" jsonschema:"required"`
		Gitcommit string    `json:"gitcommit" yaml:"gitcommit" default:"" env:"CSI_GITCOMMIT" usage:"Git commit of the software" jsonschema:"required"`
		Builddate time.Time `json:"builddate" yaml:"builddate" default:"" env:"CSI_BUILDDATE" usage:"Build date of the software" jsonschema:"required"`
		Goversion string    `json:"goversion" yaml:"goversion" default:"" env:"CSI_GOVERSION" usage:"Go version of the software" jsonschema:"required"`
		Compiler  string    `json:"compiler" yaml:"compiler" default:"" env:"CSI_COMPILER" usage:"Compiler of the software" jsonschema:"required"`
		Platform  string    `json:"platform" yaml:"platform" default:"" env:"CSI_PLATFORM" usage:"Platform of the software" jsonschema:"required"`
	} `json:"versioninfo" yaml:"versioninfo" env:"CSI_VERSIONINFO" usage:"CSI version information" jsonschema:"required"`
}

// CanuConfig is the structure of the paddle file
type CanuConfig struct {
	// Version of canu used to generate the paddle file
	CanuVersion string `json:"canu_version" env:"CANU_CANU_VERSION" default:"" flag:"canu-version" usage:"Version of canu" jsonschema:"required"`
	// Architecture of the system
	Architecture string `json:"architecture" env:"CANU_ARCHITECTURE" default:"" flag:"architecture" usage:"Architechture of the system" jsonschema:"required"`
	// Path to the SHCD file used to generate the paddle file
	ShcdFile string `json:"shcd_file" env:"CANU_SHCD_FILE" default:"" flag:"shcd-file" usage:"Path to the SHCD file" jsonschema:"required"`
	// Timestamp of when the paddle file was generated
	UpdatedAt string `json:"updated_at" env:"CANU_UPDATED_AT" default:"" flag:"updated-at" usage:"Last update timestamp" jsonschema:"required"`
	// Topology of the system
	Topology []struct {
		// Common name of the node
		CommonName string `json:"common_name" env:"CANU_TOPOLOGY_COMMON_NAME" default:"" flag:"topology-common-name" usage:"Common name of the node" jsonschema:"required"`
		// A unique identifier for the node
		ID int `json:"id" env:"CANU_TOPOLOGY_ID" default:"" flag:"topology-id" usage:"Unique ID of the node" jsonschema:"required"`
		// The architecture of the node
		Architecture string `json:"architecture" env:"CANU_TOPOLOGY_architecture" default:"" flag:"topology-architecture" usage:"Architecture of the node" jsonschema:"required"`
		// The hardware model of the node
		Model string `json:"model" env:"CANU_TOPOLOGY_model" default:"" flag:"topology-model" usage:"Model of the node" jsonschema:"required"`
		// The hardwaren type of the node
		Type string `json:"type" env:"CANU_TOPOLOGY_TYPE" default:"" flag:"topology-type" usage:"Type of the node" jsonschema:"required"`
		// The hardware vendor of the node
		Vendor string `json:"vendor" env:"CANU_TOPOLOGY_VENDOR" default:"" flag:"topology-vendor" usage:"Hardware vendor of the node" jsonschema:"required"`
		// A list of the ports on the node
		Ports []struct {
			// The port source
			Port int `json:"port" env:"CANU_PORTS_PORT" default:"" flag:"ports-port" usage:"The source port number" jsonschema:"required"`
			// The port speed
			Speed int `json:"speed" env:"CANU_PORTS_SPEED" default:"" flag:"ports-speed" usage:"The source port speed" jsonschema:"required"`
			// The port slot
			Slot interface{} `json:"slot" env:"CANU_PORTS_SLOT" default:"" flag:"ports-slot" usage:"The source port slot" jsonschema:"required"`
			// The destination node ID
			DestinationNodeID int `json:"destination_node_id" env:"CANU_PORTS_DESTINATION_NODE_ID" default:"" flag:"ports-destination-node-id" usage:"The destination ID of the port being connected to" jsonschema:"required"`
			// The destination port
			DestinationPort int `json:"destination_port" env:"CANU_PORTS_DESTINATION_PORT" default:"" flag:"ports-destination-port" usage:"The destination port number of the port being connected to" jsonschema:"required"`
			// The destination slot
			DestinationSlot interface{} `json:"destination_slot" env:"CANU_PORTS_DESTINATION_SLOT" default:"" flag:"ports-destination-slot" usage:"The destination slot of the port being connected to" jsonschema:"required"`
		} `json:"ports" env:"CANU_PORTS" default:"" flag:"ports" usage:"A list of ports on the node" jsonschema:"required"`
		// The location of the node as used in the xname
		Location struct {
			// The rack the node is in
			Rack string `json:"rack" env:"CANU_LOCATION_RACK" default:"" flag:"location-rack" usage:"The rack the node is in" jsonschema:"required"`
			// The u the node is in
			Elevation string `json:"elevation" env:"CANU_LOCATION_ELEVATION" default:"" flag:"location-elevation" usage:"The elevation of the node (Rack Unit)" jsonschema:"required"`
		} `json:"location" env:"CANU_LOCATION" default:"" flag:"location" usage:"Location of the node" jsonschema:"required"`
	} `json:"topology" env:"CANU_TOPOLOGY" default:"" flag:"topology" usage:"Topology of the system" jsonschema:"required"`
}
